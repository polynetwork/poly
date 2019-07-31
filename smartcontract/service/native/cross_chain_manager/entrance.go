package cross_chain_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"

	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/btc"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/eth"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/ont"
)

const (
	ImportExTransfer_Name   = "ImportOuterTransfer"
	RegisterChainHandler_ID = "RegisterChainHandler"
)

type CrossChainHandler func(native *native.NativeService) ([]byte, error)

//var mapping = make(map[uint64]CrossChainHandler)

func InitEntrance() {
	native.Contracts[utils.CrossChainManagerContractAddress] = RegisterCrossChianManagerContract
	//RegisterChainHandler(0,BTCHandler)
	//RegisterChainHandler(1,ETHHandler)
}

func RegisterCrossChianManagerContract(native *native.NativeService) {
	native.Register(ImportExTransfer_Name, ImportExTransfer)
}

//func RegisterChainHandler(chainid uint64, handler CrossChainHandler) {
//	mapping[chainid] = handler
//}

func GetChainHandler(chainid uint64) (inf.ChainHandler, error) {
	//handler, ok := mapping[chainid]
	//if !ok {
	//	return nil, fmt.Errorf("no handler for chainID:%d", chainid)
	//}

	switch chainid {
	case 0:
		return btc.NewBTCHandler(), nil
	case 1:
		return eth.NewETHHandler(), nil
	case 2:
		return ont.NewONTHandler(), nil
	default:
		return nil, fmt.Errorf("not a supported chainid:%d", chainid)
	}
}

func ImportExTransfer(native *native.NativeService) ([]byte, error) {
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, contract params deserialize error: %v", err)
	}

	chainid := params.SourceChainID
	handler, err := GetChainHandler(chainid)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//1. verify tx
	txParam, err := handler.Verify(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//2. make target chain tx
	targetid := txParam.ToChainID
	targetHandler, err := GetChainHandler(targetid)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//NOTE, you need to store the tx in this
	err = targetHandler.MakeTransaction(native, txParam)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}
