package cross_chain_manager

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/neo"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"

	"github.com/ontio/multi-chain/native/service/cross_chain_manager/btc"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/eth"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/ont"
	"github.com/ontio/multi-chain/native/service/side_chain_manager"
)

const (
	IMPORT_OUTER_TRANSFER_NAME = "ImportOuterTransfer"
	VOTE_NAME                  = "Vote"
	MULTI_SIGN                 = "MultiSign"
	INIT_REDEEM_SCRIPT         = "initRedeemScript"
)

func RegisterCrossChainManagerContract(native *native.NativeService) {
	native.Register(INIT_REDEEM_SCRIPT, InitRedeemScript)

	native.Register(IMPORT_OUTER_TRANSFER_NAME, ImportExTransfer)
	native.Register(VOTE_NAME, Vote)
	native.Register(MULTI_SIGN, MultiSign)
}

func GetChainHandler(router uint64) (scom.ChainHandler, error) {
	switch router {
	case utils.BTC_ROUTER:
		return btc.NewBTCHandler(), nil
	case utils.ETH_ROUTER:
		return eth.NewETHHandler(), nil
	case utils.ONT_ROUTER:
		return ont.NewONTHandler(), nil
	case utils.NEO_ROUTER:
		return neo.NewNEOHandler(), nil
	default:
		return nil, fmt.Errorf("not a supported router:%d", router)
	}
}

func ImportExTransfer(native *native.NativeService) ([]byte, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, contract params deserialize error: %v", err)
	}

	chainID := params.SourceChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.ChainId != chainID {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side chain is not registered")
	}

	handler, err := GetChainHandler(sideChain.Router)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//1. verify tx
	if sideChain.Router == utils.BTC_ROUTER {
		_, err = handler.MakeDepositProposal(native)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		return utils.BYTE_TRUE, nil
	}

	txParam, err := handler.MakeDepositProposal(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//2. make target chain tx
	targetid := txParam.ToChainID

	//check if chainid exist
	sideChain, err = side_chain_manager.GetSideChain(native, targetid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.ChainId != targetid {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, targetid chain is not registered")
	}

	if sideChain.Router == utils.BTC_ROUTER {
		err := btc.NewBTCHandler().MakeTransaction(native, txParam, chainID)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		return utils.BYTE_TRUE, nil
	}

	//NOTE, you need to store the tx in this
	err = MakeTransaction(native, txParam, chainID)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func Vote(native *native.NativeService) ([]byte, error) {
	//1. vote
	ok, txParam, fromChainID, err := btc.NewBTCHandler().Vote(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if ok {
		//2. make target chain tx
		targetid := txParam.ToChainID

		//check if chainid exist
		sideChain, err := side_chain_manager.GetSideChain(native, targetid)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
		}
		if sideChain.ChainId != targetid {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, targetid chain is not registered")
		}
		//NOTE, you need to store the tx in this
		err = MakeTransaction(native, txParam, fromChainID)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		return utils.BYTE_TRUE, nil
	}
	return utils.BYTE_TRUE, nil
}

func MultiSign(native *native.NativeService) ([]byte, error) {
	handler := btc.NewBTCHandler()

	//1. multi sign
	err := handler.MultiSign(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func InitRedeemScript(native *native.NativeService) ([]byte, error) {
	handler := btc.NewBTCHandler()

	//1. multi sign
	err := handler.InitRedeemScript(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func MakeTransaction(service *native.NativeService, params *scom.MakeTxParam, fromChainID uint64) error {
	txHash := service.GetTx().Hash()
	merkleValue := &scom.ToMerkleValue{
		TxHash:      txHash.ToArray(),
		FromChainID: fromChainID,
		MakeTxParam: params,
	}

	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	err := putRequest(service, merkleValue.TxHash, params.ToChainID, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeTransaction, putRequest error:%s", err)
	}
	service.PutMerkleVal(sink.Bytes())
	chainIDBytes := utils.GetUint64Bytes(params.ToChainID)
	key := hex.EncodeToString(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(scom.REQUEST), chainIDBytes, merkleValue.TxHash))
	scom.NotifyMakeProof(service, hex.EncodeToString(params.TxHash), params.ToChainID, key)
	return nil
}

func putRequest(native *native.NativeService, txHash []byte, chainID uint64, request []byte) error {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(scom.REQUEST), chainIDBytes, txHash), request)
	return nil
}
