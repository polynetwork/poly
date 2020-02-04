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
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"

	"bytes"
	"encoding/hex"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/header_sync/ont"
	"github.com/ontio/multi-chain/native/service/utils"
	ocommon "github.com/ontio/ontology/common"
	otypes "github.com/ontio/ontology/core/types"
	ontccm "github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	utils2 "github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ONTHandler struct {
}

func NewONTHandler() *ONTHandler {
	return &ONTHandler{}
}

func (this *ONTHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("ont MakeDepositProposal, contract params deserialize error: %v", err)
	}

	crossChainMsg, err := ont.GetCrossChainMsg(service, params.SourceChainID, params.Height)
	if crossChainMsg == nil {
		source := ocommon.NewZeroCopySource(params.HeaderOrCrossChainMsg)
		crossChainMsg = new(otypes.CrossChainMsg)
		err := crossChainMsg.Deserialization(source)
		if err != nil {
			return nil, fmt.Errorf("ont MakeDepositProposal, deserialize crossChainMsg error: %v", err)
		}
		n, _, irr, eof := source.NextVarUint()
		if irr || eof {
			return nil, fmt.Errorf("ont MakeDepositProposal, deserialization bookkeeper length error")
		}
		var bookkeepers []keypair.PublicKey
		for i := 0; uint64(i) < n; i++ {
			v, _, irr, eof := source.NextVarBytes()
			if irr || eof {
				return nil, fmt.Errorf("ont MakeDepositProposal, deserialization bookkeeper error")
			}
			bookkeeper, err := keypair.DeserializePublicKey(v)
			if err != nil {
				return nil, fmt.Errorf("ont MakeDepositProposal, keypair.DeserializePublicKey error: %v", err)
			}
			bookkeepers = append(bookkeepers, bookkeeper)
		}
		err = ont.VerifyCrossChainMsg(service, params.SourceChainID, crossChainMsg, bookkeepers)
		if err != nil {
			return nil, fmt.Errorf("ont MakeDepositProposal, VerifyCrossChainMsg error: %v", err)
		}
		err = ont.PutCrossChainMsg(service, params.SourceChainID, crossChainMsg)
		if err != nil {
			return nil, fmt.Errorf("ont MakeDepositProposal, put PutCrossChainMsg error: %v", err)
		}
	}

	value, err := VerifyFromOntTx(params.Proof, crossChainMsg)
	if err != nil {
		return nil, fmt.Errorf("ont MakeDepositProposal, VerifyOntTx error: %v", err)
	}
	if err := scom.CheckDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("ont MakeDepositProposal, check done transaction error:%s", err)
	}
	if err = scom.PutDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, putDoneTx error:%s", err)
	}
	return value, nil
}

func (this *ONTHandler) ProcessMultiChainTx(service *native.NativeService, txParam *scom.MakeTxParam) ([]byte, error) {
	//target chain is multi-chain
	if txParam.ToChainID == service.GetChainID() {
		if !bytes.Equal(txParam.ToContractAddress, utils.OntLockProxyContractAddress[:]) {
			return utils.BYTE_FALSE, fmt.Errorf("[Ont ProcessTx], to contract address id is not multi-chain Ont contract address, expect:%s, get:%s",
				utils.OntLockProxyContractAddress.ToHexString(), hex.EncodeToString(common.ToArrayReverse(txParam.ToContractAddress)))
		}

		input := getUnlockArgs(txParam.Args, txParam.FromContractAddress, ontccm.ONT_CHAIN_ID)
		res, err := service.NativeCall(utils.OntLockProxyContractAddress, txParam.Method, input)
		if !bytes.Equal(res.([]byte), utils.BYTE_TRUE) || err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[ProcessMultiChainTx] OntUnlock, error:%s", err)
		}
	}
	return utils.BYTE_TRUE, nil
}

func (this *ONTHandler) CreateTx(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(ontccm.CreateCrossChainTxParam)
	if err := params.Deserialization(ocommon.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("[CreateTx], contract params deserialize error: %v", err)
	}

	if !bytes.Equal(utils2.OntLockContractAddress[:], params.ToContractAddress) {
		return nil, fmt.Errorf("[CreateTx], ToContractAddress is ont lock proxy contract address in ontology network!")
	}

	if !service.CheckWitness(utils.OntLockProxyContractAddress) {
		return nil, fmt.Errorf("[CreateTx] should be invoked by ont lock proxy contract, checkwitness failed!")
	}
	txHash := service.GetTx().Hash()

	txParam := &scom.MakeTxParam{
		TxHash:              txHash.ToArray(),
		CrossChainID:        txHash.ToArray(),
		FromContractAddress: utils.CrossChainManagerContractAddress[:],
		ToChainID:           params.ToChainID,
		ToContractAddress:   params.ToContractAddress,
		Method:              params.Method,
		Args:                params.Args,
	}
	return txParam, nil
}
