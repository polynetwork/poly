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

package ont

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
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

type ONTHandler struct {
}

func NewONTHandler() *ONTHandler {
	return &ONTHandler{}
}

func (this *ONTHandler) Verify(service *native.NativeService) (*inf.MakeTxParam, error) {
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return nil, fmt.Errorf("ont Verify, contract params deserialize error: %v", err)
	}

	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return nil, fmt.Errorf("ont Verify, hex.DecodeString proof error: %v", err)
	}
	merkleValue, err := VerifyFromOntTx(service, proof, params.SourceChainID, params.Height)
	if err != nil {
		return nil, fmt.Errorf("ont Verify, VerifyOntTx error: %v", err)
	}

	makeTxParam := &inf.MakeTxParam{
		FromChainID:         merkleValue.CreateCrossChainTxParam.FromChainID,
		FromContractAddress: merkleValue.CreateCrossChainTxParam.FromContractAddress,
		ToChainID:           merkleValue.CreateCrossChainTxParam.ToChainID,
		Address:             merkleValue.CreateCrossChainTxParam.Address.ToBase58(),
		Amount:              merkleValue.CreateCrossChainTxParam.Amount,
	}
	return makeTxParam, nil
}

func (this *ONTHandler) MakeTransaction(service *native.NativeService, param *inf.MakeTxParam) error {
	err := MakeToOntProof(service, param)
	if err != nil {
		return fmt.Errorf("ont MakeTransaction, MakeToOntProof error: %v", err)
	}

	return nil
}

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

	err := MakeFromOntProof(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, MakeOntProof error: %v", err)
	}

	//TODO: miner fee?

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
	merkleValue, err := VerifyToOntTx(native, proof, params.FromChainID, params.Height)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, VerifyOntTx error: %v", err)
	}

	//call cross chain function
	destContractAddr, err := side_chain_manager.GetAssetContractAddress(native, params.FromChainID,
		native.ShardID.ToUint64(), merkleValue.MakeTxParam.FromContractAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, side_chain_manager.GetAssetContractAddress error: %v", err)
	}
	dest, err := common.AddressFromHexString(destContractAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, common.AddressFromHexString error: %v", err)
	}
	functionName := "unlock"
	addr, err := common.AddressFromBase58(merkleValue.MakeTxParam.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, common.AddressFromBase58 error: %v", err)
	}
	args := &OngUnlockParam{
		FromChainID: merkleValue.MakeTxParam.FromChainID,
		Address:     addr,
		Amount:      merkleValue.MakeTxParam.Amount,
	}
	sink := common.NewZeroCopySink(nil)
	args.Serialization(sink)

	if dest == utils.OngContractAddress {
		if _, err := native.NativeCall(dest, functionName, sink.Bytes()); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, native.NativeCall error: %v", err)
		}
	} else {
		res, err := native.NeoVMCall(dest, functionName, sink.Bytes())
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

	//TODO: miner fee?

	return utils.BYTE_TRUE, nil
}
