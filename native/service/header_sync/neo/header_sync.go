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
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native/service/utils"
)

type NEOHandler struct {
}

func NewNEOHandler() *NEOHandler {
	return &NEOHandler{}
}

func (this *NEOHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("NeoHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("NeoHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("NeoHandler SyncGenesisHeader, checkWitness error: %v", err)
	}
	// Deserialize neo block header
	header := new(NeoBlockHeader)
	if err := header.Deserialization(common.NewZeroCopySource(params.GenesisHeader)); err != nil {
		return fmt.Errorf("NeoHandler SyncGenesisHeader, deserialize header err: %v", err)
	}
	if neoConsensus, _ := getConsensusValByChainId(native, params.ChainID); neoConsensus == nil {
		// Put NeoConsensus.NextConsensus into storage
		if err = putConsensusValByChainId(native, &NeoConsensus{
			ChainID:       params.ChainID,
			Height:        header.Index,
			NextConsensus: header.NextConsensus,
		}); err != nil {
			return fmt.Errorf("NeoHandler SyncGenesisHeader, update ConsensusPeer error: %v", err)
		}
	}
	return nil
}

func (this *NEOHandler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	neoConsensus, err := getConsensusValByChainId(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("SyncBlockHeader, the consensus validator has not been initialized, chainId: %d", params.ChainID)
	}
	var newNeoConsensus *NeoConsensus
	for _, v := range params.Headers {
		header := new(NeoBlockHeader)
		if err := header.Deserialization(common.NewZeroCopySource(v)); err != nil {
			return fmt.Errorf("SyncBlockHeader, NeoBlockHeaderFromBytes error: %v", err)
		}
		if !header.NextConsensus.Equals(neoConsensus.NextConsensus) && header.Index > neoConsensus.Height {
			if err = verifyHeader(native, params.ChainID, header); err != nil {
				return fmt.Errorf("SyncBlockHeader, verifyHeader error: %v", err)
			}
			newNeoConsensus = &NeoConsensus{
				ChainID:       neoConsensus.ChainID,
				Height:        header.Index,
				NextConsensus: header.NextConsensus,
			}
		}
	}
	if newNeoConsensus != nil {
		if err = putConsensusValByChainId(native, newNeoConsensus); err != nil {
			return fmt.Errorf("SyncBlockHeader, update ConsensusPeer error: %v", err)
		}
	}
	return nil
}

func (this *NEOHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
