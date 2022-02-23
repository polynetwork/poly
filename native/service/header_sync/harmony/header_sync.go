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
	"encoding/json"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
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
	if err != nil { return }
	if headerExist != nil {
		return fmt.Errorf("HarmonyHandler genesis header was already set")
	}

	// Deserialize gensis header
	header := &HeaderWithSig{}
	err = json.Unmarshal(params.GenesisHeader, &header)
	if err != nil {
		return fmt.Errorf("HarmonyHandler failed to deserialize harmony header, err: %v", err)
	}

	// Extract shard epoch
	epoch, err := header.ExtractEpoch()
	if err != nil {
		return
	}

	if !IsLastEpochBlock(header.Header.Number()) {
		return fmt.Errorf("HarmonyHandler header block %s is not the last in epoch", header.Header.Number())
	}

	// Store genesis header
	err = storeGenesisHeader(native, params.ChainID, header)
	if err != nil { return }

	// Store epoch info
	err = storeEpoch(native, params.ChainID, epoch)
	return
}

// Sync block header
func (h *Handler) SyncBlockHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncBlockHeaderParam)
	err = params.Deserialization(common.NewZeroCopySource(native.GetInput()))
	if err != nil {
		return fmt.Errorf("%w, HarmonyHandler failed to deserialize headers params", err)
	}

	/*
	side, err := side_chain_manager.GetSideChain(native, params.ChainID)
	if err != nil || side == nil {
		return fmt.Errorf("HarmonyHandler failed to get side chain error: %v", err)
	}
	 */

	for idx, headerBytes := range params.Headers {
		// Deserialize gensis header
		header := &HeaderWithSig{}
		err = json.Unmarshal(headerBytes, &header)
		if err != nil {
			return fmt.Errorf("HarmonyHandler failed to deserialize harmony header, idx %v, err: %v", idx, err)
		}

		// Extract shard epoch
		epoch, err := header.ExtractEpoch()
		if err != nil {
			return err
		}

		curEpoch, err := getEpoch(native, params.ChainID)
		if err != nil { return err }
		if curEpoch == nil {
			return fmt.Errorf("HarmonyHandler failed to get current epoch info, idx %v, err: %v", idx, err)
		}

		err = curEpoch.ValidateNextEpoch(header)
		if err != nil {
			return fmt.Errorf("HarmonyHandler failed to validate next epoch, idx %v block %s, err: %v",
				idx, header.Header.Number(), err)
		}

		if !IsLastEpochBlock(header.Header.Number()) {
			return fmt.Errorf("HarmonyHandler header block %s is not the last in epoch, idx %v",
				header.Header.Number(), idx)
		}

		err = curEpoch.VerifyHeaderSig(header)
		if err != nil {
			return fmt.Errorf("HarmonyHandler, failed to verify header with signature, err: %v, idx %v block %s",
				err, idx, header.Header.Number())
		}

		storeEpoch(native, params.ChainID, epoch)
	}

	return
}

// SyncCrossChainMsg ...
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
