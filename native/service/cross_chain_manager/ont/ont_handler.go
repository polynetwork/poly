/*
 * Copyright (C) 2020 The poly network Authors
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

package ont

import (
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/header_sync/ont"
	"github.com/ontio/ontology-crypto/keypair"
	ocommon "github.com/ontio/ontology/common"
	otypes "github.com/ontio/ontology/core/types"
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
