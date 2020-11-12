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
package fabric

import (
	"encoding/hex"
	"errors"
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	hcom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/fabric"
)

type FabricHandler struct{}

func NewFabricHandler() *FabricHandler {
	return &FabricHandler{}
}

func (this *FabricHandler) MakeDepositProposal(ns *native.NativeService) (*common.MakeTxParam, error) {
	params := new(common.EntranceParam)
	if err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput())); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, contract params deserialize error: %v", err)
	}

	sideChain, err := side_chain_manager.GetSideChain(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain == nil {
		return nil, errors.New("Fabric MakeDepositProposal, side chain not found")
	}
	strategyTy := side_chain_manager.FabricVerifyStrategy(sideChain.BlocksToWait)

	val := &common.MakeTxParam{}
	if err := val.Deserialization(pcom.NewZeroCopySource(params.Extra)); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to deserialize MakeTxParam: %v", err)
	}
	if err := common.CheckDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, check done transaction error: %v", err)
	}
	if err := common.PutDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, PutDoneTx error: %v", err)
	}
	rootCerts, err := fabric.GetFabricRoot(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to get the Fabric root certs: %v", err)
	}
	l := len(rootCerts.Certs)
	rootCerts = rootCerts.ValidCAs(ns)
	validL := len(rootCerts.Certs)
	if validL == 0 {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, no valid root CA in poly's storage for Fabric chain %d", params.SourceChainID)
	}
	if validL < l {
		if err := fabric.PutFabricRoot(ns, rootCerts, params.SourceChainID); err != nil {
			return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to put valid fabric root CAs: %v", err)
		}
	}
	certs := hcom.MultiCertTrustChain(make([]*hcom.CertTrustChain, 0))
	if certs, err = certs.Deserialization(pcom.NewZeroCopySource(params.HeaderOrCrossChainMsg)); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to deserialize CertTrustChain: %v", err)
	}
	if err := certs.ValidateAll(ns); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to validate CAs: %v", err)
	}

	sigs := GetSigArr(params.Proof)
	if len(sigs) != len(certs) {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, number of siguratures %d is not equal with number of trust chains %d", len(sigs), len(certs))
	}
	isChecked := make(map[string]bool)
	for i, trustChain := range certs {
		rawKey := hex.EncodeToString(trustChain.Certs[0].Raw)
		if isChecked[rawKey] {
			continue
		}
		var isExist bool
		for _, root := range rootCerts.Certs {
			if root.Equal(trustChain.Certs[0]) {
				isExist = true
				break
			}
		}
		if !isExist {
			continue
		}
		// TODO: need the signed raw info and the sig
		//lastElem := trustChain.Certs[len(trustChain.Certs)-1]
		//for _, ou := range lastElem.Subject.OrganizationalUnit {
		//	if strings.Contains(strings.ToLower(ou), "peer") {
		//		isExist = false
		//	}
		//}
		//if isExist {
		//	continue
		//}
		if err := trustChain.CheckSig(params.Extra, sigs[i]); err != nil {
			return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to check signature for No.%d chain: %v", i, err)
		}
		isChecked[rawKey] = true
	}

	switch strategyTy {
	case side_chain_manager.JustOne:
		if len(isChecked) == 0 {
			return nil, errors.New("Fabric MakeDepositProposal, at least one valid trust chain commited but none found.")
		}
	case side_chain_manager.OverTwoThirds:
		if 3*len(isChecked) <= 2*len(rootCerts.Certs) {
			return nil, fmt.Errorf("Fabric MakeDepositProposal, your valid trust chain is not over 2/3 of required and %d valid but %d total.", len(isChecked), len(rootCerts.Certs))
		}
	case side_chain_manager.AllNeeded:
		if len(isChecked) != len(rootCerts.Certs) {
			return nil, fmt.Errorf("Fabric MakeDepositProposal, only %d valid trust chain commited but %d needed.", len(isChecked), len(rootCerts.Certs))
		}
	default:
		return nil, fmt.Errorf("Fabric MakeDepositProposal, strategy not support: %d", strategyTy)
	}

	return val, nil
}
