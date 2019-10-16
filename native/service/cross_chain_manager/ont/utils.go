/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ont

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/merkle"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/header_sync/ont"
	"github.com/ontio/multi-chain/native/service/utils"
)

func putDoneTx(native *native.NativeService, txHash common.Uint256, chainID uint64) error {
	contract := utils.CrossChainManagerContractAddress
	prefix := txHash.ToArray()
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, prefix), cstates.GenRawStorageItem(txHash.ToArray()))
	return nil
}

func checkDoneTx(native *native.NativeService, txHash common.Uint256, chainID uint64) error {
	contract := utils.CrossChainManagerContractAddress
	prefix := txHash.ToArray()
	chainIDBytes := utils.GetUint64Bytes(chainID)
	value, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, prefix))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.GetCacheDB().Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}

func putRequest(native *native.NativeService, txHash common.Uint256, chainID uint64, request []byte) error {
	contract := utils.CrossChainManagerContractAddress
	prefix := txHash.ToArray()
	chainIDBytes := utils.GetUint64Bytes(chainID)
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(REQUEST), chainIDBytes, prefix), request)
	return nil
}

func VerifyFromOntTx(native *native.NativeService, proof []byte, fromChainid uint64, height uint32) (*FromMerkleValue, error) {
	//get block header
	header, err := ont.GetHeaderByHeight(native, fromChainid, height)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, get header by height %d from chain %d error: %v",
			height, fromChainid, err)
	}

	v, err := merkle.MerkleProve(proof, header.CrossStatesRoot.ToArray())
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, merkle.MerkleProve verify merkle proof error")
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(FromMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	err = checkDoneTx(native, merkleValue.TxHash, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, checkDoneTx error:%s", err)
	}
	err = putDoneTx(native, merkleValue.TxHash, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, putDoneTx error:%s", err)
	}

	notifyVerifyFromOntProof(native, merkleValue.TxHash.ToHexString(), merkleValue.CreateCrossChainTxMerkle.ToChainID)
	return merkleValue, nil
}

func notifyVerifyFromOntProof(native *native.NativeService, txHash string, toChainID uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{VERIFY_FROM_ONT_PROOF, txHash, toChainID, native.GetHeight()},
		})
}
