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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"

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
	TransactionInfo stc.TransactionInfo `json:"transaction_info"`
	Proof           Siblings            `json:"proof"`
	EventWithProof  EventWithProof      `json:"event_proof"`
	StateWithProof  StateWithProofJson  `json:"state_proof"`
	AccessPath      *string             `json:"access_path,omitempty"`
	EventIndex      *int                `json:"event_index,omitempty"`
}

type EventWithProof struct {
	Event string   `json:"event"`
	Proof Siblings `json:"proof"`
}

type Siblings struct {
	Sibling []string `json:"siblings"`
}

type AccumulatorProof struct {
	siblings []types.HashValue
}

type TypeTag_Struct struct {
	Value StructTag `json:"Struct"`
}

type StructTag struct {
	Address    string   `json:"address"`
	Module     string   `json:"module"`
	Name       string   `json:"name"`
	TypeParams []string `json:"type_params"`
}

type Event struct {
	Key            string         `json:"key"`
	SequenceNumber int            `json:"sequence_number"`
	TypeTag        TypeTag_Struct `json:"type_tag"`
	EventData      []byte         `json:"event_data"`
}

type ContractEvent struct {
	V Event `json:"V0"`
}

type SparseMerkleProofJson struct {
	Leaf     []string `json:"leaf"`
	Siblings []string `json:"siblings"`
}

type StateProofJson struct {
	AccountState      []byte                `json:"account_state"`
	AccountProof      SparseMerkleProofJson `json:"account_proof"`
	AccountStateProof SparseMerkleProofJson `json:"account_state_proof"`
}

type StateWithProofJson struct {
	State []byte         `json:"state"`
	Proof StateProofJson `json:"proof"`
}

func hexToAccountAddress(addr string) (*types.AccountAddress, error) {
	accountBytes, err := hexToBytes(addr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var addressArray types.AccountAddress
	copy(addressArray[:], accountBytes[:16])
	return &addressArray, nil
}

func (tag *StructTag) toTypesStructTag() (*types.TypeTag__Struct, error) {
	address, err := hexToAccountAddress(tag.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	structTag := types.StructTag{
		Address:    *address,
		Module:     types.Identifier(tag.Module),
		Name:       types.Identifier(tag.Name),
		TypeParams: nil, //todo parse typetag[]
	}
	return &types.TypeTag__Struct{
		Value: structTag,
	}, nil
}

func toTypesContractEvent(event *types.ContractEventV0) *types.ContractEvent__V0 {
	return &types.ContractEvent__V0{
		Value: types.ContractEventV0{
			Key:            event.Key,
			SequenceNumber: event.SequenceNumber,
			TypeTag:        event.TypeTag,
			EventData:      event.EventData,
		},
	}
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

func verifyFromStarcoinTx(native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32, sideChain *cmanager.SideChain, headerOrCrossChainMsg []byte) (*scom.MakeTxParam, error) {
	bestHeader, err := starcoin.GetCurrentHeader(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("verifyFromStarcoinTx, get current header fail, error:%s", err)
	}
	bestHeight := uint32(bestHeader.BlockHeader.Number)
	if bestHeight < height || bestHeight-height < uint32(sideChain.BlocksToWait-1) {
		return nil, fmt.Errorf("verifyFromStarcoinTx, transaction is not confirmed, current height: %d, input height: %d", bestHeight, height)
	}

	blockData, err := starcoin.GetHeaderByHeight(native, uint64(height), fromChainID)
	if err != nil {
		return nil, fmt.Errorf("verifyFromStarcoinTx, get header by height, height:%d, error:%s", height, err)
	}

	transactionInfoProof := new(TransactionInfoProof)
	if err = json.Unmarshal(proof, transactionInfoProof); err != nil {
		return nil, fmt.Errorf("verifyFromStarcoinTx, unmarshal proof error:%s", err)
	}
	// ////////////////////////////////
	headerOrMsg := new(StarcoinToPolyHeaderOrCrossChainMsg)
	if err = json.Unmarshal(headerOrCrossChainMsg, headerOrMsg); err != nil {
		return nil, fmt.Errorf("verifyFromStarcoinTx, unmarshal headerOrCrossChainMsg error:%s", err)
	}
	transactionInfoProof.EventIndex = headerOrMsg.EventIndex
	transactionInfoProof.AccessPath = headerOrMsg.AccessPath
	// ///////////////////////////////
	typeEventV0, err := VerifyEventProof(transactionInfoProof, blockData.BlockHeader.TxnAccumulatorRoot, sideChain.CCMCAddress)
	if err != nil {
		return nil, fmt.Errorf("verifyFromStarcoinTx, verifyMerkleProof error:%v", err)
	}
	// ///////////////////
	eventData := typeEventV0.EventData
	eventRawData, err := GetCrossChainEventRawData(eventData)
	if err != nil || len(eventRawData) == 0 {
		return nil, fmt.Errorf("verifyFromStarcoinTx, get event RawData error:%x", eventData)
	}
	zcsrc := common.NewZeroCopySource(eventRawData)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(zcsrc); err != nil {
		return nil, fmt.Errorf("verifyFromStarcoinTx, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}

func GetCrossChainEventRawData(ccEvtData []byte) ([]byte, error) {
	event, err := BcsDeserializeCrossChainEvent(ccEvtData)
	if err != nil {
		return nil, fmt.Errorf("CheckEventData, deserialize error:%s", err.Error())
	}
	return event.RawData, nil
}

// func CheckCrossChainEventRawData(ccEvtData []byte, value []byte) (bool, error) {
// 	event, err := BcsDeserializeCrossChainEvent(ccEvtData)
// 	if err != nil {
// 		return false, fmt.Errorf("CheckEventData, deserialize error:%s", err)
// 	}
// 	if bytes.Equal(event.RawData, value) {
// 		return true, nil
// 	}
// 	return false, fmt.Errorf("CheckEventData, check error:%v, param: %v", event, value)
// }

// func CheckEventData(data []byte, param *scom.MakeTxParam) (bool, error) {
// 	event, err := BcsDeserializeCrossChainEvent(data)
// 	if err != nil {
// 		return false, fmt.Errorf("CheckEventData, deserialize error:%s", err)
// 	}
// 	if event.ToChainId == param.ToChainID && bytes.Equal(event.TxId, param.TxHash) && bytes.Equal(event.Sender, param.FromContractAddress) && bytes.Equal(event.ToContract, param.ToContractAddress) {
// 		return true, nil
// 	}
// 	return false, fmt.Errorf("CheckEventData, check error:%v, param: %v", event, param)
// }

func VerifyEventProof(proof *TransactionInfoProof, txnAccumulatorRoot types.HashValue, address []byte) (*types.ContractEventV0, error) {
	//var eventData []byte
	//var eventKey []byte
	//verify accumulator proof
	accumulatorProof := toSiblings(proof.Proof)
	transactionInfo, err := proof.TransactionInfo.ToTypesTransactionInfo()
	if err != nil {
		return nil, fmt.Errorf("VerifyEventProof, to types transaction info err:%v", err)
	}
	transactionHash, err := transactionInfo.CryptoHash()
	if err != nil {
		return nil, fmt.Errorf("VerifyEventProof, transaction info crypto hash error:%v", err)
	}
	globalIndex, err := strconv.Atoi(proof.TransactionInfo.TransactionGlobalIndex)
	if err != nil {
		return nil, fmt.Errorf("VerifyEventProof, transaction info global index transfer err:%v", err)
	}
	if _, err := verifyAccumulator(*accumulatorProof, txnAccumulatorRoot, *transactionHash, globalIndex); err != nil {
		return nil, fmt.Errorf("VerifyEventProof, accumulator verfied failure:%v", err)
	}
	var typeEventV0 *types.ContractEventV0
	if len(proof.EventWithProof.Event) > 0 && len(proof.EventWithProof.Proof.Sibling) > 0 {
		//verify event proof
		eventProof := toSiblings(proof.EventWithProof.Proof)
		eventByte, err := hexToBytes(proof.EventWithProof.Event)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, event decode error:%v", err)
		}
		typeEventV0, err = stc.EventToContractEventV0(eventByte)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, event to types.ContractEventV0 error:%v", err)
		}
		typeEvent := toTypesContractEvent(typeEventV0)
		eventHash, err := typeEvent.CryptoHash()
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, event crypto hash error:%v", err)
		}
		eventRootHash, err := toHashValue(proof.TransactionInfo.EventRootHash)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, event root hash deserialize error:%v", err)
		}
		if proof.EventIndex == nil {
			return nil, fmt.Errorf("VerifyEventProof, event index is nil")
		}
		if _, err = verifyAccumulator(*eventProof, eventRootHash, *eventHash, *proof.EventIndex); err != nil {
			return nil, fmt.Errorf("VerifyEventProof, event proof verfied failure:%v", err)
		}
		//eventData = typeEvent.Value.EventData
		//eventKey = typeEvent.Value.Key
	}

	// check event TypeTag:
	eventTt, err := getEventTypeTagString(typeEventV0.TypeTag)
	// TypeTag string like this: "0x3809644a7409cca52138ce747c56eaf2::CrossChainManager::CrossChainEvent"
	if err != nil {
		return nil, fmt.Errorf("VerifyEventProof, getEventTypeTagString error:%v", err)
	}
	if eventTt != string(address) {
		return nil, fmt.Errorf("VerifyEventProof, event TypeTag error:%s", eventTt)
	}

	lenAccessPath := 0
	if proof.AccessPath != nil {
		lenAccessPath = len(*proof.AccessPath)
	}
	lenStateProof := len(proof.StateWithProof.State)
	if lenStateProof < 1 && lenAccessPath > 0 {
		return nil, fmt.Errorf("VerifyEventProof, state_proof is None, cannot verify access_path:%v", proof.AccessPath)
	}
	if lenStateProof > 0 && lenAccessPath > 0 {
		//verify state proof
		stateWithProof, err := toStateProof(proof.StateWithProof)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, state with proof unmarshal error:%v", err)
		}
		stateRootHash, err := toHashValue(proof.TransactionInfo.StateRootHash)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, state root hash deserial error:%v", err)
		}
		hexAccessPath, err := hex.DecodeString(*proof.AccessPath)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, access path hex decode error:%v", err)
		}
		accessPath, err := types.BcsDeserializeAccessPath(hexAccessPath)
		if err != nil {
			return nil, fmt.Errorf("VerifyEventProof, access path deserial error:%v", err)
		}
		if _, err = verifyState(stateWithProof, &stateRootHash, accessPath); err != nil {
			return nil, fmt.Errorf("VerifyEventProof, verify state error:%v", err)
		}

	}
	return typeEventV0, nil
}

func getEventTypeTagString(tt types.TypeTag) (string, error) {
	switch tt := tt.(type) {
	case *types.TypeTag__Struct:
		return "0x" + hex.EncodeToString(tt.Value.Address[:]) + "::" + string(tt.Value.Module) + "::" + string(tt.Value.Name), nil
	default:
		return "", fmt.Errorf("unknown TypeTag type")
	}
}

func toStateProof(proofJson StateWithProofJson) (StateWithProof, error) {
	var stateWithProof StateWithProof
	accountProofLeaf, err := toHashValues(proofJson.Proof.AccountProof.Leaf)
	if err != nil {
		return stateWithProof, fmt.Errorf("toStateProof, accountProofLeaf parsee error:%v", err)
	}
	if len(accountProofLeaf) != 2 {
		return stateWithProof, fmt.Errorf("toStateProof, accountProofLeaf length error:%v", accountProofLeaf)
	}
	accountProofSiblings, err := toHashValues(proofJson.Proof.AccountProof.Siblings)
	if err != nil {
		return stateWithProof, fmt.Errorf("toStateProof, accountProofSiblings parsee error:%v", err)
	}
	accountStateProofLeaf, err := toHashValues(proofJson.Proof.AccountStateProof.Leaf)
	if err != nil {
		return stateWithProof, fmt.Errorf("toStateProof, accountStateProofLeaf parsee error:%v", err)
	}
	if len(accountStateProofLeaf) != 2 {
		return stateWithProof, fmt.Errorf("toStateProof, accountStateProofLeaf length error:%v", accountStateProofLeaf)
	}
	accountStateProofSiblings, err := toHashValues(proofJson.Proof.AccountStateProof.Siblings)
	if err != nil {
		return stateWithProof, fmt.Errorf("toStateProof, accountStateProofSiblings parsee error:%v", err)
	}
	return StateWithProof{
		state: proofJson.State,
		proof: StateProof{
			accountState: proofJson.Proof.AccountState,
			accountProof: SparseMerkleProof{
				leaf:     Leaf{accountProofLeaf[0], accountProofLeaf[1]},
				siblings: accountProofSiblings,
			},
			accountStateProof: SparseMerkleProof{
				leaf:     Leaf{accountStateProofLeaf[0], accountStateProofLeaf[1]},
				siblings: accountStateProofSiblings,
			},
		},
	}, nil
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
			return false, fmt.Errorf("verifyState, get path index err: %v", err)
		}
		accountState, err := types.BcsDeserializeAccountState(proof.proof.accountState)
		if err != nil {
			return false, fmt.Errorf("verifyState, account state deserialize err: %v", err)
		}
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
	return verifySparseMerkleProof(proof.proof.accountProof, expectedRoot, addrKeyHash, proof.proof.accountState)
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
			hash := types.HashValue(blob)
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
	println(hex.EncodeToString(*expectedRootHash))
	if !hashValueEqual(*expectedRootHash, *hash) {
		return false, fmt.Errorf("Root hashes do not match. Actual root hash: %v. Expected root hash: %v.", *hash, *expectedRootHash)
	}
	return true, nil
}

func toSiblings(obj Siblings) *AccumulatorProof {
	size := len(obj.Sibling)
	var hashes []types.HashValue
	for i := 0; i < size; i++ {
		hash, _ := toHashValue(obj.Sibling[i])
		hashes = append(hashes, hash)
	}
	proof := AccumulatorProof{siblings: hashes}
	return &proof
}

func hexToBytes(h string) ([]byte, error) {
	var bs []byte
	var err error
	if !strings.HasPrefix(h, "0x") {
		bs, err = hex.DecodeString(h)
	} else {
		bs, err = hex.DecodeString(h[2:])
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return bs, nil
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

func toHashValue(hash string) (types.HashValue, error) {
	hashByes, err := hex.DecodeString(strings.Replace(hash, "0x", "", 1))
	if err != nil {
		return nil, err
	}
	return types.HashValue(hashByes), nil
}
func toHashValues(hashes []string) ([]types.HashValue, error) {
	var result []types.HashValue
	for i := 0; i < len(hashes); i++ {
		hashByes, err := hex.DecodeString(strings.Replace(hashes[i], "0x", "", 1))
		if err != nil {
			return nil, err
		}
		result = append(result, types.HashValue(hashByes))
	}
	return result, nil
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
	elementHashBytes := []byte(hash)

	for i := 0; i < length; i++ {
		println(elementIndex)
		println(hex.EncodeToString(elementHashBytes))
		siblingBytes := []byte(proof.siblings[i])
		elementHashBytes = internalHash(elementIndex, elementHashBytes, siblingBytes)
		elementIndex = elementIndex / 2
	}
	expectedRootBytes := []byte(expectedRoot)
	println(hex.EncodeToString(elementHashBytes))
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

type StarcoinToPolyHeaderOrCrossChainMsg struct {
	EventIndex *int    `json:"event_index,omitempty"`
	AccessPath *string `json:"access_path,omitempty"`
}
