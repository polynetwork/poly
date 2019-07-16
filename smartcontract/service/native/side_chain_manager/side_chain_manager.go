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
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

const (
	//function name
	REGISTER_SIDE_CHAIN = "registerSideChain"

	//key prefix
	REGISTER_SIDE_CHAIN_REQUEST = "registerSideChainRequest"
	UPDATE_SIDE_CHAIN_REQUEST   = "updateSideChainRequest"
	QUIT_SIDE_CHAIN_REQUEST     = "quitSideChainRequest"
	SIDE_CHAIN                  = "sideChain"
)

//Init contract address
func InitSideChainManager() {
	native.Contracts[utils.SideChainManagerContractAddress] = RegisterSideChainManagerContract
}

//Register methods of governance contract
func RegisterSideChainManagerContract(native *native.NativeService) {
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
	}



	return utils.BYTE_TRUE, nil
}
