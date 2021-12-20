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

package neo3_state_manager

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
)

func SerializeStringArray(data []string) []byte {
	sink := common.NewZeroCopySink(nil)
	// serialize
	sink.WriteVarUint(uint64(len(data)))
	for _, v := range data {
		sink.WriteString(v)
	}
	return sink.Bytes()
}

func DeserializeStringArray(data []byte) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	source := common.NewZeroCopySource(data)
	n, eof := source.NextVarUint()
	if eof {
		return nil, fmt.Errorf("source.NextVarUint error")
	}
	result := make([]string, 0, n)
	for i := 0; uint64(i) < n; i++ {
		ss, eof := source.NextString()
		if eof {
			return nil, fmt.Errorf("source.NextString error")
		}
		result = append(result, ss)
	}
	return result, nil
}

func getStateValidators(native *native.NativeService) ([]byte, error) {
	contract := utils.Neo3StateManagerContractAddress
	svStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)))
	if err != nil {
		return nil, fmt.Errorf("getStateValidator, get stateValidatorListParamStore error: %v", err)
	}
	if svStore == nil {
		return []byte{}, nil
	}
	svBytes, err := cstates.GetValueFromRawStorageItem(svStore)
	if err != nil {
		return nil, fmt.Errorf("getStateValidator, deserialize from raw storage item error: %v", err)
	}
	return svBytes, nil
}

func putStateValidators(native *native.NativeService, stateValidators []string) error {
	contract := utils.Neo3StateManagerContractAddress
	// get current stored value
	oldSvBytes, err := getStateValidators(native)
	if err != nil {
		return fmt.Errorf("putStateValidator, get old state validators error: %v", err)
	}
	oldSVs, err := DeserializeStringArray(oldSvBytes)
	if err != nil {
		return fmt.Errorf("putStateValidator, convert to string array error: %v", err)
	}
	// max capacity = len(oldSVs)+len(stateValidators)
	newSVs := make([]string, 0, len(oldSVs)+len(stateValidators))
	newSVs = append(newSVs, oldSVs...)
	// filter duplicate svs
	for _, sv := range stateValidators {
		isInOld := false
		for _, oldSv := range oldSVs {
			if sv == oldSv {
				isInOld = true
				break
			}
		}
		if !isInOld {
			newSVs = append(newSVs, sv)
		}
	}
	// convert back to []byte
	data := SerializeStringArray(newSVs)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)), cstates.GenRawStorageItem(data))
	return nil
}

func removeStateValidators(native *native.NativeService, stateValidators []string) error {
	contract := utils.Neo3StateManagerContractAddress
	// get current stored value
	oldSvBytes, err := getStateValidators(native)
	if err != nil {
		return fmt.Errorf("removeStateValidator, get old state validators error: %v", err)
	}
	oldSVs, err := DeserializeStringArray(oldSvBytes)
	if err != nil {
		return fmt.Errorf("removeStateValidator, convert to string array error: %v", err)
	}
	// remove in the slice
	for _, sv := range stateValidators {
		for i, oldSv := range oldSVs {
			if sv == oldSv {
				oldSVs = append(oldSVs[:i], oldSVs[i+1:]...)
				break
			}
		}
	}
	// if no sv left, delete the storage, else put remaining back
	if len(oldSVs) == 0 {
		native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)))
		return nil
	}
	// convert back to []byte
	data := SerializeStringArray(oldSVs)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)), cstates.GenRawStorageItem(data))
	return nil
}

func getStateValidatorApply(native *native.NativeService, applyID uint64) (*StateValidatorListParam, error) {
	contract := utils.Neo3StateManagerContractAddress
	svListParamStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_APPLY), utils.GetUint64Bytes(applyID)))
	if err != nil {
		return nil, fmt.Errorf("getStateValidatorApply, get stateValidatorListParamStore error: %v", err)
	}
	if svListParamStore == nil {
		return nil, nil
	}
	svListParam := new(StateValidatorListParam)
	svListParamBytes, err := cstates.GetValueFromRawStorageItem(svListParamStore)
	if err != nil {
		return nil, fmt.Errorf("getStateValidatorApply, deserialize from raw storage item error: %v", err)
	}
	err = svListParam.Deserialization(common.NewZeroCopySource(svListParamBytes))
	if err != nil {

	}
	return svListParam, nil
}

func putStateValidatorApply(native *native.NativeService, stateValidatorListParam *StateValidatorListParam) error {
	contract := utils.Neo3StateManagerContractAddress
	applyID, err := getStateValidatorApplyID(native)
	if err != nil {
		return fmt.Errorf("putStateValidatorApply, getStateValidatorApplyID error: %v", err)
	}
	newApplyID := applyID + 1
	err = putStateValidatorApplyID(native, newApplyID)
	if err != nil {
		return fmt.Errorf("putStateValidatorApply, putStateValidatorApplyID error: %v", err)
	}
	sink := common.NewZeroCopySink(nil)
	stateValidatorListParam.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_APPLY), utils.GetUint64Bytes(applyID)), cstates.GenRawStorageItem(sink.Bytes()))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{"putStateValidatorApply", applyID},
		})
	return nil
}

func getStateValidatorApplyID(native *native.NativeService) (uint64, error) {
	contract := utils.Neo3StateManagerContractAddress
	applyIDStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_APPLY_ID)))
	if err != nil {
		return 0, fmt.Errorf("getStateValidatorApplyID, get applyIDStore error: %v", err)
	}
	var applyID uint64 = 0
	if applyIDStore != nil {
		applyIDBytes, err := cstates.GetValueFromRawStorageItem(applyIDStore)
		if err != nil {
			return 0, fmt.Errorf("getStateValidatorApplyID, deserialize from raw storage item error: %v", err)
		}
		applyID = utils.GetBytesUint64(applyIDBytes)
	}
	return applyID, nil
}

func putStateValidatorApplyID(native *native.NativeService, applyID uint64) error {
	contract := utils.Neo3StateManagerContractAddress
	applyIDByte := utils.GetUint64Bytes(applyID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_APPLY_ID)), cstates.GenRawStorageItem(applyIDByte))
	return nil
}

func getStateValidatorRemove(native *native.NativeService, removeID uint64) (*StateValidatorListParam, error) {
	contract := utils.Neo3StateManagerContractAddress
	svListParamStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_REMOVE), utils.GetUint64Bytes(removeID)))
	if err != nil {
		return nil, fmt.Errorf("getStateValidatorRemove, get stateValidatorListParamStore error: %v", err)
	}
	if svListParamStore == nil {
		return nil, nil
	}
	svListParam := new(StateValidatorListParam)
	svListParamBytes, err := cstates.GetValueFromRawStorageItem(svListParamStore)
	if err != nil {
		return nil, fmt.Errorf("getStateValidatorRemove, deserialize from raw storage item error: %v", err)
	}
	err = svListParam.Deserialization(common.NewZeroCopySource(svListParamBytes))
	if err != nil {
		return nil, fmt.Errorf("getStateValidatorRemove, svListParam.Deserialization error: %v", err)
	}
	return svListParam, nil
}

func putStateValidatorRemove(native *native.NativeService, svListParam *StateValidatorListParam) error {
	contract := utils.Neo3StateManagerContractAddress
	removeID, err := getStateValidatorRemoveID(native)
	if err != nil {
		return fmt.Errorf("putStateValidatorRemove, getStateValidatorRemoveID error: %v", err)
	}
	newRemoveID := removeID + 1
	err = putStateValidatorRemoveID(native, newRemoveID)
	if err != nil {
		return fmt.Errorf("putStateValidatorRemove, putStateValidatorRemoveID error: %v", err)
	}
	sink := common.NewZeroCopySink(nil)
	svListParam.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_REMOVE), utils.GetUint64Bytes(removeID)), cstates.GenRawStorageItem(sink.Bytes()))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{"putStateValidatorRemove", removeID},
		})
	return nil
}

func getStateValidatorRemoveID(native *native.NativeService) (uint64, error) {
	contract := utils.Neo3StateManagerContractAddress
	removeIDStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_REMOVE_ID)))
	if err != nil {
		return 0, fmt.Errorf("getStateValidatorRemoveID, get removeIDStore error: %v", err)
	}
	var removeID uint64 = 0
	if removeIDStore != nil {
		removeIDBytes, err := cstates.GetValueFromRawStorageItem(removeIDStore)
		if err != nil {
			return 0, fmt.Errorf("getStateValidatorRemoveID, deserialize from raw storage item error: %v", err)
		}
		removeID = utils.GetBytesUint64(removeIDBytes)
	}
	return removeID, nil
}

func putStateValidatorRemoveID(native *native.NativeService, removeID uint64) error {
	contract := utils.Neo3StateManagerContractAddress
	removeIDBytes := utils.GetUint64Bytes(removeID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR_REMOVE_ID)), cstates.GenRawStorageItem(removeIDBytes))
	return nil
}
