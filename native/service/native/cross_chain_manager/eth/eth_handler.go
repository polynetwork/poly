package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/native"
	crosscommon "github.com/ontio/multi-chain/native/service/native/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/native/cross_chain_manager/eth/locker"
	"github.com/ontio/multi-chain/native/service/native/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/native/utils"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) Vote(service *native.NativeService) (bool, *crosscommon.MakeTxParam, error) {
	params := new(crosscommon.VoteParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return false, nil, fmt.Errorf("eth Vote, contract params deserialize error: %v", err)
	}

	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return false, nil, fmt.Errorf("eth Vote, common.AddressFromBase58 error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(service, address)
	if err != nil {
		return false, nil, fmt.Errorf("eth Vote, utils.ValidateOwner error: %v", err)
	}

	vote, err := getEthVote(service, params.TxHash)
	if err != nil {
		return false, nil, fmt.Errorf("eth Vote, getEthVote error: %v", err)
	}
	vote.VoteMap[params.Address] = params.Address
	err = putEthVote(service, params.TxHash, vote)
	if err != nil {
		return false, nil, fmt.Errorf("eth Vote, putEthVote error: %v", err)
	}

	err = crosscommon.ValidateVote(service, vote)
	if err != nil {
		return false, nil, fmt.Errorf("eth Vote, ValidateVote error: %v", err)
	}

	proofBytes, err := getEthProof(service, params.TxHash)
	if err != nil {
		return false, nil, fmt.Errorf("eth Vote, getEth Tx error: %v", err)
	}

	proof := &Proof{}
	if err := proof.Deserialize(string(proofBytes)); err != nil {
		return false, nil, fmt.Errorf("eth Vote, eth proof deserialize error: %v", err)
	}

	return true, &crosscommon.MakeTxParam{
		FromChainID:         params.FromChainID,
		FromContractAddress: proof.FromAddress,
		ToChainID:           proof.ToChainID,
		ToAddress:           proof.ToAddress,
		Amount:              proof.Amount,
	}, nil
}

func (this *ETHHandler) MakeDepositProposal(service *native.NativeService) (*crosscommon.MakeTxParam, error) {
	params := new(crosscommon.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, contract params deserialize error: %v", err)
	}

	blockData, err := GetEthBlockByNumber(params.Height)
	if err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, GetEthBlockByNumber error:%v", err)
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
		return nil, fmt.Errorf("MakeDepositProposal, incorrect proof format")
	}

	bf := bytes.NewBuffer(utils.CrossChainManagerContractAddress[:])
	keyBytes := ethcommon.Hex2Bytes(crosscommon.KEY_PREFIX_ETH + crosscommon.Replace0x(ethProof.StorageProofs[0].Key))
	bf.Write(keyBytes)
	key := bf.Bytes()
	val, err := service.CacheDB.Get(key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return nil, fmt.Errorf("MakeDepositProposal, key:%s already solved ", ethProof.StorageProofs[0].Key)
	}
	//todo 1. verify the proof with header
	//determine where the k and v from
	proofResult, err := verifyMerkleProof(ethProof, blockData)
	if err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, verifyMerkleProof error:%v", err)
	}
	if proofResult == nil {
		return nil, fmt.Errorf("MakeDepositProposal, verifyMerkleProof failed!")
	}

	if !checkProofResult(proofResult, params.Value) {
		return nil, fmt.Errorf("MakeDepositProposal, verify proof value hash failed!")
	}

	proof := &Proof{}
	if err := proof.Deserialize(params.Value); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, eth proof deserialize error:%v", err)
	}

	fromContractAddr := strings.Split(params.Value, "#")[0]

	rawTxValue := crypto.Keccak256([]byte(params.Value))
	key = utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.KEY_PREFIX_ETH), rawTxValue)
	service.CacheDB.Put(key, []byte(params.Value))

	ret := &crosscommon.MakeTxParam{}
	ret.ToChainID = proof.ToChainID
	ret.FromContractAddress = fromContractAddr
	ret.FromChainID = params.SourceChainID
	ret.ToAddress = proof.ToAddress
	ret.Amount = proof.Amount
	//todo deal with the proof.decimal

	return ret, nil
}

func (this *ETHHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam) error {
	//todo add logic

	//1 construct tx
	contractabi, err := abi.JSON(strings.NewReader(locker.EthereumCrossChainABI))
	if err != nil {
		return err
	}

	bindaddr := ethcommon.HexToAddress(param.ToAddress)
	amount := param.Amount
	//lockAddress := ethComm.HexToAddress(LOCKER_CONTRACT_ADDR)

	targetAsset, err := side_chain_manager.GetDestAsset(service, param.FromChainID, param.ToChainID, param.FromContractAddress)
	if err != nil {
		return err
	}

	tokenAddress := ethcommon.HexToAddress(targetAsset.ContractAddress)

	txid := "1"
	v := []uint8{0}
	r := [][32]byte{[32]byte{0}}
	s := [][32]byte{[32]byte{0}}
	txData, err := contractabi.Pack("Withdraw", tokenAddress, txid, bindaddr, amount, v, r, s)
	if err != nil {
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
		p := crosscommon.Replace0x(s)
		nodeList.Put(nil, ethcommon.Hex2Bytes(p))
	}
	ns := nodeList.NodeSet()

	acctKey := crypto.Keccak256(ethcommon.Hex2Bytes(crosscommon.Replace0x(ethproof.Address)))

	//2. verify account proof
	acctVal, _, err := trie.VerifyProof(ethcommon.HexToHash(crosscommon.Replace0x(blockdata.StateRoot)), acctKey, ns)
	if err != nil {
		fmt.Printf("verifyMerkleProof, verify account proof error:%s\n", err.Error())
		return nil, err
	}

	nounce := new(big.Int)
	_, ok := nounce.SetString(crosscommon.Replace0x(ethproof.Nonce), 16)
	if !ok {
		return nil, fmt.Errorf("verifyMerkleProof, invalid format of nounce:%s\n", ethproof.Nonce)
	}

	balance := new(big.Int)
	_, ok = balance.SetString(crosscommon.Replace0x(ethproof.Balance), 16)
	if !ok {
		return nil, fmt.Errorf("verifyMerkleProof, invalid format of balance:%s\n", ethproof.Balance)
	}

	storageHash := ethcommon.HexToHash(crosscommon.Replace0x(ethproof.StorageHash))
	codeHash := ethcommon.HexToHash(crosscommon.Replace0x(ethproof.CodeHash))

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
		return nil, fmt.Errorf("verifyMerkleProof, verify account proof failed, wanted:%v, get:%v", acctrlp, acctVal)
	}

	//3.verify storage proof
	nodeList = new(light.NodeList)
	if len(ethproof.StorageProofs) != 1 {
		return nil, fmt.Errorf("verifyMerkleProof, invalid storage proof format")
	}

	sp := ethproof.StorageProofs[0]
	storageKey := crypto.Keccak256(ethcommon.Hex2Bytes(crosscommon.Replace0x(sp.Key)))

	for _, prf := range sp.Proof {
		nodeList.Put(nil, ethcommon.Hex2Bytes(crosscommon.Replace0x(prf)))
	}

	ns = nodeList.NodeSet()
	val, _, err := trie.VerifyProof(storageHash, storageKey, ns)
	if err != nil {
		fmt.Printf("verifyMerkleProof, verify storage proof error:%s\n", err.Error())
		return nil, err
	}

	return val, nil
}

func checkProofResult(result []byte, value string) bool {
	var s []byte
	err := rlp.DecodeBytes(result, &s)
	if err != nil {
		log.Errorf("checkProofResult, rlp.DecodeBytes error:%s\n", err)
		return false
	}
	hash := crypto.Keccak256([]byte(value))

	return bytes.Equal(s, hash)
}
