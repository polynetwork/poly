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

package governance

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

func registerCandidate(native *native.NativeService) error {
	params := new(RegisterPeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	contract := utils.GovernanceContractAddress

	address, err := common.AddressParseFromBytes(params.Address)
	if err != nil {
		return fmt.Errorf("registerCandidate, common.AddressParseFromBytes error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(params.PeerPubkey); err != nil {
		return fmt.Errorf("invalid peer pubkey")
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	//get black list
	blackList, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return fmt.Errorf("native.CacheDB.Get, get BlackList error: %v", err)
	}
	if blackList != nil {
		return fmt.Errorf("registerCandidate, this Peer is in BlackList")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return fmt.Errorf("registerCandidate, peerPubkey is already in peerPoolMap")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		Status:     RegisterCandidateStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return nil
}
