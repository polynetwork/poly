package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/smartcontract/event"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/eth/locker"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) Vote(service *native.NativeService) (bool, *inf.MakeTxParam, error) {
	return true, nil, nil
}

func (this *ETHHandler) MakeDepositProposal(service *native.NativeService) (*inf.MakeTxParam, error) {
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return nil, fmt.Errorf("Verify, contract params deserialize error: %v", err)
	}

	blockData, err := GetEthBlockByNumber(params.Height)
	if err != nil {
		return nil, fmt.Errorf("Verify, GetEthBlockByNumber error:%v", err)
	}

	proofData, err := hex.DecodeString(params.Proof)
	if err != nil {
		return nil, err
	}

	ethProof := new(ETHProof)
	err = json.Unmarshal(proofData, ethProof)
	if err != nil {
		return nil, err
	}

	if len(ethProof.StorageProofs) != 1 {
		return nil, fmt.Errorf("[Verify] incorrect proof format")
	}

	bf := bytes.NewBuffer(utils.CrossChainManagerContractAddress[:])
	keyBytes := ethComm.Hex2Bytes(inf.KEY_PREFIX_ETH + replace0x(ethProof.StorageProofs[0].Key))
	bf.Write(keyBytes)
	key := bf.Bytes()
	val, err := service.CacheDB.Get(key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return nil, fmt.Errorf("[Verify] key:%s already solved ", ethProof.StorageProofs[0].Key)
	}
	//todo 1. verify the proof with header
	//determine where the k and v from
	proofResult, err := verifyMerkleProof(ethProof, blockData)
	if err != nil {
		return nil, fmt.Errorf("Verify, verifyMerkleProof error:%v", err)
	}
	if proofResult == nil {
		return nil, fmt.Errorf("Verify, verifyMerkleProof failed!")
	}

	if !checkProofResult(proofResult, params.Value) {
		fmt.Printf("verify value hash failed\n")
		return nil, fmt.Errorf("Verify, verify value hash failed!")
	}

	proof := &Proof{}
	if err := proof.Deserialize(params.Value); err != nil {
		return nil, fmt.Errorf("Verify, eth proof deserialize error: %v", err)
	}

	//todo does the proof data too big??
	tmp := strings.Split(params.Value, "#")
	fromContractAddr := tmp[0]

	service.CacheDB.Put(key, proofData)

	ret := &inf.MakeTxParam{}
	ret.ToChainID = proof.ToChainID
	ret.FromContractAddress = fromContractAddr
	ret.FromChainID = params.SourceChainID
	ret.ToAddress = proof.ToAddress
	ret.Amount = proof.Amount
	//todo deal with the proof.decimal
	return ret, nil
}

func (this *ETHHandler) MakeTransaction(service *native.NativeService, param *inf.MakeTxParam) error {
	//todo add logic
	//1 construct tx
	log.Infof("===MakeTransaction param:is %v\n", param)
	log.Infof("FromContractAddress:%s\n", param.FromContractAddress)
	log.Infof("ToAddress:%s\n", param.ToAddress)
	log.Infof("Amount:%d\n", param.Amount)
	log.Infof("ToChainID:%d\n", param.ToChainID)
	log.Infof("FromChainID:%d\n", param.FromChainID)
	contractabi, err := abi.JSON(strings.NewReader(locker.EthereumCrossChainABI))
	if err != nil {
		return err
	}

	bindaddr := ethComm.HexToAddress(param.ToAddress)
	log.Infof("bindaddr:%v\n", bindaddr)
	amount := param.Amount
	//lockAddress := ethComm.HexToAddress(LOCKER_CONTRACT_ADDR)

	targetAsset, err := side_chain_manager.GetDestAsset(service, param.FromChainID, param.ToChainID, param.FromContractAddress)
	if err != nil {
		return err
	}
	log.Infof("targetTokenAddr:%s\n", targetAsset.ContractAddress)
	tokenAddress := ethComm.HexToAddress(targetAsset.ContractAddress)
	log.Infof("tokenAddress:%s\n", tokenAddress)

	txid := "1"
	v := []uint8{0}
	r := [][32]byte{[32]byte{0}}
	s := [][32]byte{[32]byte{0}}
	txData, err := contractabi.Pack("Withdraw", tokenAddress, txid, bindaddr, amount, v, r, s)
	if err != nil {
		log.Errorf("[MakeTransaction]contractabi.Pack error:%s\n", err)
		return err
	}

	//todo store the txData in storage
	//determin the key format
	bf := bytes.NewBuffer(utils.CrossChainManagerContractAddress[:])

	txhash := service.Tx.Hash()
	bf.WriteString(txhash.ToHexString())
	service.CacheDB.Put(bf.Bytes(), txData)

	service.Notifications = append(service.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"makeETHtx", hex.EncodeToString(txData)},
		})

	return nil
}

func verifyMerkleProof(ethproof *ETHProof, blockdata *EthBlock) ([]byte, error) {
	//1. prepare verify account
	nodeList := new(light.NodeList)

	for _, s := range ethproof.AccountProof {
		p := replace0x(s)
		nodeList.Put(nil, ethComm.Hex2Bytes(p))
	}
	ns := nodeList.NodeSet()

	acctKey := crypto.Keccak256(ethComm.Hex2Bytes(replace0x(ethproof.Address)))

	//2. verify account proof
	acctVal, _, err := trie.VerifyProof(ethComm.HexToHash(replace0x(blockdata.StateRoot)), acctKey, ns)
	if err != nil {
		fmt.Printf("[verifyMerkleProof]verify account err:%s\n", err.Error())
		return nil, err
	}

	nounce := new(big.Int)
	_, f := nounce.SetString(replace0x(ethproof.Nonce), 16)
	if !f {
		fmt.Printf("error format of nounce:%s\n", ethproof.Nonce)
		return nil, fmt.Errorf("error format of nounce:%s\n", ethproof.Nonce)
	}

	balance := new(big.Int)
	_, f = balance.SetString(replace0x(ethproof.Balance), 16)
	if !f {
		fmt.Printf("error format of Balance:%s\n", ethproof.Balance)
		return nil, fmt.Errorf("error format of Balance:%s\n", ethproof.Balance)
	}

	storageHash := ethComm.HexToHash(replace0x(ethproof.StorageHash))
	codeHash := ethComm.HexToHash(replace0x(ethproof.CodeHash))
	//construct the account value
	acct := &ProofAccount{
		Nounce:   nounce,
		Balance:  balance,
		Storage:  storageHash,
		Codehash: codeHash,
	}
	acctrlp, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(acctrlp, acctVal) {
		return nil, fmt.Errorf("[verifyMerkleProof]: verify account proof failed, wanted:%v, get:%v", acctrlp, acctVal)
	}
	//3.verify storage proof
	nodeList = new(light.NodeList)

	if len(ethproof.StorageProofs) != 1 {
		return nil, fmt.Errorf("[verifyMerkleProof]: storage proof fmt error")
	}

	sp := ethproof.StorageProofs[0]
	storageKey := crypto.Keccak256(ethComm.Hex2Bytes(replace0x(sp.Key)))
	for _, prf := range sp.Proof {
		nodeList.Put(nil, ethComm.Hex2Bytes(replace0x(prf)))
	}

	ns = nodeList.NodeSet()
	val, _, err := trie.VerifyProof(storageHash, storageKey, ns)
	if err != nil {
		fmt.Printf("[verifyMerkleProof]verify storage failed:%s\n", err.Error())
		return nil, err
	}
	return val, nil
}

func replace0x(s string) string {
	return strings.Replace(strings.ToLower(s), "0x", "", 1)
}

func checkProofResult(result []byte, value string) bool {
	fmt.Println("==checkProofResult==")
	var s []byte
	err := rlp.DecodeBytes(result, &s)
	if err != nil {
		log.Errorf("[checkProofResult]rlp.DecodeBytes error :%s\n", err)
		return false
	}
	log.Infof("s is %v\n", s)
	hash := crypto.Keccak256([]byte(value))
	log.Infof("hash is %v\n", hash)

	return bytes.Equal(s, hash)
}
