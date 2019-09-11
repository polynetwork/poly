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

package neo

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
)

const (
	VERIFY_FROM_NEO_PROOF = "verifyFromNeoProof"
	MAKE_TO_NEO_PROOF     = "makeToNeoProof"

	//key prefix
	DONE_TX = "doneTx"
	REQUEST = "request"
)

type NEOHandler struct {
}

func NewNEOHandler() *NEOHandler {
	return &NEOHandler{}
}

func (this *NEOHandler) Vote(service *native.NativeService) (bool, *crosscommon.MakeTxParam, error) {
	return true, nil, nil
}

func (this *NEOHandler) MakeDepositProposal(service *native.NativeService) (*crosscommon.MakeTxParam, error) {
	params := new(crosscommon.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("ont Verify, contract params deserialize error: %v", err)
	}

	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return nil, fmt.Errorf("ont Verify, hex.DecodeString proof error: %v", err)
	}
	merkleValue, err := VerifyFromNeoTx(service, proof, params.SourceChainID, params.Height)
	if err != nil {
		return nil, fmt.Errorf("ont Verify, VerifyOntTx error: %v", err)
	}

	makeTxParam := &crosscommon.MakeTxParam{
		TxHash:              merkleValue.TxHash.ToHexString(),
		FromChainID:         merkleValue.CreateCrossChainTxMerkle.FromChainID,
		FromContractAddress: merkleValue.CreateCrossChainTxMerkle.FromContractAddress,
		ToChainID:           merkleValue.CreateCrossChainTxMerkle.ToChainID,
		Method:              merkleValue.CreateCrossChainTxMerkle.Method,
		Args:                merkleValue.CreateCrossChainTxMerkle.Args,
	}
	return makeTxParam, nil
}

func (this *NEOHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam) error {
	err := MakeToNeoProof(service, param)
	if err != nil {
		return fmt.Errorf("ont MakeTransaction, MakeToOntProof error: %v", err)
	}
	return nil
}
