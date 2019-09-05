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
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native/service/native"
	crosscommon "github.com/ontio/multi-chain/native/service/native/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/native/utils"
	"github.com/ontio/multi-chain/vm/neovm/types"
)

const (
	CREATE_CROSS_CHAIN_TX  = "createCrossChainTx"
	PROCESS_CROSS_CHAIN_TX = "processCrossChainTx"
	MAKE_FROM_ONT_PROOF    = "makeFromOntProof"
	MAKE_TO_ONT_PROOF      = "makeToOntProof"

	//key prefix
	DONE_TX = "doneTx"
	REQUEST = "request"
)

type ONTHandler struct {
}

func NewONTHandler() *ONTHandler {
	return &ONTHandler{}
}

func (this *ONTHandler) Vote(service *native.NativeService) (bool, *crosscommon.MakeTxParam, error) {
	return true, nil, nil
}

func (this *ONTHandler) MakeDepositProposal(service *native.NativeService) (*crosscommon.MakeTxParam, error) {
	params := new(crosscommon.EntranceParam)
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

	makeTxParam := &crosscommon.MakeTxParam{
		FromChainID:         merkleValue.CreateCrossChainTxMerkle.FromChainID,
		FromContractAddress: merkleValue.CreateCrossChainTxMerkle.FromContractAddress,
		ToChainID:           merkleValue.CreateCrossChainTxMerkle.ToChainID,
		ToAddress:           merkleValue.CreateCrossChainTxMerkle.ToAddress,
		Amount:              new(big.Int).SetUint64(merkleValue.CreateCrossChainTxMerkle.Amount),
	}
	return makeTxParam, nil
}

func (this *ONTHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam) error {
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
	dest, err := common.AddressFromHexString(merkleValue.ToContractAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, common.AddressFromHexString error: %v", err)
	}
	functionName := "unlock"
	addr, err := common.AddressFromBase58(merkleValue.MakeTxParam.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, common.AddressFromBase58 error: %v", err)
	}
	args := &OngUnlockParam{
		FromChainID: merkleValue.MakeTxParam.FromChainID,
		Address:     addr,
		Amount:      merkleValue.MakeTxParam.Amount.Uint64(),
	}
	sink := common.NewZeroCopySink(nil)
	args.Serialization(sink)
	buf := bytes.NewBuffer(nil)
	err = args.NeoVmSerialization(buf)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, args.NeoVmSerialization error: %v", err)
	}
	if dest == utils.OngContractAddress {
		if _, err := native.NativeCall(dest, functionName, sink.Bytes()); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, native.NativeCall error: %v", err)
		}
	} else {
		res, err := native.NeoVMCall(dest, functionName, buf.Bytes())
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
