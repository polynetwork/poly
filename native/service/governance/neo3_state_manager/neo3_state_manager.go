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

package neo3_state_manager

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//function name
	GET_CURRENT_STATE_VALIDATOR      = "getCurrentStateValidator"
	REGISTER_STATE_VALIDATOR         = "registerStateValidator"
	APPROVE_REGISTER_STATE_VALIDATOR = "approveRegisterStateValidator"
	REMOVE_STATE_VALIDATOR           = "removeStateValidator"
	APPROVE_REMOVE_STATE_VALIDATOR   = "approveRemoveStateValidator"

	//key prefix
	STATE_VALIDATOR           = "stateValidator"
	STATE_VALIDATOR_APPLY     = "stateValidatorApply"
	STATE_VALIDATOR_REMOVE    = "stateValidatorRemove"
	STATE_VALIDATOR_APPLY_ID  = "stateValidatorApplyID"
	STATE_VALIDATOR_REMOVE_ID = "stateValidatorRemoveID"
)

//Register methods of node_manager contract
func RegisterStateValidatorManagerContract(native *native.NativeService) {
	native.Register(GET_CURRENT_STATE_VALIDATOR, GetCurrentStateValidator)
	native.Register(REGISTER_STATE_VALIDATOR, RegisterStateValidator)
	native.Register(APPROVE_REGISTER_STATE_VALIDATOR, ApproveRegisterStateValidator)
	native.Register(REMOVE_STATE_VALIDATOR, RemoveStateValidator)
	native.Register(APPROVE_REMOVE_STATE_VALIDATOR, ApproveRemoveStateValidator)
}

func GetCurrentStateValidator(native *native.NativeService) ([]byte, error) {
	data, err := getStateValidators(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetCurrentStateValidator, getStateValidators error: %v", err)
	}
	return data, nil
}

func RegisterStateValidator(native *native.NativeService) ([]byte, error) {
	params := new(StateValidatorListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterStateValidator, params.Deserialization error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, params.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterStateValidator, checkWitness: %s, error: %v", params.Address.ToBase58(), err)
	}
	if err := putStateValidatorApply(native, params); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterStateValidator, putStateValidatorApply error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveRegisterStateValidator(native *native.NativeService) ([]byte, error) {
	params := new(ApproveStateValidatorParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterStateValidator, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterStateValidator, checkWitness error: %v", err)
	}

	svListParam, err := getStateValidatorApply(native, params.ID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterStateValidator, getStateValidatorApply error: %v", err)
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_REGISTER_STATE_VALIDATOR, utils.GetUint64Bytes(params.ID), params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterStateValidator, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_FALSE, nil
	}

	// put all the state validators in storage
	err = putStateValidators(native, svListParam.StateValidators)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterStateValidator, putStateValidators error: %v", err)
	}

	native.GetCacheDB().Delete(utils.ConcatKey(utils.Neo3StateManagerContractAddress, []byte(STATE_VALIDATOR_APPLY), utils.GetUint64Bytes(params.ID)))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.Neo3StateManagerContractAddress,
			States:          []interface{}{"ApproveRegisterStateValidator", params.ID},
		})
	return utils.BYTE_TRUE, nil
}

func RemoveStateValidator(native *native.NativeService) ([]byte, error) {
	params := new(StateValidatorListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveStateValidator, contract params deserialize error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, params.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveStateValidator, checkWitness: %s, error: %v", params.Address.ToBase58(), err)
	}
	err := putStateValidatorRemove(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveStateValidator, putStateValidatorRemove error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveRemoveStateValidator(native *native.NativeService) ([]byte, error) {
	params := new(ApproveStateValidatorParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveStateValidator, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveStateValidator, checkWitness error: %v", err)
	}

	svListParam, err := getStateValidatorRemove(native, params.ID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveStateValidator, getStateValidatorRemove error: %v", err)
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_REMOVE_STATE_VALIDATOR, utils.GetUint64Bytes(params.ID), params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveStateValidator, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_FALSE, nil
	}

	// remove svs
	err = removeStateValidators(native, svListParam.StateValidators)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveStateValidator, removeStateValidators error: %v", err)
	}

	native.GetCacheDB().Delete(utils.ConcatKey(utils.Neo3StateManagerContractAddress, []byte(STATE_VALIDATOR_REMOVE), utils.GetUint64Bytes(params.ID)))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.Neo3StateManagerContractAddress,
			States:          []interface{}{"ApproveRemoveStateValidator", params.ID},
		})
	return utils.BYTE_TRUE, nil
}
