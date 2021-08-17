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

package relayer_manager

import (
	"fmt"
	"github.com/polynetwork/poly/native/event"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//function name
	REGISTER_RELAYER         = "registerRelayer"
	APPROVE_REGISTER_RELAYER = "approveRegisterRelayer"
	REMOVE_RELAYER           = "RemoveRelayer"
	APPROVE_REMOVE_RELAYER   = "approveRemoveRelayer"

	//key prefix
	RELAYER        = "relayer"
	RELAYER_APPLY  = "relayerApply"
	RELAYER_REMOVE = "relayerRemove"
	APPLY_ID       = "applyID"
	REMOVE_ID      = "removeID"
)

//Register methods of node_manager contract
func RegisterRelayerManagerContract(native *native.NativeService) {
	native.Register(REGISTER_RELAYER, RegisterRelayer)
	native.Register(APPROVE_REGISTER_RELAYER, ApproveRegisterRelayer)
	native.Register(REMOVE_RELAYER, RemoveRelayer)
	native.Register(APPROVE_REMOVE_RELAYER, ApproveRemoveRelayer)
}

func RegisterRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, contract params deserialize error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, params.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, checkWitness: %s, error: %v", params.Address.ToBase58(), err)
	}
	if err := putRelayerApply(native, params); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, putRelayer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveRegisterRelayer(native *native.NativeService) ([]byte, error) {
	params := new(ApproveRelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterRelayer, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterRelayer, checkWitness error: %v", err)
	}

	relayerListParam, err := getRelayerApply(native, params.ID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterRelayer, getRelayerApply error: %v", err)
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_REGISTER_RELAYER, utils.GetUint64Bytes(params.ID), params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterRelayer, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	for _, address := range relayerListParam.AddressList {
		err = putRelayer(native, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterRelayer, putRelayer error: %v", err)
		}
	}
	native.GetCacheDB().Delete(utils.ConcatKey(utils.RelayerManagerContractAddress, []byte(RELAYER_APPLY), utils.GetUint64Bytes(params.ID)))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.RelayerManagerContractAddress,
			States:          []interface{}{"ApproveRegisterRelayer", params.ID},
		})
	return utils.BYTE_TRUE, nil
}

func RemoveRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, contract params deserialize error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, params.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, checkWitness: %s, error: %v", params.Address.ToBase58(), err)
	}
	err := putRelayerRemove(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, putRelayer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveRemoveRelayer(native *native.NativeService) ([]byte, error) {
	params := new(ApproveRelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveRelayer, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveRelayer, checkWitness error: %v", err)
	}

	relayerListParam, err := getRelayerRemove(native, params.ID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveRelayer, getRelayerRemove error: %v", err)
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_REMOVE_RELAYER, utils.GetUint64Bytes(params.ID), params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRemoveRelayer, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	for _, address := range relayerListParam.AddressList {
		native.GetCacheDB().Delete(utils.ConcatKey(utils.RelayerManagerContractAddress, []byte(RELAYER), address[:]))
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.RelayerManagerContractAddress,
			States:          []interface{}{"ApproveRemoveRelayer", params.ID},
		})
	return utils.BYTE_TRUE, nil
}
