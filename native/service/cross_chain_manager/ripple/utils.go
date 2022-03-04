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

package ripple

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	crosscommon "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutMultisignInfo(native *native.NativeService, id []byte, multisignInfo *MultisignInfo) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.MULTISIGN_INFO), id)
	sink := common.NewZeroCopySink(nil)
	multisignInfo.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func GetMultisignInfo(native *native.NativeService, id []byte) (*MultisignInfo, error) {
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(crosscommon.MULTISIGN_INFO), id)
	store, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("GetMultisignInfo, get multisign info store error: %v", err)
	}
	multisignInfo := &MultisignInfo{
		SigMap: make(map[string]bool),
	}
	if store != nil {
		multisignInfoBytes, err := cstates.GetValueFromRawStorageItem(store)
		if err != nil {
			return nil, fmt.Errorf("GetMultisignInfo, deserialize from raw storage item err:%v", err)
		}
		err = multisignInfo.Deserialization(common.NewZeroCopySource(multisignInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("GetMultisignInfo, deserialize multisign info err:%v", err)
		}
	}
	return multisignInfo, nil
}
