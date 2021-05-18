package neo3_state_manager

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
)

//// putStateValidator put state validator in the contract storage using the STATE_VALIDATOR key
//func putStateValidator(native *native.NativeService, stateValidator helper.UInt160) error {
//	contract := utils.Neo3StateManagerContractAddress
//	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR), stateValidator.ToByteArray()), cstates.GenRawStorageItem(stateValidator.ToByteArray()))
//	return nil
//}

//func getStateValidator(native *native.NativeService, stateValidator helper.UInt160) error {
//	contract := utils.Neo3StateManagerContractAddress
//	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR), stateValidator.ToByteArray()), cstates.GenRawStorageItem(stateValidator.ToByteArray()))
//	return nil
//}

//
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
	source := common.NewZeroCopySource(data)
	n, eof := source.NextVarUint()
	if eof {
		return nil, fmt.Errorf("source.NextVarUint error")
	}
	result := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		ss, eof := source.NextString()
		if eof {
			return nil, fmt.Errorf("source.NextString error")
		}
		result = append(result, ss)
	}
	return result, nil
}

func getStateValidators(native *native.NativeService) ([]string, error) {
	contract := utils.Neo3StateManagerContractAddress
	svStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)))
	if err != nil {
		return nil, fmt.Errorf("getStateValidator, get stateValidatorListParamStore error: %v", err)
	}
	if svStore == nil {
		return []string{}, nil
	}
	svBytes, err := cstates.GetValueFromRawStorageItem(svStore)
	if err != nil {
		return nil, fmt.Errorf("getStateValidator, deserialize from raw storage item error: %v", err)
	}
	svs, err := DeserializeStringArray(svBytes)
	if err != nil {
		return nil, fmt.Errorf("getStateValidator, convert to UInt160 array error: %v", err)
	}
	return svs, nil
}

func putStateValidators(native *native.NativeService, stateValidators []string) error {
	contract := utils.Neo3StateManagerContractAddress
	// get current stored value
	oldSVs, err := getStateValidators(native)
	if err != nil {
		return fmt.Errorf("putStateValidator, get old state validators error: %v", err)
	}
	// use a map to filter old svs
	mm := make(map[string]string)
	for _, v := range oldSVs {
		if _, ok := mm[v]; ok {
			continue
		}
		mm[v] = v
	}
	// use the map to add new svs
	for _, v := range stateValidators {
		if _, ok := mm[v]; ok {
			continue
		}
		mm[v] = v
	}
	// convert map back to string array
	newSVs := make([]string, 0)
	for _, v := range mm {
		newSVs = append(newSVs, v)
	}
	// convert back to []byte
	data := SerializeStringArray(newSVs)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)), cstates.GenRawStorageItem(data))
	return nil
}

func removeStateValidators(native *native.NativeService, stateValidators []string) error {
	contract := utils.Neo3StateManagerContractAddress
	// get current stored value
	oldSVs, err := getStateValidators(native)
	if err != nil {
		return fmt.Errorf("removeStateValidator, get old state validators error: %v", err)
	}
	// use a map to filter old svs
	mm := make(map[string]string)
	for _, v := range oldSVs {
		if _, ok := mm[v]; ok {
			continue
		}
		mm[v] = v
	}
	// use the map to delete svs
	for _, v := range stateValidators {
		// if the sv is in map, delete it, else continue
		if _, ok := mm[v]; ok {
			delete(mm, v)
		}
	}
	// convert map back to UInt160 array
	newSVs := make([]string, 0)
	for _, v := range mm {
		newSVs = append(newSVs, v)
	}
	// if no sv left, delete the storage, else put remaining back
	if len(newSVs) == 0 {
		native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)))
		return nil
	}

	// convert back to []byte
	data := SerializeStringArray(newSVs)
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
		return nil, fmt.Errorf("getStateValidatorApply, can't find any record")
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
		return nil, fmt.Errorf("getStateValidatorRemove, can't find any record")
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
