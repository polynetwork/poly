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

package neo

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/neo"
)

type NEOHandler struct {
}

func NewNEOHandler() *NEOHandler {
	return &NEOHandler{}
}

func (this *NEOHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("neo MakeDepositProposal, contract params deserialize error: %v", err)
	}

	crossChainMsg, err := neo.GetCrossChainMsg(service, params.SourceChainID, params.Height)
	if crossChainMsg == nil {
		crossChainMsg = new(neo.NeoCrossChainMsg)
		if err := crossChainMsg.Deserialization(common.NewZeroCopySource(params.HeaderOrCrossChainMsg)); err != nil {
			return nil, fmt.Errorf("neo MakeDepositProposal, deserialize crossChainMsg error: %v", err)
		}
		err = neo.VerifyCrossChainMsg(service, params.SourceChainID, crossChainMsg)
		if err != nil {
			return nil, fmt.Errorf("neo MakeDepositProposal, VerifyCrossChainMsg error: %v", err)
		}
		err = neo.PutCrossChainMsg(service, params.SourceChainID, crossChainMsg)
		if err != nil {
			return nil, fmt.Errorf("neo MakeDepositProposal, put PutCrossChainMsg error: %v", err)
		}
	}
	sideChain, err := side_chain_manager.GetSideChain(service, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("neo MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}
	value, err := verifyFromNeoTx(params.Proof, crossChainMsg, sideChain.CCMCAddress)
	if err != nil {
		return nil, fmt.Errorf("neo MakeDepositProposal, VerifyFromNeoTx error: %v", err)
	}
	if err := scom.CheckDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("neo MakeDepositProposal, check done transaction error:%s", err)
	}
	if err = scom.PutDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("neo MakeDepositProposal, putDoneTx error:%s", err)
	}
	return value, nil
}
