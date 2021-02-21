/*
 * Copyright (C) 2021 The poly network Authors
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

package zilliqa

import (
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	verifier2 "github.com/Zilliqa/gozilliqa-sdk/verifier"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// SyncGenesisHeader ...
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("ZILHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("ZILHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("ZILHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	txBlockAndDsComm, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("ZILHandler SyncGenesisHeader: %s", err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)))
	if err != nil {
		return fmt.Errorf("ZILHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return fmt.Errorf("ZILHandler GetHeaderByHeight, genesis header had been initialized")
	}
	err = putGenesisBlockHeader(native, txBlockAndDsComm, params.ChainID)
	if err != nil {
		return fmt.Errorf("ZILHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	return
}

// SyncBlockHeader ...
func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}

	for _, v := range headerParams.Headers {
		var txBlockAndDsComm core.TxBlockAndDsComm
		err := json.Unmarshal(v, &txBlockAndDsComm)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}

		txBlock := txBlockAndDsComm.Block
		dsComm := txBlockAndDsComm.DsComm
		blockHash := txBlock.BlockHash
		exist, err := IsHeaderExist(native, blockHash[:], headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, check header exist err: %v", err)
		}
		if exist == true {
			log.Warnf("SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}
		preHash := txBlock.BlockHeader.BlockHeaderBase.PrevHash
		_, err2 := GetHeaderByHash(native, preHash[:], headerParams.ChainID)
		if err2 != nil {
			return fmt.Errorf("SyncBlockHeader, get the parent block failed. Error:%s, header: %s", err2, string(v))
		}

		verifier := &verifier2.Verifier{}
		err3 := verifier.VerifyTxBlock(txBlock, dsComm)
		if err3 != nil {
			return fmt.Errorf("SyncBlockHeader, verify block failed. Error:%s, header: %s", err3, string(v))
		}

		err4 := putBlockHeader(native, &txBlockAndDsComm, headerParams.ChainID)
		if err4 != nil {
			return fmt.Errorf("SyncBlockHeader, put blockHeader failed. Error:%s, header: %s", err3, string(v))
		}
	}

	return nil
}

func getGenesisHeader(input []byte) (core.TxBlockAndDsComm, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return core.TxBlockAndDsComm{}, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var txBlockAndDsComm core.TxBlockAndDsComm
	err := json.Unmarshal(params.GenesisHeader, &txBlockAndDsComm)
	if err != nil {
		return core.TxBlockAndDsComm{}, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}
	return txBlockAndDsComm, nil
}
