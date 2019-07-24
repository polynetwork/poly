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

package cross_chain

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/chain_manager"
	"github.com/ontio/multi-chain/smartcontract/service/native/header_sync"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
	"github.com/ontio/multi-chain/vm/neovm/types"
)

const (
	CREATE_CROSS_CHAIN_TX  = "createCrossChainTx"
	PROCESS_CROSS_CHAIN_TX = "processCrossChainTx"

	//key prefix
	REQUEST_ID  = "requestID"
	REQUEST     = "request"
	CURRENT_ID  = "currentID"
	REMAINED_ID = "remainedID"
)

//Init governance contract address
func InitCrossChain() {
	native.Contracts[utils.CrossChainContractAddress] = RegisterCrossChianContract
}

//Register methods of governance contract
func RegisterCrossChianContract(native *native.NativeService) {
	native.Register(CREATE_CROSS_CHAIN_TX, CreateCrossChainTx)
	native.Register(PROCESS_CROSS_CHAIN_TX, ProcessCrossChainTx)
}

func CreateCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(CreateCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, contract params deserialize error: %v", err)
	}

	err := MakeOntProof(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, MakeOntProof error: %v", err)
	}

	//process main chain ongx fee
	//update side chain
	sideChain, err := chain_manager.GetSideChain(native, params.ToChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, get sideChain error: %v", err)
	}
	if sideChain.Status != chain_manager.SideChainStatus && sideChain.Status != chain_manager.QuitingStatus {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, side chain status is not normal status")
	}
	ongFee, ok := common.SafeMul(uint64(params.Fee), sideChain.Ratio)
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, number is more than uint64")
	}
	sideChain.OngNum = sideChain.OngNum + ongFee
	if sideChain.OngNum > sideChain.OngPool {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, ong num in pool is full")
	}
	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, put sideChain error: %v", err)
	}
	//ong transfer
	err = appCallTransferOng(native, params.Address, utils.CrossChainContractAddress, ongFee)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("OngLock, ong transfer error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func ProcessCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(ProcessCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, contract params deserialize error: %v", err)
	}

	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, proof hex.DecodeString error: %v", err)
	}
	merkleValue, err := VerifyOntTx(native, proof, params.FromChainID, params.Height)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, VerifyOntTx error: %v", err)
	}

	//process main chain ongx fee
	//get side chain
	sideChain, err := chain_manager.GetSideChain(native, params.FromChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, get sideChain error: %v", err)
	}
	if sideChain.Status != chain_manager.SideChainStatus && sideChain.Status != chain_manager.QuitingStatus {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, side chain status is not normal status")
	}
	ongFee, ok := common.SafeMul(uint64(merkleValue.CreateCrossChainTxParam.Fee), sideChain.Ratio)
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, number is more than uint64")
	}
	if sideChain.OngNum < ongFee {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, ong num in pool is not enough")
	}
	sideChain.OngNum = sideChain.OngNum - ongFee
	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, put sideChain error: %v", err)
	}

	//get sync address
	syncAddress, err := header_sync.GetSyncAddress(native, params.FromChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, get syncAddress error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, utils.CrossChainContractAddress, syncAddress, ongFee/10)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, appCallTransferOng ong transfer error: %v", err)
	}
	err = appCallTransferOng(native, utils.CrossChainContractAddress, params.Address, ongFee-ongFee/10)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, appCallTransferOng ong transfer error: %v", err)
	}

	//call cross chain function
	destContractAddr := merkleValue.CreateCrossChainTxParam.ContractAddress
	functionName := merkleValue.CreateCrossChainTxParam.FunctionName
	args := merkleValue.CreateCrossChainTxParam.Args
	if destContractAddr == utils.OngContractAddress {
		if _, err := native.NativeCall(destContractAddr, functionName, args); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, native.NativeCall error: %v", err)
		}
	} else {
		res, err := native.NeoVMCall(destContractAddr, functionName, args)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, native.NeoVMCall error: %v", err)
		}
		r, ok := res.(*types.Integer)
		if !ok {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, res of neo vm call must be bool")
		}
		v, _ := r.GetBigInteger()
		if v.Cmp(new(big.Int).SetUint64(0)) == 0 {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, res of neo vm call is false")
		}
	}

	return utils.BYTE_TRUE, nil
}
