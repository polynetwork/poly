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

package header_sync

import (
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/bsc"
	"github.com/polynetwork/poly/native/service/header_sync/btc"
	"github.com/polynetwork/poly/native/service/header_sync/bytom"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/cosmos"
	"github.com/polynetwork/poly/native/service/header_sync/eth"
	"github.com/polynetwork/poly/native/service/header_sync/harmony"
	"github.com/polynetwork/poly/native/service/header_sync/heco"
	"github.com/polynetwork/poly/native/service/header_sync/hsc"
	"github.com/polynetwork/poly/native/service/header_sync/msc"
	"github.com/polynetwork/poly/native/service/header_sync/neo"
	"github.com/polynetwork/poly/native/service/header_sync/neo3"
	"github.com/polynetwork/poly/native/service/header_sync/neo3legacy"
	"github.com/polynetwork/poly/native/service/header_sync/okex"
	"github.com/polynetwork/poly/native/service/header_sync/ont"
	"github.com/polynetwork/poly/native/service/header_sync/pixiechain"
	"github.com/polynetwork/poly/native/service/header_sync/polygon"
	"github.com/polynetwork/poly/native/service/header_sync/quorum"
	"github.com/polynetwork/poly/native/service/header_sync/starcoin"
	"github.com/polynetwork/poly/native/service/header_sync/zilliqa"
	"github.com/polynetwork/poly/native/service/header_sync/zilliqalegacy"
	"github.com/polynetwork/poly/native/service/utils"
)

//Register methods of node_manager contract
func RegisterHeaderSyncContract(native *native.NativeService) {
	native.Register(hscommon.SYNC_GENESIS_HEADER, SyncGenesisHeader)
	native.Register(hscommon.SYNC_BLOCK_HEADER, SyncBlockHeader)
	native.Register(hscommon.SYNC_CROSS_CHAIN_MSG, SyncCrossChainMsg)
}

func GetChainHandler(router uint64) (hscommon.HeaderSyncHandler, error) {
	switch router {
	case utils.BTC_ROUTER:
		return btc.NewBTCHandler(), nil
	case utils.ETH_ROUTER:
		return eth.NewETHHandler(), nil
	case utils.ONT_ROUTER:
		return ont.NewONTHandler(), nil
	case utils.NEO_ROUTER:
		return neo.NewNEOHandler(), nil
	case utils.NEO3_ROUTER:
		return neo3.NewNeo3Handler(), nil
	case utils.NEO3_LEGACY_ROUTER:
		return neo3legacy.NewNeo3Handler(), nil
	case utils.COSMOS_ROUTER:
		return cosmos.NewCosmosHandler(), nil
	case utils.QUORUM_ROUTER:
		return quorum.NewQuorumHandler(), nil
	case utils.BSC_ROUTER:
		return bsc.NewHandler(), nil
	case utils.HECO_ROUTER:
		return heco.NewHecoHandler(), nil
	case utils.ZILLIQA_LEGACY_ROUTER:
		return zilliqalegacy.NewHandler(), nil
	case utils.ZILLIQA_ROUTER:
		return zilliqa.NewHandler(), nil
	case utils.MSC_ROUTER:
		return msc.NewHandler(), nil
	case utils.OKEX_ROUTER:
		return okex.NewHandler(), nil
	case utils.POLYGON_HEIMDALL_ROUTER:
		return polygon.NewHeimdallHandler(), nil
	case utils.POLYGON_BOR_ROUTER:
		return polygon.NewBorHandler(), nil
	case utils.PIXIECHAIN_ROUTER:
		return pixiechain.NewPixieHandler(), nil
	case utils.STARCOIN_ROUTER:
		return starcoin.NewSTCHandler(), nil
	case utils.HSC_ROUTER:
		return hsc.NewHscHandler(), nil
	case utils.HARMONY_ROUTER:
		return harmony.NewHandler(), nil
	case utils.BYTOM_ROUTER:
		return bytom.NewHandler(), nil
	default:
		return nil, fmt.Errorf("not a supported router:%d", router)
	}
}

func SyncGenesisHeader(native *native.NativeService) ([]byte, error) {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	chainID := params.ChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, side chain is not registered")
	}

	handler, err := GetChainHandler(sideChain.Router)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	err = utils.CheckRouterStartBlock(sideChain.Router, native.GetHeight())
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = handler.SyncGenesisHeader(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func SyncBlockHeader(native *native.NativeService) ([]byte, error) {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	chainID := params.ChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, side chain is not registered")
	}

	handler, err := GetChainHandler(sideChain.Router)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	err = utils.CheckRouterStartBlock(sideChain.Router, native.GetHeight())
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = handler.SyncBlockHeader(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func SyncCrossChainMsg(native *native.NativeService) ([]byte, error) {
	params := new(hscommon.SyncCrossChainMsgParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncCrossChainMsg, contract params deserialize error: %v", err)
	}
	chainID := params.ChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncCrossChainMsg, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncCrossChainMsg, side chain is not registered")
	}

	handler, err := GetChainHandler(sideChain.Router)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	err = utils.CheckRouterStartBlock(sideChain.Router, native.GetHeight())
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = handler.SyncCrossChainMsg(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}
