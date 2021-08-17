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
package cross_chain_manager

import (
	"encoding/json"
	"fmt"
	"strings"

	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutBlackChain(native *native.NativeService, chainID uint64) {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(BLACKED_CHAIN), chainIDBytes),
		cstates.GenRawStorageItem(chainIDBytes))
}

func CheckIfChainBlacked(native *native.NativeService, chainID uint64) (bool, error) {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	chainIDStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(BLACKED_CHAIN), chainIDBytes))
	if err != nil {
		return true, fmt.Errorf("CheckBlackChain, get black chainIDStore error: %v", err)
	}
	if chainIDStore == nil {
		return false, nil
	}
	return true, nil
}

func RemoveBlackChain(native *native.NativeService, chainID uint64) {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(BLACKED_CHAIN), chainIDBytes))
}

// lock proxy address white list
func PutWhiteAddress(native *native.NativeService, addresses []string) error {
	contract := utils.CrossChainManagerContractAddress
	addressBytes, err := json.Marshal(addresses)
	if err != nil {
		return fmt.Errorf("PutWhiteAddress, addresses marshal error: %v", err)
	}

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(WHITE_ADDRESS)),
		cstates.GenRawStorageItem(addressBytes))
	return nil
}

func CheckIfAddressWhite(native *native.NativeService, address string) (bool, []string, error) {
	contract := utils.CrossChainManagerContractAddress
	store, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(WHITE_ADDRESS)))
	if err != nil {
		return false, nil, fmt.Errorf("CheckIfAddressWhite, get data error: %v", err)
	}
	if store == nil {
		return false, nil, fmt.Errorf("CheckIfAddressWhite, store is nil")
	}
	value, err := cstates.GetValueFromRawStorageItem(store)
	if err != nil {
		return false, nil, fmt.Errorf("CheckIfAddressWhite, GetValueFromRawStorageItem err: %v", err)
	}
	var addJson []string
	err = json.Unmarshal(value, &addJson)
	if err != nil {
		return false, nil, fmt.Errorf("CheckIfAddressWhite, Unmarshal err: %v", err)
	}
	for _,v := range addJson {
		if strings.EqualFold(v, address) {
			return true, addJson, nil
		}
	}
	
	return false, addJson, nil
}