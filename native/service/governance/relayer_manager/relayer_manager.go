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
	"bytes"
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	//function name
	REGISTER_RELAYER   = "registerRelayer"
	UNREGISTER_RELAYER = "unRegisterRelayer"
	APPROVE_RELAYER    = "approveRelayer"
	REJECT_RELAYER     = "rejectRelayer"
	BLACK_RELAYER      = "blackRelayer"
	WHITE_RELAYER      = "whiteRelayer"
	QUIT_RELAYER       = "quitRelayer"

	//key prefix
	RELAYER_APPLY = "relayerApply"
	RELAYER       = "relayer"
	RELAYER_BLACK = "relayerBlack"
)

//Register methods of node_manager contract
func RegisterRelayerManagerContract(native *native.NativeService) {
	native.Register(REGISTER_RELAYER, RegisterRelayer)
	native.Register(UNREGISTER_RELAYER, UnRegisterRelayer)
	native.Register(QUIT_RELAYER, QuitRelayer)
	native.Register(APPROVE_RELAYER, ApproveRelayer)
	native.Register(REJECT_RELAYER, RejectRelayer)
	native.Register(BLACK_RELAYER, BlackRelayer)
	native.Register(WHITE_RELAYER, WhiteRelayer)
}

//Register a candidate node, used by users.
func RegisterRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, contract params deserialize error: %v", err)
	}

	address, err := common.AddressParseFromBytes(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, common.AddressParseFromBytes error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, checkWitness error: %v", err)
	}

	relayerApply, err := GetRelayerApply(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, get relayer error: %v", err)
	}
	if relayerApply != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, relayer already applied")
	}
	relayer, err := GetRelayer(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, get relayer error: %v", err)
	}
	if relayer != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, relayer already registered")
	}
	blacked, err := checkIfBlacked(native, params.Address)
	if blacked {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, relayer is blacked")
	}

	err = putRelayerApply(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRelayer, put putRelayerApply error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func UnRegisterRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnRegisterRelayer, contract params deserialize error: %v", err)
	}
	address, err := common.AddressParseFromBytes(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnRegisterRelayer, common.AddressParseFromBytes error: %v", err)
	}
	contract := utils.RelayerManagerContractAddress

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnRegisterRelayer, checkWitness error: %v", err)
	}

	//get relayer apply
	relayer, err := GetRelayerApply(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnRegisterRelayer, get relayer error: %v", err)
	}
	if relayer == nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnRegisterRelayer, relayer is not applied")
	}
	if !bytes.Equal(relayer.Address, params.Address) {
		return utils.BYTE_FALSE, fmt.Errorf("UnRegisterRelayer, address is not relayer owner")
	}

	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(RELAYER_APPLY), params.Address))

	return utils.BYTE_TRUE, nil
}

func ApproveRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRelayer, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRelayer, checkWitness error: %v", err)
	}

	err = approveRelayer(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRelayer, approveRelayer error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func RejectRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectRelayer, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectRelayer, checkWitness error: %v", err)
	}
	contract := utils.RelayerManagerContractAddress

	//get relayer apply
	relayerRaw, err := GetRelayerApplyRaw(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectRelayer, get relayerRaw error: %v", err)
	}
	if relayerRaw == nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectRelayer, relayer is not applied")
	}

	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(RELAYER_APPLY), params.Address))

	return utils.BYTE_TRUE, nil
}

func BlackRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackRelayer, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackRelayer, checkWitness error: %v", err)
	}
	contract := utils.RelayerManagerContractAddress

	for _, address := range params.AddressList {
		//get relayer
		relayerRaw, err := GetRelayerRaw(native, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("BlackRelayer, get relayer error: %v", err)
		}
		if relayerRaw == nil {
			return utils.BYTE_FALSE, fmt.Errorf("BlackRelayer, relayer is not registered")
		}

		//put peer into black list
		native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(RELAYER_BLACK), address), relayerRaw)

		native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(RELAYER), address))
	}

	return utils.BYTE_TRUE, nil
}

func WhiteRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerListParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WhiteRelayer, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, checkWitness error: %v", err)
	}
	contract := utils.RelayerManagerContractAddress

	for _, address := range params.AddressList {
		blacked, err := checkIfBlacked(native, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("whiteNode, checkIfBlacked error: %v", err)
		}
		if !blacked {
			return utils.BYTE_FALSE, fmt.Errorf("whiteNode, relayer is not blacked")
		}
		native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(RELAYER_BLACK), address))
	}

	return utils.BYTE_TRUE, nil
}

func QuitRelayer(native *native.NativeService) ([]byte, error) {
	params := new(RelayerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitRelayer, contract params deserialize error: %v", err)
	}
	address, err := common.AddressParseFromBytes(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitRelayer, common.AddressParseFromBytes error: %v", err)
	}
	contract := utils.RelayerManagerContractAddress

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitRelayer, checkWitness error: %v", err)
	}

	//get relayer
	relayer, err := GetRelayer(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitRelayer, get relayer error: %v", err)
	}
	if relayer == nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitRelayer, relayer is not registered")
	}
	if !bytes.Equal(relayer.Address, params.Address) {
		return utils.BYTE_FALSE, fmt.Errorf("QuitRelayer, address is not relayer owner")
	}

	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(RELAYER), params.Address))

	return utils.BYTE_TRUE, nil
}
