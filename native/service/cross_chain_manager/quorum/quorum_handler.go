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
package quorum

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/quorum"
)

type QuorumHandler struct{}

func NewQuorumHandler() *QuorumHandler {
	return &QuorumHandler{}
}

func (this *QuorumHandler) MakeDepositProposal(ns *native.NativeService) (*common.MakeTxParam, error) {
	params := new(common.EntranceParam)
	if err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput())); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, contract params deserialize error: %v", err)
	}

	sideChain, err := side_chain_manager.GetSideChain(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain == nil {
		return nil, errors.New("Quorum MakeDepositProposal, side chain not found")
	}

	val := &common.MakeTxParam{}
	if err := val.Deserialization(pcom.NewZeroCopySource(params.Extra)); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, failed to deserialize MakeTxParam: %v", err)
	}
	if err := common.CheckDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, check done transaction error: %v", err)
	}
	if err := common.PutDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, PutDoneTx error: %v", err)
	}

	header := &types.Header{}
	if err := json.Unmarshal(params.HeaderOrCrossChainMsg, header); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, deserialize header err: %v", err)
	}
	valh, err := quorum.GetCurrentValHeight(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, failed to get current validators height: %v", err)
	}
	if header.Number.Uint64() < valh {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, height of header %d is less than epoch height %d", header.Number.Uint64(), valh)
	}
	vs, err := quorum.GetValSet(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, failed to get quorum validators: %v", err)
	}
	if _, err := quorum.VerifyQuorumHeader(vs, header, false); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, failed to verify quorum header %s: %v", header.Hash().String(), err)
	}

	if err := verifyFromQuorumTx(params.Proof, params.Extra, header, sideChain); err != nil {
		return nil, fmt.Errorf("Quorum MakeDepositProposal, verifyFromEthTx error: %s", err)
	}

	return val, nil
}
