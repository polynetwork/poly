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
package fisco

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutFiscoRoot(native *native.NativeService, root *FiscoRoot, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	sink := pcom.NewZeroCopySink(nil)
	root.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(common.ROOT_CERT), utils.GetUint64Bytes(chainID)),
		states.GenRawStorageItem(sink.Bytes()))

	common.NotifyPutCertificate(native, chainID, root.RootCA.Raw)
	return nil
}

func GetFiscoRoot(native *native.NativeService, chainID uint64) (*FiscoRoot, error) {
	store, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(common.ROOT_CERT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return nil, fmt.Errorf("GetFiscoRoot, get root error: %v", err)
	}
	if store == nil {
		return nil, fmt.Errorf("GetFiscoRoot, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return nil, fmt.Errorf("GetFiscoRoot, deserialize from raw storage item err: %v", err)
	}
	root := &FiscoRoot{}
	if err = root.Deserialization(pcom.NewZeroCopySource(raw)); err != nil {
		return nil, fmt.Errorf("GetFiscoRoot, failed to deserialize FiscoRoot: %v", err)
	}
	return root, nil
}
