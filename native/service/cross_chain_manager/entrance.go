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
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
)

const (
	IMPORT_OUTER_TRANSFER_NAME = "ImportOuterTransfer"
	MULTI_SIGN                 = "MultiSign"
)

func RegisterCrossChainManagerContract(native *native.NativeService) {
	native.Register(IMPORT_OUTER_TRANSFER_NAME, ImportExTransfer)
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
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side chain %d is not registered", chainID)
	}

	handler, err := GetChainHandler(sideChain.Router)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//1. verify tx
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
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side chain %d is not registered", targetid)
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

func MultiSign(native *native.NativeService) ([]byte, error) {
	handler := btc.NewBTCHandler()

	//1. multi sign
	err := handler.MultiSign(native)
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
	err := PutRequest(service, merkleValue.TxHash, params.ToChainID, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeTransaction, putRequest error:%s", err)
	}
	service.PutMerkleVal(sink.Bytes())
	chainIDBytes := utils.GetUint64Bytes(params.ToChainID)
	key := hex.EncodeToString(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(scom.REQUEST), chainIDBytes, merkleValue.TxHash))
	scom.NotifyMakeProof(service, fromChainID, params.ToChainID, hex.EncodeToString(params.TxHash), key)
	return nil
}

func PutRequest(native *native.NativeService, txHash []byte, chainID uint64, request []byte) error {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(scom.REQUEST), chainIDBytes, txHash), request)
	return nil
}
