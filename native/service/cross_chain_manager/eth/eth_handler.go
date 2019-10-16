package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/multi-chain/native/service/header_sync/eth"
	"math/big"

	ecom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	NOTIFY_ETH_PROOF = "notifyEthProof"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, contract params deserialize error: %v", err)
	}

	blockData, err := eth.GetHeaderByHeight(service, uint64(params.Height))
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
	keyBytes := ecom.Hex2Bytes(scom.KEY_PREFIX_ETH + scom.Replace0x(ethProof.StorageProofs[0].Key))
	bf.Write(keyBytes)
	key := bf.Bytes()
	val, err := service.GetCacheDB().Get(key)
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

	rawTxValue := crypto.Keccak256([]byte(params.Value))
	key = utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(scom.KEY_PREFIX_ETH), rawTxValue)
	service.GetCacheDB().Put(key, []byte(params.Value))

	notifyEthroof(service, hex.EncodeToString([]byte(params.Value)))
	return nil, nil
}

func verifyMerkleProof(ethProof *ETHProof, blockData types.Header) ([]byte, error) {
	//1. prepare verify account
	nodeList := new(light.NodeList)

	for _, s := range ethProof.AccountProof {
		p := scom.Replace0x(s)
		nodeList.Put(nil, ecom.Hex2Bytes(p))
	}
	ns := nodeList.NodeSet()

	acctKey := crypto.Keccak256(ecom.Hex2Bytes(scom.Replace0x(ethProof.Address)))

	//2. verify account proof
	acctVal, _, err := trie.VerifyProof(blockData.Root, acctKey, ns)
	if err != nil {
		return nil, fmt.Errorf("F, verify account proof error:%s\n", err)
	}

	nounce := new(big.Int)
	_, ok := nounce.SetString(scom.Replace0x(ethProof.Nonce), 16)
	if !ok {
		return nil, fmt.Errorf("verifyMerkleProof, invalid format of nounce:%s\n", ethProof.Nonce)
	}

	balance := new(big.Int)
	_, ok = balance.SetString(scom.Replace0x(ethProof.Balance), 16)
	if !ok {
		return nil, fmt.Errorf("verifyMerkleProof, invalid format of balance:%s\n", ethProof.Balance)
	}

	storageHash := ecom.HexToHash(scom.Replace0x(ethProof.StorageHash))
	codeHash := ecom.HexToHash(scom.Replace0x(ethProof.CodeHash))

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
	if len(ethProof.StorageProofs) != 1 {
		return nil, fmt.Errorf("verifyMerkleProof, invalid storage proof format")
	}

	sp := ethProof.StorageProofs[0]
	storageKey := crypto.Keccak256(ecom.Hex2Bytes(scom.Replace0x(sp.Key)))

	for _, prf := range sp.Proof {
		nodeList.Put(nil, ecom.Hex2Bytes(scom.Replace0x(prf)))
	}

	ns = nodeList.NodeSet()
	val, _, err := trie.VerifyProof(storageHash, storageKey, ns)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify storage proof error:%s\n", err)
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
