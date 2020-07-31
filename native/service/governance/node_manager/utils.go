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

package node_manager

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/polynetwork/poly/native/event"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/utils"
)

func GetPeerApply(native *native.NativeService, peerPubkey string) (*RegisterPeerParam, error) {
	contract := utils.NodeManagerContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("GetPeerApply, peerPubkey format error: %v", err)
	}
	peerBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(PEER_APPLY), peerPubkeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("GetPeerApply, get peer error: %v", err)
	}
	if peerBytes == nil {
		return nil, nil
	}
	peerStore, err := cstates.GetValueFromRawStorageItem(peerBytes)
	if err != nil {
		return nil, fmt.Errorf("GetPeerApply, deserialize from raw storage item err:%v", err)
	}
	peer := new(RegisterPeerParam)
	if err := peer.Deserialization(common.NewZeroCopySource(peerStore)); err != nil {
		return nil, fmt.Errorf("GetPeerApply, deserialize peer error: %v", err)
	}
	return peer, nil
}

func putPeerApply(native *native.NativeService, peer *RegisterPeerParam) error {
	contract := utils.NodeManagerContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(peer.PeerPubkey)
	if err != nil {
		return fmt.Errorf("putPeerApply, peerPubkey format error: %v", err)
	}
	sink := common.NewZeroCopySink(nil)
	peer.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(PEER_APPLY), peerPubkeyPrefix), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetPeerPoolMap(native *native.NativeService, view uint32) (*PeerPoolMap, error) {
	contract := utils.NodeManagerContractAddress
	viewBytes := utils.GetUint32Bytes(view)
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	peerPoolMapBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes))
	if err != nil {
		return nil, fmt.Errorf("getPeerPoolMap, get all peerPoolMap error: %v", err)
	}
	if peerPoolMapBytes == nil {
		return nil, fmt.Errorf("getPeerPoolMap, peerPoolMap is nil")
	}
	item := cstates.StorageItem{}
	err = item.Deserialize(bytes.NewBuffer(peerPoolMapBytes))
	if err != nil {
		return nil, fmt.Errorf("deserialize PeerPoolMap error:%v", err)
	}
	peerPoolMapStore := item.Value
	if err := peerPoolMap.Deserialization(common.NewZeroCopySource(peerPoolMapStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize peerPoolMap error: %v", err)
	}
	return peerPoolMap, nil
}

func putPeerPoolMap(native *native.NativeService, peerPoolMap *PeerPoolMap, view uint32) {
	contract := utils.NodeManagerContractAddress
	viewBytes := utils.GetUint32Bytes(view)
	sink := common.NewZeroCopySink(nil)
	peerPoolMap.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), cstates.GenRawStorageItem(sink.Bytes()))
}

func CheckVBFTConfig(configuration *config.VBFTConfig) error {
	if configuration.BlockMsgDelay < 5000 {
		return fmt.Errorf("initConfig. BlockMsgDelay must >= 5000")
	}
	if configuration.HashMsgDelay < 5000 {
		return fmt.Errorf("initConfig. HashMsgDelay must >= 5000")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return fmt.Errorf("initConfig. PeerHandshakeTimeout must >= 10")
	}
	if len(configuration.VrfProof) < 128 {
		return fmt.Errorf("initConfig. VrfProof must >= 128")
	}
	if len(configuration.VrfValue) < 128 {
		return fmt.Errorf("initConfig. VrfValue must >= 128")
	}

	indexMap := make(map[uint32]struct{})
	peerPubkeyMap := make(map[string]struct{})
	for _, peer := range configuration.Peers {
		_, ok := indexMap[peer.Index]
		if ok {
			return fmt.Errorf("initConfig, peer index is duplicated")
		}
		indexMap[peer.Index] = struct{}{}

		_, ok = peerPubkeyMap[peer.PeerPubkey]
		if ok {
			return fmt.Errorf("initConfig, peerPubkey is duplicated")
		}
		peerPubkeyMap[peer.PeerPubkey] = struct{}{}

		if peer.Index <= 0 {
			return fmt.Errorf("initConfig, peer index in config must > 0")
		}
		//check peerPubkey
		if err := utils.ValidatePeerPubKeyFormat(peer.PeerPubkey); err != nil {
			return fmt.Errorf("invalid peer pubkey")
		}
		_, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return fmt.Errorf("common.AddressFromBase58, address format error: %v", err)
		}
	}
	return nil
}

func GetConfig(native *native.NativeService) (*Configuration, error) {
	contract := utils.NodeManagerContractAddress
	config := new(Configuration)
	configBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return nil, fmt.Errorf("native.CacheDB.Get, get configBytes error: %v", err)
	}
	if configBytes == nil {
		return nil, fmt.Errorf("getConfig, configBytes is nil")
	}
	value, err := cstates.GetValueFromRawStorageItem(configBytes)
	if err != nil {
		return nil, fmt.Errorf("getConfig, deserialize from raw storage item err:%v", err)
	}
	if err := config.Deserialization(common.NewZeroCopySource(value)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize config error: %v", err)
	}
	return config, nil
}

func putConfig(native *native.NativeService, config *Configuration) {
	contract := utils.NodeManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	config.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(VBFT_CONFIG)), cstates.GenRawStorageItem(sink.Bytes()))
}

func getCandidateIndex(native *native.NativeService) (uint32, error) {
	contract := utils.NodeManagerContractAddress
	candidateIndexBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)))
	if err != nil {
		return 0, fmt.Errorf("native.CacheDB.Get, get candidateIndex error: %v", err)
	}
	if candidateIndexBytes == nil {
		return 0, fmt.Errorf("getCandidateIndex, candidateIndex is not init")
	} else {
		candidateIndexStore, err := cstates.GetValueFromRawStorageItem(candidateIndexBytes)
		if err != nil {
			return 0, fmt.Errorf("getCandidateIndex, deserialize from raw storage item err:%v", err)
		}
		candidateIndex := utils.GetBytesUint32(candidateIndexStore)
		return candidateIndex, nil
	}
}

func putCandidateIndex(native *native.NativeService, candidateIndex uint32) {
	contract := utils.NodeManagerContractAddress
	candidateIndexBytes := utils.GetUint32Bytes(candidateIndex)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), cstates.GenRawStorageItem(candidateIndexBytes))
}

func GetGovernanceView(native *native.NativeService) (*GovernanceView, error) {
	contract := utils.NodeManagerContractAddress
	governanceViewBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return nil, fmt.Errorf("getGovernanceView, get governanceViewBytes error: %v", err)
	}
	if governanceViewBytes != nil {
		governanceView := new(GovernanceView)
		value, err := cstates.GetValueFromRawStorageItem(governanceViewBytes)
		if err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize from raw storage item err:%v", err)
		}
		if err := governanceView.Deserialization(common.NewZeroCopySource(value)); err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize governanceView error: %v", err)
		}
		return governanceView, nil
	}
	return nil, fmt.Errorf("getGovernanceView, get nil governanceViewBytes")
}

func putGovernanceView(native *native.NativeService, governanceView *GovernanceView) {
	contract := utils.NodeManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	governanceView.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), cstates.GenRawStorageItem(sink.Bytes()))
}

func GetView(native *native.NativeService) (uint32, error) {
	governanceView, err := GetGovernanceView(native)
	if err != nil {
		return 0, fmt.Errorf("getView, getGovernanceView error: %v", err)
	}
	return governanceView.View, nil
}

func getConsensusSigns(native *native.NativeService, key common.Uint256) (*ConsensusSigns, error) {
	contract := utils.NodeManagerContractAddress
	consensusSignsStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(CONSENSUS_SIGNS), key.ToArray()))
	if err != nil {
		return nil, fmt.Errorf("GetConsensusSigns, get consensusSignsStore error: %v", err)
	}
	consensusSigns := &ConsensusSigns{
		SignsMap: make(map[common.Address]bool),
	}
	if consensusSignsStore != nil {
		consensusSignsBytes, err := cstates.GetValueFromRawStorageItem(consensusSignsStore)
		if err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize from raw storage item err:%v", err)
		}
		if err := consensusSigns.Deserialization(common.NewZeroCopySource(consensusSignsBytes)); err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize governanceView error: %v", err)
		}
	}
	return consensusSigns, nil
}

func putConsensusSigns(native *native.NativeService, key common.Uint256, consensusSigns *ConsensusSigns) {
	contract := utils.NodeManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	consensusSigns.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(CONSENSUS_SIGNS), key.ToArray()), cstates.GenRawStorageItem(sink.Bytes()))
}

func deleteConsensusSigns(native *native.NativeService, key common.Uint256) {
	contract := utils.NodeManagerContractAddress
	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(CONSENSUS_SIGNS), key.ToArray()))
}

func CheckConsensusSigns(native *native.NativeService, method string, input []byte, address common.Address) (bool, error) {
	message := append([]byte(method), input...)
	key := sha256.Sum256(message)
	consensusSigns, err := getConsensusSigns(native, key)
	if err != nil {
		return false, fmt.Errorf("CheckConsensusSigns, GetConsensusSigns error: %v", err)
	}
	consensusSigns.SignsMap[address] = true
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"CheckConsensusSigns", len(consensusSigns.SignsMap)},
		})
	//check signs num
	//get view
	view, err := GetView(native)
	if err != nil {
		return false, fmt.Errorf("CheckConsensusSigns, GetView error: %v", err)
	}
	//get consensus peer
	peerPoolMap, err := GetPeerPoolMap(native, view)
	if err != nil {
		return false, fmt.Errorf("CheckConsensusSigns, GetPeerPoolMap error: %v", err)
	}
	num := 0
	sum := 0
	for key, v := range peerPoolMap.PeerPoolMap {
		if v.Status == ConsensusStatus {
			k, err := hex.DecodeString(key)
			if err != nil {
				return false, fmt.Errorf("CheckConsensusSigns, hex.DecodeString public key error: %v", err)
			}
			publicKey, err := keypair.DeserializePublicKey(k)
			if err != nil {
				return false, fmt.Errorf("CheckConsensusSigns, keypair.DeserializePublicKey error: %v", err)
			}
			_, ok := consensusSigns.SignsMap[types.AddressFromPubKey(publicKey)]
			if ok {
				num = num + 1
			}
			sum = sum + 1
		}
	}
	if num >= (2*sum+2)/3 {
		deleteConsensusSigns(native, key)
		return true, nil
	} else {
		putConsensusSigns(native, key, consensusSigns)
		return false, nil
	}
}

// Get current epoch operator derived from current epoch consensus book keepers' public keys
func GetCurConOperator(native *native.NativeService) (common.Address, error) {
	view, err := GetView(native)
	if err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("GetCurConOperator, GetView error: %v", err)
	}
	//get consensus peer
	peerPoolMap, err := GetPeerPoolMap(native, view)
	if err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("GetCurConOperator, GetPeerPoolMap error: %v", err)
	}
	if peerPoolMap == nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("GetCurConOperator, GetPeerPoolMap empty peerPoolMap")
	}
	publicKeys := make([]keypair.PublicKey, 0)
	for key, v := range peerPoolMap.PeerPoolMap {
		if v.Status == ConsensusStatus {
			k, err := hex.DecodeString(key)
			if err != nil {
				return common.ADDRESS_EMPTY, fmt.Errorf("GetCurConOperator, hex.DecodeString public key error: %v", err)
			}
			publicKey, err := keypair.DeserializePublicKey(k)
			if err != nil {
				return common.ADDRESS_EMPTY, fmt.Errorf("GetCurConOperator, keypair.DeserializePublicKey error: %v", err)
			}
			publicKeys = append(publicKeys, publicKey)
		}
	}
	operator, err := types.AddressFromBookkeepers(publicKeys)
	if err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("GetCurConOperator, AddressFromBookkeepers error: %v", err)
	}
	return operator, nil
}
