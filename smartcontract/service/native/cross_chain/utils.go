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

package cross_chain

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func putSideChain(native *native.NativeService, sideChain *chain_manager.SideChain) error {
	contract := utils.ChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	sideChain.Serialize(sink)
	chainIDBytes, err := utils.GetUint64Bytes(sideChain.ChainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(chain_manager.SIDE_CHAIN), chainIDBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

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

func notifyCreateCrossChainTx(native *native.NativeService, chainID uint64, requestID uint64, height uint32, ongxFee uint64) {
	contract := utils.CrossChainContractAddress
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{CREATE_CROSS_CHAIN_TX, chainID, requestID, height, ongxFee},
		})
}

func notifyProcessCrossChainTx(native *native.NativeService, chainID uint64, requestID uint64, height uint32, ongFee uint64) {
	contract := utils.CrossChainContractAddress
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{PROCESS_CROSS_CHAIN_TX, chainID, requestID, height, ongFee},
		})
}
