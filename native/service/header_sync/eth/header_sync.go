package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	cty "github.com/ethereum/go-ethereum/core/types"
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
	params := new(scom.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("ETHHandler SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range params.Headers {
		var header cty.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("ETHHandler SyncBlockHeader, deserialize header err: %v", err)
		}
		prevHeader, err := GetHeaderByHeight(native, header.Number.Uint64()-1)
		if err != nil {
			return fmt.Errorf("ETHHandler SyncBlockHeader, height:%d, error:%s", header.Number.Uint64()-1, err)
		}

		if bytes.Equal(prevHeader.ParentHash.Bytes(), header.Hash().Bytes()) {
			return fmt.Errorf("current header height:%d, hash:%x, prevent header height:%d, parentHash:%x", header.Number, header.Hash(), prevHeader.Number, prevHeader.ParentHash)
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
