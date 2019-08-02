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
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/merkle"
	"github.com/ontio/multi-chain/smartcontract/event"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/header_sync"
	"github.com/ontio/multi-chain/smartcontract/service/native/ont"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
	"github.com/ontio/ontology/common/config"
)

func appCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := appCallTransfer(native, utils.OngContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOng, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []ont.State
	sts = append(sts, ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := ont.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return fmt.Errorf("appCallTransfer, appCall error: %v", err)
	}
	return nil
}

func putRequestID(native *native.NativeService, requestID uint64, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	requestIDBytes, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return fmt.Errorf("putRequestID, get requestIDBytes error: %v", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequestID, get chainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(REQUEST_ID), chainIDBytes), cstates.GenRawStorageItem(requestIDBytes))
	return nil
}

func getRequestID(native *native.NativeService, chainID uint64) (uint64, error) {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return 0, fmt.Errorf("getRequestID, get chainIDBytes error: %v", err)
	}
	var requestID uint64 = 0
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(REQUEST_ID), chainIDBytes))
	if err != nil {
		return 0, fmt.Errorf("getRequestID, get requestID value error: %v", err)
	}
	if value != nil {
		requestIDBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return 0, fmt.Errorf("getRequestID, deserialize from raw storage item err:%v", err)
		}
		requestID, err = utils.GetBytesUint64(requestIDBytes)
		if err != nil {
			return 0, fmt.Errorf("getRequestID, get requestID error: %v", err)
		}
	}
	return requestID, nil
}

func putRequest(native *native.NativeService, requestID uint64, chainID uint64, request []byte) error {
	contract := utils.CrossChainContractAddress
	prefix, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return fmt.Errorf("putRequest, GetUint64Bytes error:%s", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequest, get chainIDBytes error: %v", err)
	}
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(REQUEST), chainIDBytes, prefix), request)
	return nil
}

//must be called before putCurrentID
func putRemainedIDs(native *native.NativeService, requestID, currentID uint64, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	for i := currentID + 1; i < requestID; i++ {
		requestIDBytes, err := utils.GetUint64Bytes(i)
		if err != nil {
			return fmt.Errorf("putRemainedID, get requestIDBytes error: %v", err)
		}
		chainIDBytes, err := utils.GetUint64Bytes(chainID)
		if err != nil {
			return fmt.Errorf("putRemainedID, get chainIDBytes error: %v", err)
		}
		native.CacheDB.Put(utils.ConcatKey(contract, []byte(REMAINED_ID), chainIDBytes, requestIDBytes), cstates.GenRawStorageItem(requestIDBytes))
	}
	return nil
}

func checkIfRemained(native *native.NativeService, requestID uint64, chainID uint64) (bool, error) {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return false, fmt.Errorf("checkIfRemained, get chainIDBytes error: %v", err)
	}
	requestIDBytes, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return false, fmt.Errorf("checkIfRemained, get requestIDBytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(REMAINED_ID), chainIDBytes, requestIDBytes))
	if err != nil {
		return false, fmt.Errorf("checkIfRemained, get value error: %v", err)
	}
	if value == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func removeRemained(native *native.NativeService, requestID uint64, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("removeRemained, get chainIDBytes error: %v", err)
	}
	requestIDBytes, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return fmt.Errorf("removeRemained, get requestIDBytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(REMAINED_ID), chainIDBytes, requestIDBytes))
	return nil
}

func putCurrentID(native *native.NativeService, currentID uint64, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	currentIDBytes, err := utils.GetUint64Bytes(currentID)
	if err != nil {
		return fmt.Errorf("putCurrentID, get currentIDBytes error: %v", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequestID, get chainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CURRENT_ID), chainIDBytes), cstates.GenRawStorageItem(currentIDBytes))
	return nil
}

func getCurrentID(native *native.NativeService, chainID uint64) (uint64, error) {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return 0, fmt.Errorf("getCurrentID, get chainIDBytes error: %v", err)
	}
	var currentID uint64 = 0
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CURRENT_ID), chainIDBytes))
	if err != nil {
		return 0, fmt.Errorf("getCurrentID, get currentID value error: %v", err)
	}
	if value != nil {
		currentIDBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return 0, fmt.Errorf("getCurrentID, deserialize from raw storage item err:%v", err)
		}
		currentID, err = utils.GetBytesUint64(currentIDBytes)
		if err != nil {
			return 0, fmt.Errorf("getCurrentID, get currentID error: %v", err)
		}
	}
	return currentID, nil
}

func VerifyFromOntTx(native *native.NativeService, proof []byte, fromChainid uint64, height uint32) (*FromMerkleValue, error) {
	//get block header
	header, err := header_sync.GetHeaderByHeight(native, fromChainid, height)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, get header by height error: %v", err)
	}

	v := merkle.MerkleProve(proof, header.CrossStatesRoot)
	if v == nil {
		return nil, fmt.Errorf("VerifyFromOntTx, merkle.MerkleProve verify merkle proof error")
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(FromMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	oldCurrentID, err := getCurrentID(native, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, getCurrentID error: %v", err)
	}
	if merkleValue.RequestID > oldCurrentID {
		err = putRemainedIDs(native, merkleValue.RequestID, oldCurrentID, fromChainid)
		if err != nil {
			return nil, fmt.Errorf("VerifyFromOntTx, putRemainedIDs error: %v", err)
		}
		err = putCurrentID(native, merkleValue.RequestID, fromChainid)
		if err != nil {
			return nil, fmt.Errorf("VerifyFromOntTx, putCurrentID error: %v", err)
		}
	} else {
		ok, err := checkIfRemained(native, merkleValue.RequestID, fromChainid)
		if err != nil {
			return nil, fmt.Errorf("VerifyFromOntTx, checkIfRemained error: %v", err)
		}
		if !ok {
			return nil, fmt.Errorf("VerifyFromOntTx, tx already done")
		} else {
			err = removeRemained(native, merkleValue.RequestID, fromChainid)
			if err != nil {
				return nil, fmt.Errorf("VerifyFromOntTx, removeRemained error: %v", err)
			}
		}
	}
	return merkleValue, nil
}

func MakeFromOntProof(native *native.NativeService, params *CreateCrossChainTxParam) error {
	//record cross chain tx
	requestID, err := getRequestID(native, params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, getRequestID error:%s", err)
	}
	newID := requestID + 1
	merkleValue := &FromMerkleValue{
		RequestID: newID,
		CreateCrossChainTxMerkle: &CreateCrossChainTxMerkle{
			FromChainID:         native.ShardID.ToUint64(),
			FromContractAddress: native.ContextRef.CallingContext().ContractAddress.ToHexString(),
			ToChainID:           params.ToChainID,
			Fee:                 params.Fee,
			Address:             params.Address,
			Amount:              params.Amount,
		},
	}
	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	err = putRequest(native, newID, params.ToChainID, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, putRequest error:%s", err)
	}
	native.ContextRef.PutMerkleVal(sink.Bytes())
	err = putRequestID(native, newID, params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, putRequestID error:%s", err)
	}
	prefix, err := utils.GetUint64Bytes(newID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, GetUint64Bytes error:%s", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, get chainIDBytes error: %v", err)
	}
	key := hex.EncodeToString(utils.ConcatKey(utils.CrossChainContractAddress, []byte(REQUEST), chainIDBytes, prefix))
	notifyMakeFromOntProof(native, params.ToChainID, key)
	return nil
}

func VerifyToOntTx(native *native.NativeService, proof []byte, fromChainid uint64, height uint32) (*ToMerkleValue, error) {
	//get block header
	header, err := header_sync.GetHeaderByHeight(native, fromChainid, height)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, get header by height error: %v", err)
	}

	v := merkle.MerkleProve(proof, header.CrossStatesRoot)
	if v == nil {
		return nil, fmt.Errorf("VerifyFromOntTx, merkle.MerkleProve verify merkle proof error")
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(ToMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	oldCurrentID, err := getCurrentID(native, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, getCurrentID error: %v", err)
	}
	if merkleValue.RequestID > oldCurrentID {
		err = putRemainedIDs(native, merkleValue.RequestID, oldCurrentID, fromChainid)
		if err != nil {
			return nil, fmt.Errorf("VerifyFromOntTx, putRemainedIDs error: %v", err)
		}
		err = putCurrentID(native, merkleValue.RequestID, fromChainid)
		if err != nil {
			return nil, fmt.Errorf("VerifyFromOntTx, putCurrentID error: %v", err)
		}
	} else {
		ok, err := checkIfRemained(native, merkleValue.RequestID, fromChainid)
		if err != nil {
			return nil, fmt.Errorf("VerifyFromOntTx, checkIfRemained error: %v", err)
		}
		if !ok {
			return nil, fmt.Errorf("VerifyFromOntTx, tx already done")
		} else {
			err = removeRemained(native, merkleValue.RequestID, fromChainid)
			if err != nil {
				return nil, fmt.Errorf("VerifyFromOntTx, removeRemained error: %v", err)
			}
		}
	}
	return merkleValue, nil
}

func MakeToOntProof(native *native.NativeService, params *inf.MakeTxParam) error {
	//record cross chain tx
	requestID, err := getRequestID(native, native.ShardID.ToUint64())
	if err != nil {
		return fmt.Errorf("MakeToOntProof, getRequestID error:%s", err)
	}
	newID := requestID + 1
	merkleValue := &ToMerkleValue{
		RequestID:   newID,
		MakeTxParam: params,
	}
	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	err = putRequest(native, newID, native.ShardID.ToUint64(), sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeToOntProof, putRequest error:%s", err)
	}
	native.ContextRef.PutMerkleVal(sink.Bytes())
	err = putRequestID(native, newID, native.ShardID.ToUint64())
	if err != nil {
		return fmt.Errorf("MakeToOntProof, putRequestID error:%s", err)
	}
	prefix, err := utils.GetUint64Bytes(newID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, GetUint64Bytes error:%s", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, get chainIDBytes error: %v", err)
	}
	key := hex.EncodeToString(utils.ConcatKey(utils.CrossChainContractAddress, []byte(REQUEST), chainIDBytes, prefix))
	notifyMakeToOntProof(native, params.ToChainID, key)
	return nil
}

func notifyMakeFromOntProof(native *native.NativeService, toChainID uint64, key string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.OngContractAddress,
			States:          []interface{}{MAKE_FROM_ONT_PROOF, toChainID, native.Height, key},
		})
}

func notifyMakeToOntProof(native *native.NativeService, toChainID uint64, key string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{MAKE_TO_ONT_PROOF, toChainID, native.Height, key},
		})
}
