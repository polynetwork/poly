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
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

func GetPeeApply(native *native.NativeService, peerPubkey string) (*RegisterPeerParam, error) {
	contract := utils.NodeManagerContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("GetPeeApply, peerPubkey format error: %v", err)
	}
	peerBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(PEER_APPLY), peerPubkeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("GetPeeApply, get peer error: %v", err)
	}
	if peerBytes == nil {
		return nil, nil
	}
	peerStore, err := cstates.GetValueFromRawStorageItem(peerBytes)
	if err != nil {
		return nil, fmt.Errorf("GetPeeApply, deserialize from raw storage item err:%v", err)
	}
	peer := new(RegisterPeerParam)
	if err := peer.Deserialization(common.NewZeroCopySource(peerStore)); err != nil {
		return nil, fmt.Errorf("GetPeeApply, deserialize peer error: %v", err)
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

func putPeerPoolMap(native *native.NativeService, peerPoolMap *PeerPoolMap, view uint32) error {
	contract := utils.NodeManagerContractAddress
	viewBytes := utils.GetUint32Bytes(view)
	sink := common.NewZeroCopySink(nil)
	peerPoolMap.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
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

func putConfig(native *native.NativeService, config *Configuration) error {
	contract := utils.NodeManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	config.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(VBFT_CONFIG)), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
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

func putCandidateIndex(native *native.NativeService, candidateIndex uint32) error {
	contract := utils.NodeManagerContractAddress
	candidateIndexBytes := utils.GetUint32Bytes(candidateIndex)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), cstates.GenRawStorageItem(candidateIndexBytes))
	return nil
}

func GetGovernanceView(native *native.NativeService) (*GovernanceView, error) {
	contract := utils.NodeManagerContractAddress
	governanceViewBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return nil, fmt.Errorf("getGovernanceView, get governanceViewBytes error: %v", err)
	}
	governanceView := new(GovernanceView)
	if governanceViewBytes == nil {
		return nil, fmt.Errorf("getGovernanceView, get nil governanceViewBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(governanceViewBytes)
		if err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize from raw storage item err:%v", err)
		}
		if err := governanceView.Deserialization(common.NewZeroCopySource(value)); err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize governanceView error: %v", err)
		}
	}
	return governanceView, nil
}

func putGovernanceView(native *native.NativeService, governanceView *GovernanceView) error {
	contract := utils.NodeManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	governanceView.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetView(native *native.NativeService) (uint32, error) {
	governanceView, err := GetGovernanceView(native)
	if err != nil {
		return 0, fmt.Errorf("getView, getGovernanceView error: %v", err)
	}
	return governanceView.View, nil
}
