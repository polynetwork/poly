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

package common

import (
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//key prefix
	CROSS_CHAIN_MSG             = "crossChainMsg"
	CURRENT_MSG_HEIGHT          = "currentMsgHeight"
	BLOCK_HEADER                = "blockHeader"
	CURRENT_HEADER_HEIGHT       = "currentHeaderHeight"
	HEADER_INDEX                = "headerIndex"
	CONSENSUS_PEER              = "consensusPeer"
	CONSENSUS_PEER_BLOCK_HEIGHT = "consensusPeerBlockHeight"
	KEY_HEIGHTS                 = "keyHeights"
	ETH_CACHE                   = "ethCaches"
	GENESIS_HEADER              = "genesisHeader"
	MAIN_CHAIN                  = "mainChain"
	EPOCH_SWITCH                = "epochSwitch"
	SYNC_HEADER_NAME            = "syncHeader"
	SYNC_CROSSCHAIN_MSG         = "syncCrossChainMsg"
	POLYGON_SPAN                = "polygonSpan"
)

type HeaderSyncHandler interface {
	SyncGenesisHeader(service *native.NativeService) error
	SyncBlockHeader(service *native.NativeService) error
	SyncCrossChainMsg(service *native.NativeService) error
}

type SyncGenesisHeaderParam struct {
	ChainID       uint64
	GenesisHeader []byte
}

func (this *SyncGenesisHeaderParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteVarBytes(this.GenesisHeader)
}

func (this *SyncGenesisHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	genesisHeader, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisHeader count error")
	}
	this.ChainID = chainID
	this.GenesisHeader = genesisHeader
	return nil
}

type SyncBlockHeaderParam struct {
	ChainID uint64
	Address common.Address
	Headers [][]byte
}

func (this *SyncBlockHeaderParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteAddress(this.Address)
	sink.WriteUint64(uint64(len(this.Headers)))
	for _, v := range this.Headers {
		sink.WriteVarBytes(v)
	}
}

func (this *SyncBlockHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	address, eof := source.NextAddress()
	if eof {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error")
	}
	var headers [][]byte
	for i := 0; uint64(i) < n; i++ {
		header, eof := source.NextVarBytes()
		if eof {

			return fmt.Errorf("utils.DecodeVarBytes, deserialize header error")
		}
		headers = append(headers, header)
	}
	this.ChainID = chainID
	this.Address = address
	this.Headers = headers
	return nil
}

type SyncCrossChainMsgParam struct {
	ChainID        uint64
	Address        common.Address
	CrossChainMsgs [][]byte
}

func (this *SyncCrossChainMsgParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteAddress(this.Address)
	sink.WriteUint64(uint64(len(this.CrossChainMsgs)))
	for _, v := range this.CrossChainMsgs {
		sink.WriteVarBytes(v)
	}
}

func (this *SyncCrossChainMsgParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	address, eof := source.NextAddress()
	if eof {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error")
	}
	var crossChainMsgs [][]byte
	for i := 0; uint64(i) < n; i++ {
		crossChainMsg, eof := source.NextVarBytes()
		if eof {

			return fmt.Errorf("utils.DecodeVarBytes, deserialize crossChainMsg error")
		}
		crossChainMsgs = append(crossChainMsgs, crossChainMsg)
	}
	this.ChainID = chainID
	this.Address = address
	this.CrossChainMsgs = crossChainMsgs
	return nil
}

func NotifyPutHeader(native *native.NativeService, chainID uint64, height uint64, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{SYNC_HEADER_NAME, chainID, height, blockHash, native.GetHeight()},
		})
}

func NotifyPutCrossChainMsg(native *native.NativeService, chainID uint64, height uint32) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{SYNC_CROSSCHAIN_MSG, chainID, height, native.GetHeight()},
		})
}
