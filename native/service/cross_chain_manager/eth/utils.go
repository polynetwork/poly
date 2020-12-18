/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	ecom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	cmanager "github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/eth"
)

func verifyFromEthTx(native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32, sideChain *cmanager.SideChain) (*scom.MakeTxParam, error) {
	bestHeader, _, err := eth.GetCurrentHeader(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, get current header fail, error:%s", err)
	}
	bestHeight := uint32(bestHeader.Number.Uint64())
	if bestHeight < height || bestHeight-height < uint32(sideChain.BlocksToWait-1) {
		return nil, fmt.Errorf("VerifyFromEthProof, transaction is not confirmed, current height: %d, input height: %d", bestHeight, height)
	}

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
	proofResult, err := VerifyMerkleProof(ethProof, blockData, sideChain.CCMCAddress)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, verifyMerkleProof error:%v", err)
	}
	if proofResult == nil {
		return nil, fmt.Errorf("VerifyFromEthProof, verifyMerkleProof failed!")
	}

	if !CheckProofResult(proofResult, extra) {
		return nil, fmt.Errorf("VerifyFromEthProof, verify proof value hash failed, proof result:%x, extra:%x", proofResult, extra)
	}

	data := common.NewZeroCopySource(extra)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}

func VerifyMerkleProof(ethProof *ETHProof, blockData *types.Header, contractAddr []byte) ([]byte, error) {
	//1. prepare verify account
	nodeList := new(light.NodeList)

	for _, s := range ethProof.AccountProof {
		p := scom.Replace0x(s)
		nodeList.Put(nil, ecom.Hex2Bytes(p))
	}
	ns := nodeList.NodeSet()

	addr := ecom.Hex2Bytes(scom.Replace0x(ethProof.Address))
	if !bytes.Equal(addr, contractAddr) {
		return nil, fmt.Errorf("verifyMerkleProof, contract address is error, proof address: %s, side chain address: %s", ethProof.Address, hex.EncodeToString(contractAddr))
	}
	acctKey := crypto.Keccak256(addr)

	// 2. verify account proof
	acctVal, err := trie.VerifyProof(blockData.Root, acctKey, ns)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify account proof error:%s\n", err)
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
	storageKey := crypto.Keccak256(ecom.HexToHash(scom.Replace0x(sp.Key)).Bytes())

	for _, prf := range sp.Proof {
		nodeList.Put(nil, ecom.Hex2Bytes(scom.Replace0x(prf)))
	}

	ns = nodeList.NodeSet()
	val, err := trie.VerifyProof(storageHash, storageKey, ns)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify storage proof error:%s\n", err)
	}

	return val, nil
}

func CheckProofResult(result, value []byte) bool {
	var s_temp []byte
	err := rlp.DecodeBytes(result, &s_temp)
	if err != nil {
		log.Errorf("checkProofResult, rlp.DecodeBytes error:%s\n", err)
		return false
	}
	//
	var s []byte
	for i := len(s_temp); i < 32; i++ {
		s = append(s, 0)
	}
	s = append(s, s_temp...)
	hash := crypto.Keccak256(value)
	return bytes.Equal(s, hash)
}
