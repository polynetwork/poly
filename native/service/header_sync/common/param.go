package common

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/utils"
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

func NotifyPutHeader(native *native.NativeService, chainID uint64, height uint32, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, blockHash, native.GetHeight()},
		})
}

func NotifyPutCrossChainMsg(native *native.NativeService, chainID uint64, height uint32) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, native.GetHeight()},
		})
}
