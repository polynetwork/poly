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

package neo3legacy

import (
	"fmt"
	"github.com/joeqian10/neo3-gogogo-legacy/helper"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/neo3legacy"
)

type Neo3Handler struct {
}

func NewNeo3Handler() *Neo3Handler {
	return &Neo3Handler{}
}

func (this *Neo3Handler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, contract params deserialize error: %v", err)
	}
	// Deserialize neo cross chain msg and verify its signature
	crossChainMsg := new(neo3legacy.NeoCrossChainMsg)
	if err := crossChainMsg.Deserialization(common.NewZeroCopySource(params.HeaderOrCrossChainMsg)); err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, deserialize crossChainMsg error: %v", err)
	}
	// Verify the validity of proof with the help of state root in verified neo cross chain msg
	sideChain, err := side_chain_manager.GetSideChain(service, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}
	if err := neo3legacy.VerifyCrossChainMsgSig(service, helper.BytesToUInt32(sideChain.ExtraInfo), crossChainMsg); err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, VerifyCrossChainMsg error: %v", err)
	}

	// when register neo N3, convert ccmc id to []byte
	// convert neo3 contract address bytes to id, it is different from other chains
	// need to store int in a []byte, contract id can be get from "getcontractstate" api
	// neo3 native contracts have negative ids, while custom contracts have positive ones
	id := int(int32(helper.BytesToUInt32(sideChain.CCMCAddress)))

	value, err := verifyFromNeoTx(params.Proof, crossChainMsg, id)
	if err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, VerifyFromNeoTx error: %v", err)
	}
	// Ensure the tx has not been processed before, and mark the tx as processed
	if err := scom.CheckDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, check done transaction error:%s", err)
	}
	if err = scom.PutDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("neo3 MakeDepositProposal, putDoneTx error:%s", err)
	}
	return value, nil
}
