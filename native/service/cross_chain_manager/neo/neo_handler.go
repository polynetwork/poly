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
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
)

type NEOHandler struct {
}

func NewNEOHandler() *NEOHandler {
	return &NEOHandler{}
}

func (this *NEOHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	//params := new(scom.EntranceParam)
	//if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
	//	return nil, fmt.Errorf("ont MakeDepositProposal, contract params deserialize error: %v", err)
	//}
	//
	//if err := scom.CheckDoneTx(service, params.TxHash, params.Proof, params.SourceChainID); err != nil {
	//	return nil, fmt.Errorf("MakeDepositProposal, check done transaction error:%s", err)
	//}
	//
	//value, err := verifyFromNEOTx(service, params.Proof, params.TxHash, params.SourceChainID, params.Height)
	//if err != nil {
	//	return nil, fmt.Errorf("ont MakeDepositProposal, VerifyOntTx error: %v", err)
	//}
	//
	//if err = scom.PutDoneTx(service, value.TxHash, params.Proof, params.SourceChainID); err != nil {
	//	return nil, fmt.Errorf("VerifyFromOntTx, putDoneTx error:%s", err)
	//}
	//return value, nil
	return nil, nil
}
