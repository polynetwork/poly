package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/consensus"
	"time"

	cty "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) SyncGenesisHeader(native *native.NativeService) error {
	//// get operator from database
	//operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	//if err != nil {
	//	return err
	//}

	////check witness
	//err = utils.ValidateOwner(native, operatorAddress)
	//if err != nil {
	//	return fmt.Errorf("ETHHandler SyncGenesisHeader, checkWitness error: %v", err)
	//}

	header, headerByte, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader: %s", err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.HEADER_INDEX), utils.ETH_CHAIN_ID_BYTE, header.Number.Bytes()))
	if err != nil {
		return fmt.Errorf("ETHHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return fmt.Errorf("ETHHandler GetHeaderByHeight, genesis header had been initialized")
	}

	//block header storage
	err = putBlockHeader(native, header, headerByte)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	return nil
}

func (this *ETHHandler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("ETHHandler SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range headerParams.Headers {
		var header cty.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("ETHHandler SyncBlockHeader, deserialize header err: %v", err)
		}
		prevHeader, err := GetHeaderByHeight(native, header.Number.Uint64()-1)
		if err != nil {
			return fmt.Errorf("ETHHandler SyncBlockHeader, height:%d, error:%s", header.Number.Uint64()-1, err)
		}
		/**
		this code source refer to https://github.com/ethereum/go-ethereum/blob/master/consensus/ethash/consensus.go
		verify header need to verify:
		1. parent hash
		2. extra size
		3. current time
		*/
		//verify whether parent hash validity
		if !bytes.Equal(prevHeader.Hash().Bytes(), header.ParentHash.Bytes()) {
			return fmt.Errorf("current header height:%d, hash:%x, prevent header height:%d, parentHash:%x", header.Number, header.Hash(), prevHeader.Number, prevHeader.ParentHash)
		}
		//verify whether extra size validity
		if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
			return fmt.Errorf("SyncBlockHeader extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
		}
		//verify current time validity
		if header.Time > uint64(time.Now().Add(allowedFutureBlockTime).Unix()) {
			return fmt.Errorf("SyncBlockHeader, verify header time error:%s", consensus.ErrFutureBlock)
		}
		//verify whether current header time and prevent header time validity
		if header.Time <= prevHeader.Time {
			return fmt.Errorf("SyncBlockHeader, verify header time fail, current header time:%d, prevent header time:%d", header.Time, prevHeader.Time)
		}

		// Verify that the gas limit is <= 2^63-1
		cap := uint64(0x7fffffffffffffff)
		if header.GasLimit > cap {
			return fmt.Errorf("SyncBlockHeader invalid gasLimit: have %v, max %v", header.GasLimit, cap)
		}
		// Verify that the gasUsed is <= gasLimit
		if header.GasUsed > header.GasLimit {
			return fmt.Errorf("SyncBlockHeader invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
		}

		//block header storage
		err = putBlockHeader(native, header, v)
		if err != nil {
			return fmt.Errorf("ETHHandler SyncGenesisHeader, put blockHeader error: %v", err)
		}
	}
	return nil
}

func getGenesisHeader(input []byte) (cty.Header, []byte, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return cty.Header{}, nil, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var header cty.Header
	err := json.Unmarshal(params.GenesisHeader, &header)
	if err != nil {
		return cty.Header{}, nil, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}
	return header, params.GenesisHeader, nil
}
