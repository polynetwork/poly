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
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/governance/node_manager"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	//function name
	REGISTER_RELAYER = "registerRelayer"
	REMOVE_RELAYER   = "RemoveRelayer"

	//key prefix
	RELAYER = "relayer"
)

//Register methods of node_manager contract
func RegisterRelayerManagerContract(native *native.NativeService) {
	native.Register(REGISTER_RELAYER, RegisterRelayer)
	native.Register(REMOVE_RELAYER, RemoveRelayer)
}

func RegisterRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, checkWitness error: %v", err)
	}
	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, REGISTER_RELAYER, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	for _, address := range params.AddressList {
		err = putRelayer(native, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, putRelayer error: %v", err)
		}
	}

	return utils.BYTE_TRUE, nil
}

func RemoveRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, contract params deserialize error: %v", err)
	}
	contract := utils.RelayerManagerContractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, checkWitness error: %v", err)
	}
	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, REMOVE_RELAYER, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	for _, address := range params.AddressList {
		//get relayer
		relayerRaw, err := GetRelayerRaw(native, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, get relayer error: %v", err)
		}
		if relayerRaw == nil {
			return utils.BYTE_FALSE, fmt.Errorf("RemoveRelayer, relayer is not registered")
		}

		native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(RELAYER), address[:]))
	}

	return utils.BYTE_TRUE, nil
}
