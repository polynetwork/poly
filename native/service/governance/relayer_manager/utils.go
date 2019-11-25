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

package relayer_manager

import (
	"fmt"

	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

func putRelayer(native *native.NativeService, relayer *RelayerParam) error {
	contract := utils.RelayerManagerContractAddress

	sink := common.NewZeroCopySink(nil)
	relayer.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(RELAYER), relayer.Address), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetRelayer(native *native.NativeService, address []byte) (*RelayerParam, error) {
	contract := utils.RelayerManagerContractAddress

	relayerBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(RELAYER), address))
	if err != nil {
		return nil, fmt.Errorf("GetRelayer, get relayerBytes error: %v", err)
	}
	if relayerBytes == nil {
		return nil, nil
	}
	relayerStore, err := cstates.GetValueFromRawStorageItem(relayerBytes)
	if err != nil {
		return nil, fmt.Errorf("GetRelayer, deserialize from raw storage item err:%v", err)
	}
	relayer := new(RelayerParam)
	if err := relayer.Deserialization(common.NewZeroCopySource(relayerStore)); err != nil {
		return nil, fmt.Errorf("GetRelayer, deserialize relayer error: %v", err)
	}
	return relayer, nil
}

func GetRelayerRaw(native *native.NativeService, address []byte) ([]byte, error) {
	contract := utils.RelayerManagerContractAddress

	relayerBytes, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(RELAYER), address))
	if err != nil {
		return nil, fmt.Errorf("GetRelayerRaw, get relayerBytes error: %v", err)
	}
	if relayerBytes == nil {
		return nil, nil
	}
	return relayerBytes, nil
}
