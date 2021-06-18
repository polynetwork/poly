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
package quorum

import (
	"fmt"
	ecom "github.com/ethereum/go-ethereum/common"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

var (
	IstanbulExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
	IstanbulDigest      = ecom.HexToHash("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365")
)

func putValSet(ns *native.NativeService, chainID, height uint64, vals []ecom.Address) {
	vs := QuorumValSet(vals)
	sink := pcom.NewZeroCopySink(nil)
	vs.Serialize(sink)

	rawChainID := utils.GetUint64Bytes(chainID)
	rawHeight := utils.GetUint64Bytes(height)
	ns.GetCacheDB().Put(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(common.CONSENSUS_PEER), rawChainID), states.GenRawStorageItem(sink.Bytes()))
	ns.GetCacheDB().Put(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(common.CONSENSUS_PEER_BLOCK_HEIGHT), rawChainID),
		states.GenRawStorageItem(rawHeight))
}

func GetValSet(ns *native.NativeService, chainID uint64) (QuorumValSet, error) {
	rawChainID := utils.GetUint64Bytes(chainID)
	store, err := ns.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(common.CONSENSUS_PEER), rawChainID))
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, fmt.Errorf("GetValSet, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return nil, fmt.Errorf("GetValSet, deserialize from raw storage item err: %v", err)
	}
	vs := QuorumValSet(make([]ecom.Address, 0))
	if err = vs.Deserialize(pcom.NewZeroCopySource(raw)); err != nil {
		return nil, err
	}
	return vs, nil
}

func GetCurrentValHeight(ns *native.NativeService, chainID uint64) (uint64, error) {
	rawChainID := utils.GetUint64Bytes(chainID)
	store, err := ns.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(common.CONSENSUS_PEER_BLOCK_HEIGHT), rawChainID))
	if err != nil {
		return 0, err
	}
	if store == nil {
		return 0, fmt.Errorf("getCurrentValHeight, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return 0, fmt.Errorf("getCurrentValHeight, deserialize from raw storage item err: %v", err)
	}

	return utils.GetBytesUint64(raw), nil
}

func GetSigners(hash ecom.Hash, sealArr [][]byte) ([]ecom.Address, error) {
	proposalSeal := PrepareCommittedSeal(hash)
	addrs := make([]ecom.Address, 0, len(sealArr))
	for _, seal := range sealArr {
		addr, err := GetSignatureAddress(proposalSeal, seal)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
