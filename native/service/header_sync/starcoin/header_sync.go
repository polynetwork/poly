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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	stc "github.com/starcoinorg/starcoin-go/client"
	"github.com/starcoinorg/starcoin-go/core/consensus"
	"github.com/starcoinorg/starcoin-go/types"

	"github.com/pkg/errors"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	// don't modify this, which trigger relayer rollback to common ancestor.
	GET_PARENT_BLOCK_FAILED_FORMAT = "SyncBlockHeader, get the parent block failed. Error:%s, header: %s"
)

var MAXU256 = &big.Int{}

func init() {
	MAXU256.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
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

	jsonHeaderAndInfo, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("StarcoinHandler SyncGenesisHeader: %s", err)
	}

	header, err := jsonHeaderAndInfo.BlockHeader.ToTypesHeader()
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)))
	if err != nil {
		return errors.Errorf("STCHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return errors.Errorf("STCHandler GetHeaderByHeight, genesis header had been initialized")
	}

	blockInfo, err := jsonHeaderAndInfo.BlockInfo.ToTypesBlockInfo()
	if err != nil {
		return errors.Errorf("SyncGenesisHeader, block info parse error: %v, header: %v", err, jsonHeaderAndInfo)
	}

	hdrAndInfo := types.BlockHeaderAndBlockInfo{
		BlockHeader: *header,
		BlockInfo:   *blockInfo,
	}
	err = putGenesisBlockHeader(native, hdrAndInfo, params.ChainID)
	if err != nil {
		return fmt.Errorf("STCHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	// err = putBlockInfo(native, *blockInfo, params.ChainID)
	// if err != nil {
	// 	return errors.Errorf("SyncGenesisHeader, put block info error: %v, header: %v", err, jsonHeaderAndInfo)
	// }

	return nil
}

func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return errors.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}

	//get genesis header
	genesisHeader, err := GetGenesisBlockHeader(native, headerParams.ChainID)
	if err != nil {
		return errors.Errorf("SyncBlockHeader,get genesis header error: %v", err)
	}

	for _, v := range headerParams.Headers {
		var jsonHeader stc.BlockHeaderWithDifficultyInfo
		err = json.Unmarshal(v, &jsonHeader)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}
		hdr, err := jsonHeader.BlockHeader.ToTypesHeader()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, to types.BlockHeader err: %v", err)
		}
		blkInfo, err := jsonHeader.BlockInfo.ToTypesBlockInfo()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, to types.BlockInfo err: %v", err)
		}
		header := types.BlockHeaderAndBlockInfo{
			BlockHeader: *hdr,
			BlockInfo:   *blkInfo,
		}
		headerHash, err := header.BlockHeader.GetHash()
		headerHeight := header.BlockHeader.Number
		timeTarget := jsonHeader.BlockTimeTarget
		difficultyWindow := jsonHeader.BlockDifficutyWindow
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get header hash err: %v", err)
		}

		exist, err := IsHeaderExist(native, *headerHash, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, check header exist err: %v", err)
		}
		if exist {
			log.Warnf("SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}
		// get pre header
		parentHeader, err := GetHeaderByHash(native, header.BlockHeader.ParentHash, headerParams.ChainID)
		if err != nil {
			return errors.Errorf(GET_PARENT_BLOCK_FAILED_FORMAT, err, string(v))
		}
		if header.BlockHeader.Number != parentHeader.BlockHeader.Number+1 {
			return errors.Errorf("SyncBlockHeader, the parent block number: %d, header number: %d", parentHeader.BlockHeader.Number, header.BlockHeader.Number)
		}
		if err := verifyTotalDifficulty(&header, parentHeader); err != nil {
			return err
		}
		parentHeaderHash, err := parentHeader.BlockHeader.GetHash()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the parent block header hash failed. Error:%s, header: %s", err, string(v))
		}

		//verify whether parent hash validity
		if !bytes.Equal(*parentHeaderHash, header.BlockHeader.ParentHash) {
			return errors.Errorf("SyncBlockHeader, parent header is not right. Header: %s", string(v))
		}
		//verify whether extra size validity
		if uint64(len(header.BlockHeader.Extra)) > params.MaximumExtraDataSize {
			return errors.Errorf("SyncBlockHeader, SyncBlockHeader extra-data too long: %d > %d, header: %s", len(header.BlockHeader.Extra), params.MaximumExtraDataSize, string(v))
		}
		//verify current time validity
		if header.BlockHeader.Timestamp > uint64(time.Now().Add(allowedFutureBlockTime).Unix()*1000) {
			return errors.Errorf("SyncBlockHeader,  verify header time error: checktime: %d, header: %s", time.Now().Add(allowedFutureBlockTime).Unix(), string(v))
		}
		//verify whether current header time and prevent header time validity
		if header.BlockHeader.Timestamp <= parentHeader.BlockHeader.Timestamp {
			return errors.Errorf("SyncBlockHeader, verify header time fail. parent: %d, Header: %s", parentHeader.BlockHeader.Timestamp, string(v))
		}
		// Verify that the gas limit is <= 2^63-1
		cap := uint64(0x7fffffffffffffff)
		if header.BlockHeader.GasUsed > cap {
			return errors.Errorf("SyncBlockHeader, invalid gasuseed: have %v, max %v, header: %s", header.BlockHeader.GasUsed, cap, string(v))
		}

		difficultyWindowU64 := uint64(difficultyWindow)
		if (headerHeight - genesisHeader.BlockHeader.Number) >= difficultyWindowU64 {
			//verify difficulty
			var expected *big.Int
			expected, err = difficultyCalculator(native, &header.BlockHeader, headerParams.ChainID, uint64(timeTarget), difficultyWindowU64)
			if err != nil {
				return errors.Errorf("difficulty calculator error: %v, header: %s", err, string(v))
			}
			if err := verifyHeaderDifficulty(expected, &header.BlockHeader); err != nil {
				return err
			}
		}

		//block header storage
		err = putBlockHeader(native, header, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncGenesisHeader, put blockHeader error: %v, header: %s", err, string(v))
		}

		// get current header of main
		currentHeader, err := GetCurrentHeader(native, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the current block failed. error:%s", err)
		}
		currentHeaderHash, err := currentHeader.BlockHeader.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}

		if bytes.Equal(*currentHeaderHash, header.BlockHeader.ParentHash) {
			err := appendHeader2Main(native, header.BlockHeader.Number, *headerHash, headerParams.ChainID)
			_ = err //todo ignore error?
			// if err != nil {
			// 	log.Warnf("SyncBlockHeader, appendHeader2Main error:%s", err.Error())
			// }
		} else {
			// //get current total difficulty
			// blockInfo, err := getBlockInfo(native, *currentHeaderHash, headerParams.ChainID)
			// if err != nil {
			// 	return errors.Errorf("get block info error, hash:%x  error:%s", currentHeaderHash, err)
			// }
			currentTotalDifficulty := new(uint256.Int).SetBytes(currentHeader.BlockInfo.TotalDifficulty[:])
			// //get fork header parent total difficulty
			// parentBlockInfo, err := getBlockInfo(native, header.ParentHash, headerParams.ChainID)
			// if err != nil {
			// 	return errors.Errorf("get parent block info error, hash:%x  error:%s", currentHeaderHash, err)
			// }
			//parentBlockInfo := parentHeader.BlockInfo
			//parentTotalDifficulty := new(uint256.Int).SetBytes(parentBlockInfo.TotalDifficulty[:])
			//
			// -------- eth handlder: --------
			// if hederDifficultySum.Cmp(currentDifficultySum) > 0 { ...
			//
			headerDifficulty := new(uint256.Int).SetBytes(header.BlockHeader.Difficulty[:])
			parentTotalDifficulty := new(uint256.Int).SetBytes(parentHeader.BlockInfo.TotalDifficulty[:])
			if new(uint256.Int).Add(parentTotalDifficulty, headerDifficulty).Cmp(currentTotalDifficulty) > 0 {
				err := ReStructChain(native, currentHeader, &header, headerParams.ChainID)
				_ = err //todo ignore error?
				// if err != nil {
				// 	log.Warnf("SyncBlockHeader, ReStructChain error:%s", err.Error())
				// 	return err
				// }
			}
		}
	}
	return nil
}

type starcoinConsensus interface {
	VerifyHeaderDifficulty(difficulty uint256.Int, headerDifficulty uint256.Int, headerBlob []byte, nonce uint32, extra []byte) (bool, error)
}

type defaultConsensus struct {
}

func (c defaultConsensus) VerifyHeaderDifficulty(difficulty uint256.Int, headerDifficulty uint256.Int, headerBlob []byte, nonce uint32, extra []byte) (bool, error) {
	return consensus.VerifyHeaderDifficulty(difficulty, headerDifficulty, headerBlob, nonce, extra)
}

var cryptonightConsensus starcoinConsensus = defaultConsensus{}
var argonConsensus starcoinConsensus = consensus.ArgonConsensus{}

func verifyHeaderDifficulty(expected *big.Int, header *types.BlockHeader) error {
	// if expected.Cmp(header.BlockHeader.GetDiffculty()) != 0 {
	// 	return errors.Errorf("SyncBlockHeader, invalid difficulty: have %v, want %v, header: %s", header.BlockHeader.Difficulty, expected, string(v))
	// }
	e := new(uint256.Int)
	overflow := e.SetFromBig(expected)
	if overflow {
		return errors.Errorf("verifyHeaderDifficulty, SetFromBig overflow: %d", expected)
	}
	hd := new(uint256.Int).SetBytes(header.Difficulty[:])
	hb, err := header.ToHeaderBlob()
	if err != nil {
		return errors.Errorf("verifyHeaderDifficulty, header.ToHeaderBlob error: %v", header)
	}
	// //////////////////////////////////////////////////////////////////////////////////////////////
	chainID := config.GetChainIdByNetId(config.DefConfig.P2PNode.NetworkId)
	//_ = chainID
	var ok bool
	var useArgonConsensus = header.Number >= 5061625 && config.TESTNET_CHAIN_ID == chainID
	if useArgonConsensus {
		ok, err = argonConsensus.VerifyHeaderDifficulty(*e, *hd, hb, header.Nonce, header.Extra[:])
		if err != nil {
			return errors.Errorf("verifyHeaderDifficulty, argonConsensus.VerifyHeaderDifficulty error. Header.number: %v, expectedDifficulty: %v, header: %v, error: %v", header.Number, expected, header, err)
		}
	} else {
		ok, err = cryptonightConsensus.VerifyHeaderDifficulty(*e, *hd, hb, header.Nonce, header.Extra[:])
		if err != nil {
			return errors.Errorf("verifyHeaderDifficulty, cryptonightConsensus.VerifyHeaderDifficulty error. Header.number: %v, expectedDifficulty: %v, header: %v, error: %v", header.Number, expected, header, err)
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	if !ok {
		return errors.Errorf("verifyHeaderDifficulty, consensus.VerifyHeaderDifficulty failed: %v", header)
	}
	return nil
}

func getGenesisHeader(input []byte) (stc.BlockHeaderAndBlockInfo, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return stc.BlockHeaderAndBlockInfo{}, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var jsonHeaderAndInfo stc.BlockHeaderAndBlockInfo
	err := json.Unmarshal(params.GenesisHeader, &jsonHeaderAndInfo)
	if err != nil {
		return stc.BlockHeaderAndBlockInfo{}, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}
	return jsonHeaderAndInfo, err
}

func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func difficultyCalculator(native *native.NativeService, blockHeader *types.BlockHeader, chainId uint64, timeTarget uint64, difficultyWindow uint64) (*big.Int, error) {
	var currentHeight uint64 = blockHeader.Number
	//get last difficulty
	var lastDifficulties []BlockDiffInfo
	hash := blockHeader.ParentHash
	for height := currentHeight - 1; height > currentHeight-difficultyWindow-1; height-- {
		//hdr, err := GetHeaderByHeight(native, height, chainId)
		hdr, err := GetHeaderByHash(native, hash, chainId)
		if err != nil {
			return nil, fmt.Errorf("difficultyCalculator, get header error: %s, height: %d, hash: %s", err.Error(), height, hex.EncodeToString(hash))
		}
		target := targetToDiff(new(uint256.Int).SetBytes(hdr.BlockHeader.Difficulty[:]))
		lastDifficulties = append(lastDifficulties, BlockDiffInfo{hdr.BlockHeader.Timestamp, *target})
		hash = hdr.BlockHeader.ParentHash
	}
	nextTarget, err := getNextTarget(lastDifficulties, timeTarget)
	return targetToDiff(&nextTarget).ToBig(), err
}

func getNextTarget(blocks []BlockDiffInfo, timePlan uint64) (uint256.Int, error) {
	//todo online to debug level
	log.Infof("get next target: %v, time plan: %d", blocks, timePlan)
	nextTarget := uint256.NewInt(0)
	length := len(blocks)
	if length < 1 {
		return *nextTarget, fmt.Errorf("get next target blocks is null.")
	}
	if length == 1 {
		return blocks[0].Target, nil
	}

	totalTarget := new(big.Int)
	for _, block := range blocks {
		totalTarget.Add(totalTarget, block.Target.ToBig())
	}

	totalTargetU256, overflow := uint256.FromBig(totalTarget)
	if overflow {
		return *nextTarget, fmt.Errorf("get next target, total target overflow: %d.", totalTarget)
	}

	lengthU256 := new(uint256.Int).SetUint64(uint64(length))
	avgTarget := new(uint256.Int).Div(totalTargetU256, lengthU256)

	var avgTime uint64
	if length == 2 {
		avgTime = blocks[0].Timestamp - blocks[1].Timestamp
	}
	if length > 2 {
		latestTimestamp := blocks[0].Timestamp
		totalBlockTime := uint64(0)
		vblocks := 0
		for idx, block := range blocks {
			if idx == 0 {
				continue
			}
			totalBlockTime = totalBlockTime + (latestTimestamp - block.Timestamp)
			vblocks = vblocks + idx
		}
		avgTime = totalBlockTime / uint64(vblocks)
	}
	if avgTime == 0 {
		avgTime = 1
	}

	timePlanU256 := new(uint256.Int).SetUint64(timePlan)
	avgTimeU256 := new(uint256.Int).SetUint64(avgTime)
	nextTarget.Div(avgTarget, timePlanU256)
	_, overflow = nextTarget.MulOverflow(nextTarget, avgTimeU256)
	if overflow {
		return *nextTarget, fmt.Errorf("get next target, next target overflow: avgTimeU256: %d, nextTarget: %d, avgTimeU256: %d .", avgTimeU256, nextTarget, avgTimeU256)
	}
	tempNextTarget := nextTarget.Clone()
	tempNumber := new(uint256.Int).SetUint64(2)
	tempNextTarget.Div(tempNextTarget, tempNumber)
	tempAvgTarget := new(uint256.Int).Div(avgTarget, tempNumber)
	if tempNextTarget.Gt(avgTarget) {
		nextTarget.Mul(avgTarget, tempNumber)
	} else if nextTarget.Lt(tempAvgTarget) {
		nextTarget = tempAvgTarget.Clone()
	}
	return *nextTarget, nil
}

func targetToDiff(target *uint256.Int) *uint256.Int {
	bigint := &big.Int{}
	diff, _ := uint256.FromBig(bigint.Div(MAXU256, target.ToBig()))
	return diff
}
