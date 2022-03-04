/*
 * Copyright (C) 2022 The poly network Authors
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

package harmony

import (
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

// Harmony Header Sync Handler
type Handler struct {}

func NewHandler() *Handler {
	return new(Handler)
}

// Sync Genesis header
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	err = params.Deserialization(common.NewZeroCopySource(native.GetInput()))
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to deserialize genesis header params, err: %v", err)
	}

	// Get current consensus address
	operator, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to parse current consensus operator, err: %v", err)
	}

	// Check consensus witness
	err = utils.ValidateOwner(native, operator)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to check operator witness, err: %v", err)
	}

	// Check genesis header existence
	headerExist, err := getGenesisHeader(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("HarmonyHandler get genesis header from storage failed, err:%v", err)
	}
	if headerExist != nil {
		return fmt.Errorf("HarmonyHandler genesis header was already set")
	}

	// Deserialize genesis header
	header, err := DecodeHeaderWithSig(params.GenesisHeader)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to deserialize harmony header, err: %v", err)
	}

	// Get side chain instance
	side, err := side_chain_manager.GetSideChain(native, params.ChainID)
	if err != nil || side == nil {
		return fmt.Errorf("HarmonyHandler failed to get side chain err: %v", err)
	}
	ctx, err := DecodeHarmonyContext(side.ExtraInfo)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to decode context, err: %v", err)
	}
	err = ctx.Init()
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to get network config and shard schedule: %v", err)
	}

	// Extract shard epoch
	epoch, err := header.ExtractEpoch()
	if err != nil {
		return fmt.Errorf("HarmonyHandler, failed to extract Epoch from header, err: %v", err)
	}

	if !ctx.IsLastBlock(header.Header.Number().Uint64()) {
		return fmt.Errorf("HarmonyHandler header block %s is not the last in epoch", header.Header.Number())
	}

	// Store genesis header
	err = storeGenesisHeader(native, params.ChainID, header)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to store genesis header, err: %v", err)
	}

	// Store epoch info
	err = storeEpoch(native, params.ChainID, epoch)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to store epoch info, err: %v", err)
	}

	// Update state
	err = updateWithHeader(native, params.ChainID, header.Header)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to update state, err: %v", err)
	}
	return
}

// Sync block header
func (h *Handler) SyncBlockHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncBlockHeaderParam)
	err = params.Deserialization(common.NewZeroCopySource(native.GetInput()))
	if err != nil {
		return fmt.Errorf("%w, HarmonyHandler failed to deserialize headers params", err)
	}

	// Get side chain instance
	side, err := side_chain_manager.GetSideChain(native, params.ChainID)
	if err != nil || side == nil {
		return fmt.Errorf("HarmonyHandler failed to get side chain err: %v", err)
	}
	ctx, err := DecodeHarmonyContext(side.ExtraInfo)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to decode context, err: %v", err)
	}
	err = ctx.Init()
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to get network config and shard schedule: %v", err)
	}

	for idx, headerBytes := range params.Headers {
		// Deserialize gensis header
		header, err := DecodeHeaderWithSig(headerBytes)
		if err != nil {
			return fmt.Errorf("HarmonyHandler failed to deserialize harmony header, idx %v, err: %v", idx, err)
		}

		// Extract shard epoch
		epoch, err := header.ExtractEpoch()
		if err != nil {
			return fmt.Errorf("HarmonyHandler, failed to extract Epoch from header, err: %v", err)
		}

		curEpoch, err := GetEpoch(native, params.ChainID)
		if err != nil {
			return fmt.Errorf("HarmonyHandler, failed to get current epoch info, err: %v", err)
		}
		if curEpoch == nil {
			return fmt.Errorf("HarmonyHandler failed to get current epoch info, idx %v, err: %v", idx, err)
		}

		// Valdiate next epoch with context
		err = ctx.VerifyNextEpoch(int(curEpoch.ShardID), epoch)
		if err != nil {
			return fmt.Errorf("HarmonyHandler failed to validate next epoch with context, idx %v block %s, err: %v",
				idx, header.Header.Number(), err)
		}

		// Validate next epoch with current epoch
		err = curEpoch.ValidateNextEpoch(ctx, header)
		if err != nil {
			return fmt.Errorf("HarmonyHandler failed to validate next epoch, idx %v block %s, err: %v",
				idx, header.Header.Number(), err)
		}

		err = curEpoch.VerifyHeaderSig(ctx, header)
		if err != nil {
			return fmt.Errorf("HarmonyHandler, failed to verify header with signature, err: %v, idx %v block %s",
				err, idx, header.Header.Number())
		}

		// Save new epoch
		err = storeEpoch(native, params.ChainID, epoch)
		if err != nil {
			return fmt.Errorf("HarmonyHandler, failed to store epoch info, err: %v", err)
		}

		// Update state
		err = updateWithHeader(native, params.ChainID, header.Header)
		if err != nil {
			return fmt.Errorf("HarmonyHandler failed to update state, err: %v", err)
		}
	}

	return
}

// SyncCrossChainMsg ...
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
