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
package chainsql

import (
	"fmt"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutChainsqlLatestHeightInProcessing(ns *native.NativeService, chainID uint64, fromContract []byte, height uint32) {
	last, _ := GetChainsqlLatestHeightInProcessing(ns, chainID, fromContract)
	if height <= last {
		return
	}
	ns.GetCacheDB().Put(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(common.LATEST_HEIGHT_IN_PROCESSING), utils.GetUint64Bytes(chainID), fromContract),
		utils.GetUint32Bytes(height))
}

func GetChainsqlLatestHeightInProcessing(ns *native.NativeService, chainID uint64, fromContract []byte) (uint32, error) {
	store, err := ns.GetCacheDB().Get(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(common.LATEST_HEIGHT_IN_PROCESSING), utils.GetUint64Bytes(chainID), fromContract))
	if err != nil {
		return 0, fmt.Errorf("GetChainsqlRoot, get root error: %v", err)
	}
	if store == nil {
		return 0, fmt.Errorf("GetChainsqlRoot, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return 0, fmt.Errorf("GetChainsqlRoot, deserialize from raw storage item err: %v", err)
	}
	return utils.GetBytesUint32(raw), nil
}
