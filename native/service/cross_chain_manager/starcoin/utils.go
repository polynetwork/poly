/*
 * Copyright (C) 2021 The poly network Authors
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

package starcoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	cmanager "github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/starcoin"
	stc "github.com/starcoinorg/starcoin-go/client"
	"github.com/starcoinorg/starcoin-go/types"
	"golang.org/x/crypto/sha3"
)

type TransactionInfoProof struct {
	TransactionInfo stc.TransactionInfo
	Proof           string `json:"proof"`
	eventIndex      int    `json:"event_index"`
	EventWithProof  string `json:"event_with_proof"`
	StateWithProof  string `json:"state_with_proof"`
	accessPath      string `json:"access_path"`
}

type AccumulatorProof struct {
	siblings []types.HashValue
}

type EventWithProof struct {
	event types.ContractEvent
	proof AccumulatorProof
}

type Leaf struct {
	requestedKey types.HashValue
	accountBlob  types.HashValue
}

type SparseMerkleProof struct {
	leaf     Leaf
	siblings []types.HashValue
}

type StateProof struct {
	accountState      []byte
	accountProof      SparseMerkleProof
	accountStateProof SparseMerkleProof
}

type StateWithProof struct {
	state []byte
	proof StateProof
}

var SPARSE_MERKLE_PLACEHOLDER_HASH, _ = types.CreateLiteralHash("SPARSE_MERKLE_PLACEHOLDER_HASH")

func verifyFromStarcoinTx(native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32, sideChain *cmanager.SideChain) (*scom.MakeTxParam, error) {
	bestHeader, err := starcoin.GetCurrentHeader(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, get current header fail, error:%s", err)
	}
	bestHeight := uint32(bestHeader.Number)
	if bestHeight < height || bestHeight-height < uint32(sideChain.BlocksToWait-1) {
		return nil, fmt.Errorf("VerifyFromEthProof, transaction is not confirmed, current height: %d, input height: %d", bestHeight, height)
	}

	blockData, err := starcoin.GetHeaderByHeight(native, uint64(height), fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, get header by height, height:%d, error:%s", height, err)
	}

	transactionInfoProof := new(TransactionInfoProof)
	err = json.Unmarshal(proof, transactionInfoProof)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromSTCProof, unmarshal proof error:%s", err)
	}

	_, err = VerifyEventProof(transactionInfoProof, blockData, sideChain.CCMCAddress)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, verifyMerkleProof error:%v", err)
	}

	data := common.NewZeroCopySource(extra)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}

func VerifyEventProof(proof *TransactionInfoProof, data *types.BlockHeader, address []byte) (bool, error) {
	//verify accumulator proof
	accumulatorProof := new(AccumulatorProof)
	err := json.Unmarshal([]byte(proof.Proof), accumulatorProof)
	if err != nil {
		return false, fmt.Errorf("VerifyEventProof, accumulator proof unmarshal error:%v", err)
	}

	transactionHash, _ := types.BcsDeserializeHashValue([]byte(proof.TransactionInfo.TransactionHash))
	verified, err := verifyAccumulator(*accumulatorProof, data.TxnAccumulatorRoot, transactionHash, proof.TransactionInfo.TransactionIndex)
	if err != nil {
		return false, fmt.Errorf("VerifyEventProof, accumulator verfied failure:%v", err)
	}

	if len(proof.EventWithProof) > 0 {
		//verify event proof
		eventWithProof := new(EventWithProof)
		err := json.Unmarshal([]byte(proof.StateWithProof), eventWithProof)
		if err != nil {
			return false, fmt.Errorf("VerifyEventProof, event with proof unmarshal error:%v", err)
		}
		eventHash, err := eventWithProof.event.CryptoHash()
		if err != nil {
			return false, fmt.Errorf("VerifyEventProof, event hash calculate error:%v", err)
		}
		eventRootHash, _ := types.BcsDeserializeHashValue([]byte(proof.TransactionInfo.EventRootHash))
		_, err = verifyAccumulator(eventWithProof.proof, eventRootHash, *eventHash, proof.eventIndex)
		if err != nil {
			return false, fmt.Errorf("VerifyEventProof, event proof verfied failure:%v", err)
		}
	}
	lenAccessPath := len(proof.accessPath)
	lenStateProof := len(proof.StateWithProof)
	if lenStateProof < 1 && lenAccessPath > 0 {
		return false, fmt.Errorf("VerifyEventProof, state_proof is None, cannot verify access_path:%v", proof.accessPath)
	}
	if lenStateProof > 0 && lenAccessPath > 0 {
		//verify state proof
		stateWithProof := new(StateWithProof)
		err := json.Unmarshal([]byte(proof.StateWithProof), stateWithProof)
		if err != nil {
			return false, fmt.Errorf("VerifyEventProof, state with proof unmarshal error:%v", err)
		}
		stateRootHash, err := types.BcsDeserializeHashValue([]byte(proof.TransactionInfo.StateRootHash))
		if err != nil {
			return false, fmt.Errorf("VerifyEventProof, state root hash deserial error:%v", err)
		}
		accessPath, err := types.BcsDeserializeAccessPath([]byte(proof.accessPath))
		if err != nil {
			return false, fmt.Errorf("VerifyEventProof, access path deserial error:%v", err)
		}
		return verifyState(*stateWithProof, &stateRootHash, accessPath)
	}
	return verified, nil
}

func verifyState(proof StateWithProof, expectedRoot *types.HashValue, path types.AccessPath) (bool, error) {
	lenAccountState := len(proof.proof.accountState)
	lenResourceBlob := len(proof.state)
	if lenAccountState < 1 && lenResourceBlob > 0 {
		return false, fmt.Errorf("verifyState, accessed resource should not exists")
	}
	accountAddress := path.Field0
	if lenAccountState > 1 {
		dataPath := path.Field1
		dataPathIndex, err := getPathIndex(dataPath)
		if err != nil {
			return false, fmt.Errorf("verifyState, %v", err)
		}
		accountState, _ := types.BcsDeserializeAccountState(proof.proof.accountState)
		if len(accountState.StorageRoots) <= (dataPathIndex + 1) {
			storageRoot := accountState.StorageRoots[dataPathIndex]
			if storageRoot == nil && lenResourceBlob > 0 {
				return false, fmt.Errorf("verifyState, accessed resource should not exists")
			}
			pathKeyHash, err := keyHash(dataPath)
			if err != nil {
				return false, fmt.Errorf("verifyState, path key hash err: %v", err)
			}
			_, err = verifySparseMerkleProof(proof.proof.accountStateProof, storageRoot, pathKeyHash, proof.state)
			if err != nil {
				return false, fmt.Errorf("verifySparseMerkleProof account state err: %v", err)
			}
		} else {
			return false, fmt.Errorf("verifyState, storage root length too large: %v", accountState.StorageRoots)
		}
	}
	addrKeyHash, err := AddressKeyHash(accountAddress)
	if err != nil {
		return false, fmt.Errorf("verifySparseMerkleProof account address key hash err: %v", err)
	}
	verifySparseMerkleProof(proof.proof.accountProof, expectedRoot, addrKeyHash, proof.proof.accountState)
	return true, nil
}

func verifySparseMerkleProof(proof SparseMerkleProof, expectedRootHash *types.HashValue, elementKey types.HashValue, blob []byte) (bool, error) {
	lenSibling := len(proof.siblings)
	if lenSibling > (32 * 8) {
		return false, fmt.Errorf("verifySparseMerkleProof, siblings length too long: %v", lenSibling)
	}
	lenBlob := len(blob)
	lenRequestKey := len(proof.leaf.requestedKey)
	lenAccountBlob := len(proof.leaf.accountBlob)

	if lenBlob > 0 {
		if lenRequestKey > 0 && lenAccountBlob > 0 {
			if !hashValueEqual(elementKey, proof.leaf.requestedKey) {
				return false, fmt.Errorf("verifySparseMerkleProof, elementKey not equal leaf requestKey: %v, %v", elementKey, proof.leaf.requestedKey)
			}
			hash, err := types.BcsDeserializeHashValue(blob)
			if err != nil {
				return false, fmt.Errorf("verifySparseMerkleProof, block hash err: %v", err)
			}
			if !hashValueEqual(hash, proof.leaf.accountBlob) {
				return false, fmt.Errorf("verifySparseMerkleProof, blob hash not equal: %v, %v", hash, proof.leaf.accountBlob)
			}
		} else {
			return false, fmt.Errorf("verifySparseMerkleProof, Expected inclusion proof. Found non-inclusion proof.")
		}
	} else {
		//non-inclusion proof
		if lenRequestKey > 0 {
			if hashValueEqual(elementKey, proof.leaf.requestedKey) {
				return false, fmt.Errorf("verifySparseMerkleProof, Expected non-inclusion proof, but key exists in proof: %v", elementKey)
			}
			//todo common_prefix_bits
		}
	}

	key := proof.leaf.requestedKey
	value := proof.leaf.accountBlob
	if lenRequestKey < 1 {
		key = *SPARSE_MERKLE_PLACEHOLDER_HASH
	}
	if lenAccountBlob < 1 {
		value = *SPARSE_MERKLE_PLACEHOLDER_HASH
	}
	node := types.SparseMerkleLeafNode{Key: key, ValueHash: value}
	currentHash, _ := node.CryptoHash()

	hash := currentHash
	elementKeyBits := Bytes2Bits(elementKey)
	for i := 0; i < lenSibling; i++ {
		sibling := proof.siblings[i]
		//get element_key bit , skip := 32 * 8 - lenSibling
		bit := elementKeyBits[lenSibling-i]
		if bit != 0 {
			hash, _ = types.SparseMerkleInternalNode{LeftChild: sibling, RightChild: *hash}.CryptoHash()
		} else {
			hash, _ = types.SparseMerkleInternalNode{LeftChild: *hash, RightChild: sibling}.CryptoHash()
		}
	}
	if !hashValueEqual(*expectedRootHash, *hash) {
		return false, fmt.Errorf("Root hashes do not match. Actual root hash: %v. Expected root hash: %v.", *hash, *expectedRootHash)
	}
	return true, nil
}

func Bytes2Bits(data []byte) []int {
	dst := make([]int, 0)
	for _, v := range data {
		for i := 0; i < 8; i++ {
			move := uint(7 - i)
			dst = append(dst, int((v>>move)&1))
		}
	}
	fmt.Println(len(dst))
	return dst
}

func hashValueEqual(hash1, hash2 types.HashValue) bool {
	hash1Bytes, _ := hash1.BcsSerialize()
	hash2Bytes, _ := hash2.BcsSerialize()
	return bytes.Equal(hash1Bytes, hash2Bytes)
}

func AddressKeyHash(address types.AccountAddress) (types.HashValue, error) {
	bytes, _ := address.BcsSerialize()
	return types.HashValue(HashSha(bytes)), nil
}

func keyHash(path types.DataPath) (types.HashValue, error) {
	if IsInstanceOf(path, (*types.DataPath__Code)(nil)) {
		bytes, _ := (path.(*types.DataPath__Code)).Value.BcsSerialize()
		return types.HashValue(HashSha(bytes)), nil
	}
	if IsInstanceOf(path, (*types.DataPath__Resource)(nil)) {
		bytes, _ := (path.(*types.DataPath__Resource)).Value.BcsSerialize()
		return types.HashValue(HashSha(bytes)), nil
	}
	return nil, fmt.Errorf("keyHash wrong instance of data path: %v", path)
}

func HashSha(data []byte) []byte {
	concatData := bytes.Buffer{}
	concatData.Write(data)
	hashData := sha3.Sum256(concatData.Bytes())
	return hashData[:]
}

func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}

func getPathIndex(dataPath types.DataPath) (int, error) {
	if IsInstanceOf(dataPath, (*types.DataPath__Code)(nil)) {
		return 0, nil
	}
	if IsInstanceOf(dataPath, (*types.DataPath__Resource)(nil)) {
		return 1, nil
	}
	return -1, fmt.Errorf("getPathIndex wrong index of data path")
}

func verifyAccumulator(proof AccumulatorProof, expectedRoot types.HashValue, hash types.HashValue, index int) (bool, error) {
	length := len(proof.siblings)
	if length > 63 {
		return false, fmt.Errorf("verifyAccumulator, Accumulator proof has more than (%d) siblings.", length)
	}
	elementIndex := index
	elementHashBytes, _ := hash.BcsSerialize()

	for i := 0; i < length; i++ {
		siblingBytes, _ := proof.siblings[i].BcsSerialize()
		elementHashBytes = internalHash(elementIndex, elementHashBytes, siblingBytes)
		elementIndex = elementIndex / 2
	}
	expectedRootBytes, _ := expectedRoot.BcsSerialize()
	if bytes.Equal(expectedRootBytes, elementHashBytes) {
		return true, nil
	} else {
		return false, fmt.Errorf("verifyAccumulator, root hash not expected: except: %v, actual: %v", expectedRootBytes, elementHashBytes)
	}
}

func internalHash(index int, elements, sibling []byte) []byte {
	if index%2 == 0 {
		return parentHash(elements, sibling)
	} else {
		return parentHash(sibling, elements)
	}
}

func parentHash(left, right []byte) []byte {
	concatData := bytes.Buffer{}
	concatData.Write(left)
	concatData.Write(right)
	hashData := sha3.Sum256(concatData.Bytes())
	return hashData[:]
}
