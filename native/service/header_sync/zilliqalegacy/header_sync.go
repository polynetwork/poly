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

package zilliqalegacy

import (
	"encoding/json"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/renlulu/gozilliqa-sdklegacy/core"
	"github.com/renlulu/gozilliqa-sdklegacy/util"
	verifier2 "github.com/renlulu/gozilliqa-sdklegacy/verifier"
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

	side, err := side_chain_manager.GetSideChain(native, headerParams.ChainID)
	if err != nil {
		return fmt.Errorf("zil Handler SyncBlockHeader, GetSideChain error: %v", err)
	}

	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("zil Handler SyncBlockHeader, ExtraInfo Unmarshal error: %v", err)
	}

	verifier := &verifier2.Verifier{
		NumOfDsGuard: extraInfo.NumOfGuardList,
	}

	// ...txblock1-1,txblock1-2...dsblock2,txblock2-1,txblock2-2...
	for _, v := range headerParams.Headers {
		var txBlockAndDsComm core.TxBlockOrDsBlock
		err := json.Unmarshal(v, &txBlockAndDsComm)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}

		txBlock := txBlockAndDsComm.TxBlock
		dsBlock := txBlockAndDsComm.DsBlock

		if dsBlock != nil {
			// if ds block is not nil, we need to verify itself, then update DsComm list

			// 1. if ds block exist already
			blockHash := dsBlock.BlockHash
			exist, err := IsHeaderExist(native, blockHash[:], headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncDsBlockHeader, check header exist err: %v", err)
			}
			if exist == true {
				log.Warnf("SyncDsBlockHeader, header has exist. Header: %s", string(v))
				continue
			}

			// 2. check parent block
			preHash := util.DecodeHex(dsBlock.PrevDSHash)
			_, err = GetDsHeaderByHash(native, preHash[:], headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncDsBlockHeader, get the parent block failed. parent hash is: %s, Error:%s, header: %s", dsBlock.PrevDSHash, err, string(v))
			}

			// 3. get old ds comm list
			dsBlockNum := dsBlock.BlockHeader.BlockNum
			dscomm, err := getDsComm(native, dsBlockNum-1, headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncDsBlockHeader, get dscomm err: %v", err)
			}
			dsList := dsCommListFromArray(dscomm)

			// 4. verify ds block, generate new ds comm list
			newDsList, err2 := verifier.VerifyDsBlock(dsBlock, dsList)
			if err2 != nil {
				return fmt.Errorf("SyncDsBlockHeader, verify ds block err: %v", err2)
			}

			// 5. update ds comm list, put ds block
			putDsComm(native, dsBlockNum, dsCommArrayFromList(newDsList), headerParams.ChainID)
			err = putDsBlockHeader(native, dsBlock, headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncDsBlockHeader, put blockHeader failed. Error:%s, header: %s", err, string(v))
			}
		}

		if txBlock != nil {
			// 1. if tx block exist
			blockHash := txBlock.BlockHash
			exist, err := IsHeaderExist(native, blockHash[:], headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncTxBlockHeader, check header exist err: %v", err)
			}
			if exist == true {
				log.Warnf("SyncTxBlockHeader, header has exist. Header: %s", string(v))
				continue
			}

			// 2. check parent tx block
			preHash := txBlock.BlockHeader.BlockHeaderBase.PrevHash
			_, err = GetTxHeaderByHash(native, preHash[:], headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncTxBlockHeader, get the parent block failed. Error:%s, header: %s", err, string(v))
			}

			// 3. get comm list
			dscomm, err := getDsComm(native, txBlock.BlockHeader.DSBlockNum, headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncTxBlockHeader, get dscomm for tx block err: %s", err.Error())
			}

			// 4. verify tx block and store it
			err = verifier.VerifyTxBlock(txBlock, dsCommListFromArray(dscomm))
			if err != nil {
				return fmt.Errorf("SyncTxBlockHeader, verify block failed. Error:%s, header: %s", err, string(v))
			}

			err = putTxBlockHeader(native, txBlock, headerParams.ChainID)
			if err != nil {
				return fmt.Errorf("SyncTxBlockHeader, put blockHeader failed. Error:%s, header: %s", err, string(v))
			}

			// 5. update header of main
			AppendHeader2Main(native, txBlock.BlockHeader.BlockNum, txBlock.BlockHash[:], headerParams.ChainID)
		}
	}

	return nil
}

// SyncCrossChainMsg ...
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

type TxBlockAndDsComm struct {
	TxBlock *core.TxBlock
	DsBlock *core.DsBlock
	DsComm  []core.PairOfNode
}

func getGenesisHeader(input []byte) (TxBlockAndDsComm, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return TxBlockAndDsComm{}, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var txBlockAndDsComm TxBlockAndDsComm
	err := json.Unmarshal(params.GenesisHeader, &txBlockAndDsComm)
	if err != nil {
		return TxBlockAndDsComm{}, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}
	return txBlockAndDsComm, nil
}

// ExtraInfo ...
type ExtraInfo struct {
	NumOfGuardList int // for zilliqalegacy
}
