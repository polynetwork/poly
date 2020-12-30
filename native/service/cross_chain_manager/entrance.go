/*
 * Copyright (C) 2020 The poly network Authors
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
package cross_chain_manager

import (
	"encoding/hex"
	"fmt"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/quorum"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/heco"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/bsc"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/btc"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/cosmos"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/eth"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/neo"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/ont"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	IMPORT_OUTER_TRANSFER_NAME = "ImportOuterTransfer"
	MULTI_SIGN                 = "MultiSign"
	BLACK_CHAIN                = "BlackChain"
	WHITE_CHAIN                = "WhiteChain"

	BLACKED_CHAIN = "BlackedChain"
)

func RegisterCrossChainManagerContract(native *native.NativeService) {
	native.Register(IMPORT_OUTER_TRANSFER_NAME, ImportExTransfer)
	native.Register(MULTI_SIGN, MultiSign)

	native.Register(BLACK_CHAIN, BlackChain)
	native.Register(WHITE_CHAIN, WhiteChain)
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
	case utils.COSMOS_ROUTER:
		return cosmos.NewCosmosHandler(), nil
	case utils.QUORUM_ROUTER:
		return quorum.NewQuorumHandler(), nil
	case utils.BSC_ROUTER:
		return bsc.NewHandler(), nil
	case utils.HECO_ROUTER:
		return heco.NewHecoHandler(), nil
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
	blacked, err := CheckIfChainBlacked(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, CheckIfChainBlacked error: %v", err)
	}
	if blacked {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, source chain is blacked")
	}

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
	blacked, err = CheckIfChainBlacked(native, targetid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, CheckIfChainBlacked error: %v", err)
	}
	if blacked {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, target chain is blacked")
	}

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

func BlackChain(native *native.NativeService) ([]byte, error) {
	params := new(BlackChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackChain, contract params deserialize error: %v", err)
	}

	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackChain, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackChain, checkWitness error: %v", err)
	}

	PutBlackChain(native, params.ChainID)
	return utils.BYTE_TRUE, nil
}

func WhiteChain(native *native.NativeService) ([]byte, error) {
	params := new(BlackChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WhiteChain, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WhiteChain, get current consensus operator address error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackChain, checkWitness error: %v", err)
	}

	RemoveBlackChain(native, params.ChainID)
	return utils.BYTE_TRUE, nil
}
