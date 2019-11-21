/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package node_manager

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/common/serialization"
	"github.com/ontio/multi-chain/core/genesis"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	//status
	RegisterCandidateStatus Status = iota
	ConsensusStatus
)

const (
	//function name
	INIT_CONFIG          = "initConfig"
	REGISTER_CANDIDATE   = "registerCandidate"
	UNREGISTER_CANDIDATE = "unRegisterCandidate"
	APPROVE_CANDIDATE    = "approveCandidate"
	REJECT_CANDIDATE     = "rejectCandidate"
	BLACK_NODE           = "blackNode"
	WHITE_NODE           = "whiteNode"
	QUIT_NODE            = "quitNode"
	UPDATE_CONFIG        = "updateConfig"

	//key prefix
	VBFT_CONFIG     = "vbftConfig"
	CANDIDITE_INDEX = "candidateIndex"
	PEER_POOL       = "peerPool"
	PEER_INDEX      = "peerIndex"
	BLACK_LIST      = "blackList"
)

//Register methods of node_manager contract
func RegisterNodeManagerContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(UNREGISTER_CANDIDATE, UnRegisterCandidate)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(REJECT_CANDIDATE, RejectCandidate)
	native.Register(BLACK_NODE, BlackNode)
	native.Register(WHITE_NODE, WhiteNode)
	native.Register(UPDATE_CONFIG, UpdateConfig)
}

//Init node_manager contract, include vbft config, global param and ontid admin.
func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.GetInput()))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig, contract params deserialize error: %v", err)
	}
	if err := configuration.Deserialize(bytes.NewBuffer(buf)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig, contract params deserialize error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	// check if initConfig is already execute
	peerPoolMapBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(PEER_POOL)))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig, get peerPoolMap error: %v", err)
	}
	if peerPoolMapBytes != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig. initConfig is already executed")
	}

	//check the configuration
	err = CheckVBFTConfig(configuration)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig, checkVBFTConfig failed: %v", err)
	}

	var maxId uint32

	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	for _, peer := range configuration.Peers {
		if peer.Index > maxId {
			maxId = peer.Index
		}
		address, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("initConfig, address format error: %v", err)
		}

		peerPoolItem := new(PeerPoolItem)
		peerPoolItem.Index = peer.Index
		peerPoolItem.PeerPubkey = peer.PeerPubkey
		peerPoolItem.Address = address[:]
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem

		peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("initConfig, peerPubkey format error: %v", err)
		}
		index := peerPoolItem.Index
		indexBytes, err := GetUint32Bytes(index)
		if err != nil {
			return nil, fmt.Errorf("initConfig, getUint32Bytes error: %v", err)
		}
		native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), cstates.GenRawStorageItem(indexBytes))
	}

	//init peer pool
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig, put peerPoolMap error: %v", err)
	}
	indexBytes, err := GetUint32Bytes(maxId + 1)
	if err != nil {
		return nil, fmt.Errorf("initConfig, get indexBytes error: %v", err)
	}
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), cstates.GenRawStorageItem(indexBytes))

	//init config
	config := &Configuration{
		N:                    configuration.N,
		C:                    configuration.C,
		K:                    configuration.K,
		L:                    configuration.L,
		BlockMsgDelay:        configuration.BlockMsgDelay,
		HashMsgDelay:         configuration.HashMsgDelay,
		PeerHandshakeTimeout: configuration.PeerHandshakeTimeout,
		MaxBlockChangeView:   configuration.MaxBlockChangeView,
	}
	err = putConfig(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig, put config error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Register a candidate node, used by users.
func RegisterCandidate(native *native.NativeService) ([]byte, error) {
	err := registerCandidate(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerCandidate error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

//Unregister a registered candidate node, will remove node from pool
func UnRegisterCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RegisterPeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, contract params deserialize error: %v", err)
	}
	address, err := common.AddressParseFromBytes(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, common.AddressParseFromBytes error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, checkWitness error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, get peerPoolMap error: %v", err)
	}

	//check if exist in PeerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, peerPubkey is not in peerPoolMap: %v", err)
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, peer status is not RegisterCandidateStatus")
	}

	//check owner address
	if !bytes.Equal(peerPoolItem.Address, params.Address) {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, address is not peer owner")
	}

	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Approve a registered candidate node
func ApproveCandidate(native *native.NativeService) ([]byte, error) {
	params := new(PeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, checkWitness error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, get peerPoolMap error: %v", err)
	}

	//get peerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, peerPubkey is not in peerPoolMap")
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, peer status is not RegisterCandidateStatus")
	}
	peerPoolItem.Status = ConsensusStatus

	//check if has index
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, peerPubkey format error: %v", err)
	}
	indexBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, get indexBytes error: %v", err)
	}
	if indexBytes != nil {
		value, err := cstates.GetValueFromRawStorageItem(indexBytes)
		if err != nil {
			return nil, fmt.Errorf("approveCandidate, get value from raw storage item error:%v", err)
		}
		index, err := GetBytesUint32(value)
		if err != nil {
			return nil, fmt.Errorf("approveCandidate, get index error: %v", err)
		}
		peerPoolItem.Index = index
	} else {
		//get candidate index
		candidateIndex, err := getCandidateIndex(native, contract)
		if err != nil {
			return nil, fmt.Errorf("approveCandidate, get candidateIndex error: %v", err)
		}
		peerPoolItem.Index = candidateIndex

		//update candidateIndex
		newCandidateIndex := candidateIndex + 1
		err = putCandidateIndex(native, contract, newCandidateIndex)
		if err != nil {
			return nil, fmt.Errorf("approveCandidate, put candidateIndex error: %v", err)
		}

		indexBytes, err := GetUint32Bytes(peerPoolItem.Index)
		if err != nil {
			return nil, fmt.Errorf("approveCandidate, get indexBytes error: %v", err)
		}
		native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), cstates.GenRawStorageItem(indexBytes))
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Reject a registered candidate node, remove node from pool
func RejectCandidate(native *native.NativeService) ([]byte, error) {
	params := new(PeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, checkWitness error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, get peerPoolMap error: %v", err)
	}

	//draw back init pos
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, peerPubkey is not RegisterCandidateStatus")
	}

	//remove peerPubkey from peerPool
	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Put a node into black list, remove node from pool
//Node in black list can't be registered.
func BlackNode(native *native.NativeService) ([]byte, error) {
	params := new(PeerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("blackNode, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("blackNode, checkWitness error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("blackNode, get peerPoolMap error: %v", err)
	}

	for _, peerPubkey := range params.PeerPubkeyList {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("blackNode, peerPubkey format error: %v", err)
		}
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, fmt.Errorf("blackNode, peerPubkey is not in peerPoolMap")
		}

		blackListItem := &BlackListItem{
			PeerPubkey: peerPoolItem.PeerPubkey,
			Address:    peerPoolItem.Address,
		}
		sink := common.NewZeroCopySink(nil)
		blackListItem.Serialization(sink)
		//put peer into black list
		native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix), cstates.GenRawStorageItem(sink.Bytes()))

		delete(peerPoolMap.PeerPoolMap, peerPubkey)
	}
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("blackNode, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Remove a node from black list, allow it to be registered
func WhiteNode(native *native.NativeService) ([]byte, error) {
	params := new(PeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, checkWitness error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, peerPubkey format error: %v", err)
	}

	//check black list
	blackListBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, get BlackList error: %v", err)
	}
	if blackListBytes == nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, this Peer is not in BlackList: %v", err)
	}

	//remove peer from black list
	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))

	return utils.BYTE_TRUE, nil
}

//Quit a registered node, used by node owner.
//Remove node from pool
func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(RegisterPeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, contract params deserialize error: %v", err)
	}
	address, err := common.AddressParseFromBytes(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, common.AddressParseFromBytes error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, checkWitness error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, get peerPoolMap error: %v", err)
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not in peerPoolMap")
	}

	if !bytes.Equal(params.Address, peerPoolItem.Address) {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not registered by this address")
	}
	if peerPoolItem.Status != ConsensusStatus {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not ConsensusStatus")
	}

	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Update VBFT config
func UpdateConfig(native *native.NativeService) ([]byte, error) {
	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig, checkWitness error: %v", err)
	}
	contract := utils.NodeManagerContractAddress

	configuration := new(Configuration)
	if err := configuration.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig, deserialize configuration error: %v", err)
	}

	//check the configuration
	if configuration.C == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. C can not be 0 in config")
	}
	if configuration.L < 16*configuration.K || configuration.L%configuration.K != 0 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. L can not be less than 16*K and K must be times of L in config")
	}
	if configuration.K < 2*configuration.C+1 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. K can not be less than 2*C+1 in config")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. config not match N >= K >= 7")
	}
	if configuration.BlockMsgDelay < 5000 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. BlockMsgDelay must >= 5000")
	}
	if configuration.HashMsgDelay < 5000 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. HashMsgDelay must >= 5000")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. PeerHandshakeTimeout must >= 10")
	}

	err = putConfig(native, contract, configuration)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPreConfig, put preConfig error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}
