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

package starcoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/params"
	stc "github.com/starcoinorg/starcoin-go/client"
	"github.com/starcoinorg/starcoin-go/types"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

var NETURLMAP = make(map[uint64]string)

func init() {
	NETURLMAP[254] = "http://localhost:9850"
	NETURLMAP[251] = "https://barnard-seed.starcoin.org"
	NETURLMAP[252] = "https://proxima-seed.starcoin.org"
	NETURLMAP[253] = "https://halley-seed.starcoin.org"
	NETURLMAP[1] = "https://main-seed.starcoin.org"
}

func findNetwork(chainId uint64) (string, error) {
	if url, found := NETURLMAP[chainId]; found {
		return url, nil
	} else {
		return "", fmt.Errorf("cant't found url by chainid %d", chainId)
	}
}

// Handler ...
type Handler struct {
}

// NewSTCHandler ...
func NewSTCHandler() *Handler {
	return &Handler{}
}

// SyncGenesisHeader ...
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return errors.Errorf("StarcoinHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return errors.Errorf("StarcoinHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return errors.Errorf("StarcoinHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	header, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("StarcoinHandler SyncGenesisHeader: %s", err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)))
	if err != nil {
		return errors.Errorf("STCHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return errors.Errorf("STCHandler GetHeaderByHeight, genesis header had been initialized")
	}

	if err != nil {
		return errors.WithStack(err)
	}
	err = putGenesisBlockHeader(native, header, params.ChainID)
	if err != nil {
		return fmt.Errorf("STCHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	return nil
}

func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return errors.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}

	//caches := NewCaches(3, native)
	for _, v := range headerParams.Headers {
		var header types.BlockHeader
		err := json.Unmarshal(v, &header)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}

		headerHash, err := header.GetHash()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get header hash err: %v", err)
		}

		exist, err := IsHeaderExist(native, *headerHash, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, check header exist err: %v", err)
		}
		if exist == true {
			log.Warnf("SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}
		// get pre header
		parentHeader, err := GetHeaderByHash(native, header.ParentHash, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the parent block failed. Error:%s, header: %s", err, string(v))
		}
		parentHeaderHash, err := parentHeader.GetHash()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the parent block header hash failed. Error:%s, header: %s", err, string(v))
		}
		/**
		this code source refer to https://github.com/ethereum/go-ethereum/blob/master/consensus/ethash/consensus.go
		verify header need to verify:
		1. parent hash
		2. extra size
		3. current time
		*/
		//verify whether parent hash validity
		if !bytes.Equal(*parentHeaderHash, header.ParentHash) {
			return errors.Errorf("SyncBlockHeader, parent header is not right. Header: %s", string(v))
		}
		//verify whether extra size validity
		if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
			return errors.Errorf("SyncBlockHeader, SyncBlockHeader extra-data too long: %d > %d, header: %s", len(header.Extra), params.MaximumExtraDataSize, string(v))
		}
		//verify current time validity
		if header.Timestamp > uint64(time.Now().Add(allowedFutureBlockTime).Unix()) {
			return errors.Errorf("SyncBlockHeader,  verify header time error: checktime: %d, header: %s", time.Now().Add(allowedFutureBlockTime).Unix(), string(v))
		}
		//verify whether current header time and prevent header time validity
		if header.Timestamp >= parentHeader.Timestamp {
			return errors.Errorf("SyncBlockHeader, verify header time fail. Header: %s", string(v))
		}
		// Verify that the gas limit is <= 2^63-1
		cap := uint64(0x7fffffffffffffff)
		if header.GasUsed > cap {
			return errors.Errorf("SyncBlockHeader, invalid gasuseed: have %v, max %v, header: %s", header.GasUsed, cap, string(v))
		}

		//verify difficulty
		var expected *big.Int
		expected = difficultyCalculator(new(big.Int).SetUint64(header.Timestamp), parentHeader)
		if expected.Cmp(header.GetDiffculty()) != 0 {
			return errors.Errorf("SyncBlockHeader, invalid difficulty: have %v, want %v, header: %s", header.Difficulty, expected, string(v))
		}
		// verfify header
		err = h.verifyHeader(&header)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, verify header error: %v, header: %s", err, string(v))
		}
		//block header storage
		hederDifficultySum := new(big.Int).Add(header.GetDiffculty(), parentHeader.GetDiffculty())
		err = putBlockHeader(native, header, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncGenesisHeader, put blockHeader error: %v, header: %s", err, string(v))
		}
		// get current header of main
		currentHeader, err := GetCurrentHeader(native, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the current block failed. error:%s", err)
		}
		currentHeaderHash, err := currentHeader.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}
		if bytes.Equal(*currentHeaderHash, header.ParentHash) {
			appendHeader2Main(native, header.Number, *headerHash, headerParams.ChainID)
		} else {
			//
			if hederDifficultySum.Cmp(currentHeader.GetDiffculty()) > 0 {
				ReStructChain(native, currentHeader, &header, headerParams.ChainID)
			}
		}
	}
	//caches.deleteCaches()
	return nil
}

func getGenesisHeader(input []byte) (types.BlockHeader, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return types.BlockHeader{}, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var jsonHeader stc.BlockHeader
	err := json.Unmarshal(params.GenesisHeader, &jsonHeader)
	if err != nil {
		return types.BlockHeader{}, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}

	return jsonHeader.ToTypesHeader(), nil
}

func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func (h *Handler) verifyHeader(header *types.BlockHeader) error {
	return nil
}

func difficultyCalculator(time *big.Int, parent *types.BlockHeader) *big.Int {
	return nil
}
