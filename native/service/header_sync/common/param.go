package common

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	//key prefix
	BLOCK_HEADER                = "blockHeader"
	CURRENT_HEIGHT              = "currentHeight"
	HEADER_INDEX                = "headerIndex"
	CONSENSUS_PEER              = "consensusPeer"
	CONSENSUS_PEER_BLOCK_HEIGHT = "consensusPeerBlockHeight"
	KEY_HEIGHTS                 = "keyHeights"
)

type HeaderSyncHandler interface {
	SyncGenesisHeader(service *native.NativeService) error
	SyncBlockHeader(service *native.NativeService) error
}

type SyncGenesisHeaderParam struct {
	ChainID       uint64
	GenesisHeader []byte
}

func (this *SyncGenesisHeaderParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.ChainID)
	sink.WriteVarBytes(this.GenesisHeader)
	return nil
}

func (this *SyncGenesisHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	genesisHeader, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisHeader count error:%s", err)
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
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error:%s", err)
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error")
	}
	var headers [][]byte
	for i := 0; uint64(i) < n; i++ {
		header, err := utils.DecodeVarBytes(source)
		if err != nil {
			return fmt.Errorf("utils.DecodeVarBytes, deserialize header error: %v", err)
		}
		headers = append(headers, header)
	}
	this.ChainID = chainID
	this.Address = address
	this.Headers = headers
	return nil
}