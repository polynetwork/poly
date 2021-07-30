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
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native/service/utils"
)

type Neo3Handler struct {
}

func NewNeo3Handler() *Neo3Handler {
	return &Neo3Handler{}
}

func (this *Neo3Handler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("Neo3Handler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("Neo3Handler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("Neo3Handler SyncGenesisHeader, checkWitness error: %v", err)
	}
	// Deserialize neo block header
	header := new(NeoBlockHeader)
	if err := header.Deserialization(common.NewZeroCopySource(params.GenesisHeader)); err != nil {
		return fmt.Errorf("Neo3Handler SyncGenesisHeader, deserialize header err: %v", err)
	}

	if neoConsensus, _ := getConsensusValByChainId(native, params.ChainID); neoConsensus == nil {
		// Put NeoConsensus.NextConsensus into storage
		if err = putConsensusValByChainId(native, &NeoConsensus{
			ChainID:       params.ChainID,
			Height:        header.GetIndex(),
			NextConsensus: header.GetNextConsensus(),
		}); err != nil {
			return fmt.Errorf("Neo3Handler SyncGenesisHeader, update ConsensusPeer error: %v", err)
		}
	}
	return nil
}

func (this *Neo3Handler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("Neo3Handler SyncBlockHeader, contract params deserialize error: %v", err)
	}
	neoConsensus, err := getConsensusValByChainId(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("Neo3Handler SyncBlockHeader, the consensus validator has not been initialized, chainId: %d", params.ChainID)
	}
	sideChain, err := side_chain_manager.GetSideChain(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("neo3 MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}
	var newNeoConsensus *NeoConsensus
	for _, v := range params.Headers {
		header := new(NeoBlockHeader)
		if err := header.Deserialization(common.NewZeroCopySource(v)); err != nil {
			return fmt.Errorf("Neo3Handler SyncBlockHeader, NeoBlockHeaderFromBytes error: %v", err)
		}
		if !header.GetNextConsensus().Equals(neoConsensus.NextConsensus) && header.GetIndex() > neoConsensus.Height {
			if err = verifyHeader(native, params.ChainID, header, helper.BytesToUInt32(sideChain.ExtraInfo)); err != nil {
				return fmt.Errorf("Neo3Handler SyncBlockHeader, verifyHeader error: %v", err)
			}
			newNeoConsensus = &NeoConsensus{
				ChainID:       neoConsensus.ChainID,
				Height:        header.GetIndex(),
				NextConsensus: header.GetNextConsensus(),
			}
		}
	}
	if newNeoConsensus != nil {
		if err = putConsensusValByChainId(native, newNeoConsensus); err != nil {
			return fmt.Errorf("Neo3Handler SyncBlockHeader, update ConsensusPeer error: %v", err)
		}
	}
	return nil
}

func (this *Neo3Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
