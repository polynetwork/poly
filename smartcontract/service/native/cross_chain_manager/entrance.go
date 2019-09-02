package cross_chain_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"

	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/btc"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/eth"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/ont"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
)

const (
	ImportExTransfer_Name = "ImportOuterTransfer"
	Vote_Name             = "Vote"
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
	native.Register(Vote_Name, Vote)
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
	log.Debugf("-===ImportExTransfer")
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, contract params deserialize error: %v", err)
	}
	log.Infof("SourceChainID:%v\n", params.SourceChainID)
	log.Infof("TargetChainID:%v\n", params.TargetChainID)
	log.Infof("Proof:%v\n", params.Proof)
	log.Infof("TxData:%v\n", params.TxData)
	log.Infof("Height:%v\n", params.Height)
	log.Infof("RelayerAddress:%v\n", params.RelayerAddress)
	log.Infof("value:%v\n", params.Value)

	chainid := params.SourceChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.Chainid != chainid {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side chain is not registered")
	}

	handler, err := GetChainHandler(chainid)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//1. verify tx
	err = handler.Verify(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	return utils.BYTE_TRUE, nil
}

func Vote(native *native.NativeService) ([]byte, error) {
	log.Debugf("-===Vote")
	params := new(inf.VoteParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Vote, contract params deserialize error: %v", err)
	}

	from := params.FromChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, from)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Vote, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.Chainid != from {
		return utils.BYTE_FALSE, fmt.Errorf("Vote, side chain is not registered")
	}

	handler, err := GetChainHandler(from)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//1. vote
	ok, txParam, err := handler.Vote(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if ok {
		//2. make target chain tx
		targetid := txParam.ToChainID

		//check if chainid exist
		sideChain, err = side_chain_manager.GetSideChain(native, targetid)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
		}
		if sideChain.Chainid != targetid {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, targetid chain is not registered")
		}

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
	return utils.BYTE_TRUE, nil
}