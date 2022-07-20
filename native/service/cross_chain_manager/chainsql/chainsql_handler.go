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
package chainsql

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/header_sync/chainsql"
	hcom "github.com/polynetwork/poly/native/service/header_sync/common"
)

type ChainsqlHandler struct{}

func NewChainsqlHandler() *ChainsqlHandler {
	return &ChainsqlHandler{}
}

func (this *ChainsqlHandler) MakeDepositProposal(ns *native.NativeService) (*common.MakeTxParam, error) {
	params := new(common.EntranceParam)
	if err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput())); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, contract params deserialize error: %v", err)
	}
	val := &common.MakeTxParam{}
	if err := val.Deserialization(pcom.NewZeroCopySource(params.Extra)); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, failed to deserialize MakeTxParam: %v", err)
	}
	if err := common.CheckDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, check done transaction error: %v", err)
	}
	if err := common.PutDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, PutDoneTx error: %v", err)
	}

	root, err := chainsql.GetChainsqlRoot(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, failed to get the chainsql root cert: %v", err)
	}
	now := ns.GetBlockTime()
	if now.After(root.RootCA.NotAfter) || now.Before(root.RootCA.NotBefore) {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, Chainsql root CA need to update for chain %d: (start: %d, end: %d, block_time: %d)",
			params.SourceChainID, root.RootCA.NotBefore.Unix(), root.RootCA.NotAfter.Unix(), now.Unix())
	}
	certs := &hcom.CertTrustChain{}
	if err := certs.Deserialization(pcom.NewZeroCopySource(params.HeaderOrCrossChainMsg)); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, failed to deserialize CertTrustChain: %v", err)
	}
	if err := certs.Validate(ns); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, validate not pass: %v", err)
	}
	if err := certs.CheckSigWithRootCert(root.RootCA, params.Extra, params.Proof); err != nil {
		return nil, fmt.Errorf("Chainsql MakeDepositProposal, failed to check sig: %v", err)
	}
	PutChainsqlLatestHeightInProcessing(ns, params.SourceChainID, val.FromContractAddress, params.Height)

	return val, nil
}
