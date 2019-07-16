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

package side_chain_manager

import (
	"fmt"

	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

func GetMainChain(native *native.NativeService) (uint64, error) {
	contract := utils.ChainManagerContractAddress
	mainChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(MAIN_CHAIN)))
	if err != nil {
		return 0, fmt.Errorf("get mainChainStore error: %v", err)
	}
	if mainChainStore == nil {
		return 0, fmt.Errorf("GetMainChain, can not find any record")
	}
	mainChainBytes, err := cstates.GetValueFromRawStorageItem(mainChainStore)
	if err != nil {
		return 0, fmt.Errorf("GetMainChain, deserialize from raw storage item err:%v", err)
	}
	mainChainID, err := utils.GetBytesUint64(mainChainBytes)
	if err != nil {
		return 0, fmt.Errorf("GetMainChain, utils.GetBytesUint64 err:%v", err)
	}
	return mainChainID, nil
}

func putRegisterSideChainRequest(native *native.NativeService, chainID uint64) error {
	contract := utils.ChainManagerContractAddress
	mainChainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(MAIN_CHAIN)), cstates.GenRawStorageItem(mainChainIDBytes))
	return nil
}
