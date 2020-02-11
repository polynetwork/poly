package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	ecom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/header_sync/eth"
)

func verifyFromEthTx(native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32) (*scom.MakeTxParam, error) {
	blockData, _, err := eth.GetHeaderByHeight(native, uint64(height), fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, get header by height, height:%d, error:%s", height, err)
	}

	ethProof := new(ETHProof)
	err = json.Unmarshal(proof, ethProof)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, unmarshal proof error:%s", err)
	}

	if len(ethProof.StorageProofs) != 1 {
		return nil, fmt.Errorf("VerifyFromEthProof, incorrect proof format")
	}

	//todo 1. verify the proof with header
	//determine where the k and v from
	proofResult, err := verifyMerkleProof(ethProof, blockData)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, verifyMerkleProof error:%v", err)
	}
	if proofResult == nil {
		return nil, fmt.Errorf("VerifyFromEthProof, verifyMerkleProof failed!")
	}

	if !checkProofResult(proofResult, extra) {
		return nil, fmt.Errorf("VerifyFromEthProof, verify proof value hash failed, proof result:%x, extra:%x", proofResult, extra)
	}

	data := common.NewZeroCopySource(extra)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}

func verifyMerkleProof(ethProof *ETHProof, blockData *types.Header) ([]byte, error) {
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

func checkProofResult(result, value []byte) bool {
	var s_temp []byte
	err := rlp.DecodeBytes(result, &s_temp)
	if err != nil {
		log.Errorf("checkProofResult, rlp.DecodeBytes error:%s\n", err)
		return false
	}
	//
	var s []byte
	for i := len(s_temp);i < 32;i ++ {
		s = append(s, 0)
	}
	s = append(s, s_temp...)
	hash := crypto.Keccak256(value)
	return bytes.Equal(s, hash)
}
