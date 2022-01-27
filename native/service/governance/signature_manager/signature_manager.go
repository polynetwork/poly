/*
 * Copyright (C) 2022 The poly network Authors
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
package signature_manager

import (
	"crypto/sha256"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//function name
	ADD_SIGNATURE = "addSignature"
)

//Register methods of signature_manager contract
func RegisterSignatureManagerContract(native *native.NativeService) {
	native.Register(ADD_SIGNATURE, AddSignature)
}

func AddSignature(native *native.NativeService) ([]byte, error) {
	params := new(AddSignatureParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddSignature, contract params deserialize error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, params.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddSignature, checkWitness: %s, error: %v", params.Address.ToBase58(), err)
	}

	temp := sha256.Sum256(params.Subject)
	id := temp[:]
	//check consensus signs
	ok, err := CheckSigns(native, id, params.Signature, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddSignature, CheckSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.SignatureManagerContractAddress,
			States:          []interface{}{"AddSignatureQuorum", id, params.Subject, params.SideChainID},
		})
	return utils.BYTE_TRUE, nil

}
