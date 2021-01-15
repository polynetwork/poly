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
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

type QuorumHandler struct{}

func NewQuorumHandler() *QuorumHandler {
	return &QuorumHandler{}
}

func (h *QuorumHandler) SyncGenesisHeader(ns *native.NativeService) error {
	params := new(common.SyncGenesisHeaderParam)
	if err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput())); err != nil {
		return fmt.Errorf("QuorumHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	op, err := node_manager.GetCurConOperator(ns)
	if err != nil {
		return fmt.Errorf("QuorumHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(ns, op)
	if err != nil {
		return fmt.Errorf("QuorumHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	header := &types.Header{}
	if err = json.Unmarshal(params.GenesisHeader, header); err != nil {
		return fmt.Errorf("QuorumHandler SyncGenesisHeader, deserialize header err: %v", err)
	}
	extra, err := ExtractIstanbulExtra(header)
	if err != nil {
		return fmt.Errorf("QuorumHandler SyncGenesisHeader, failed to ExtractIstanbulExtra: %v", err)
	}

	putValSet(ns, params.ChainID, header.Number.Uint64(), extra.Validators)
	return nil
}

func (h *QuorumHandler) SyncBlockHeader(ns *native.NativeService) error {
	params := new(common.SyncBlockHeaderParam)
	err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput()))
	if err != nil {
		return fmt.Errorf("QuorumHandler SyncBlockHeader, contract params deserialize error: %v", err)
	}
	if len(params.Headers) == 0 {
		return errors.New("QuorumHandler SyncBlockHeader, none headers in input")
	}

	currh, err := GetCurrentValHeight(ns, params.ChainID)
	if err != nil {
		return fmt.Errorf("QuorumHandler SyncBlockHeader, failed to get current validator height: %v", err)
	}
	vs, err := GetValSet(ns, params.ChainID)
	if err != nil {
		return fmt.Errorf("QuorumHandler SyncBlockHeader, failed to get validators: %v", err)
	}
	header := &types.Header{}
	for i, v := range params.Headers {
		if err := json.Unmarshal(v, header); err != nil {
			return fmt.Errorf("QuorumHandler SyncBlockHeader, deserialize No.%d header err: %v", i, err)
		}
		h := header.Number.Uint64()
		if currh >= h {
			return fmt.Errorf("QuorumHandler SyncBlockHeader, wrong height of No.%d header: (curr: %d, commit: %d)", i, currh, h)
		}

		extra, err := VerifyQuorumHeader(vs, header, true)
		if err != nil {
			return fmt.Errorf("QuorumHandler SyncBlockHeader, failed to verify No.%d quorum header %s: %v", i, GetQuorumHeaderHash(header).String(), err)
		}

		currh, vs = h, extra.Validators
	}

	putValSet(ns, params.ChainID, currh, vs)
	return nil
}

func (h *QuorumHandler) SyncCrossChainMsg(ns *native.NativeService) error {
	return nil
}
