/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package rest

import (
	"encoding/hex"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/log"
	scom "github.com/polynetwork/poly/core/store/common"
	"github.com/polynetwork/poly/core/types"
	ontErrors "github.com/polynetwork/poly/errors"
	bactor "github.com/polynetwork/poly/http/base/actor"
	bcomn "github.com/polynetwork/poly/http/base/common"
	berr "github.com/polynetwork/poly/http/base/error"
	"strconv"
)

const TLS_PORT int = 443

type ApiServer interface {
	Start() error
	Stop()
}

// get node verison
func GetNodeVersion(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = config.Version
	return resp
}

// get networkid
func GetNetworkId(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = config.DefConfig.P2PNode.NetworkId
	return resp
}

//get connection node count
func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := bactor.GetConnectionCnt()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//get block height
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	height := bactor.GetCurrentBlockHeight()
	resp["Result"] = height
	return resp
}

//get block hash by height
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash := bactor.GetBlockHashFromStore(uint32(height))
	if hash == common.UINT256_EMPTY {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	resp["Result"] = hash.ToHexString()
	return resp
}

func getBlock(hash common.Uint256, getTxBytes bool) (interface{}, int64) {
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if block == nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if block.Header == nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if getTxBytes {
		return common.ToHexString(block.ToArray()), berr.SUCCESS
	}
	return bcomn.GetBlockInfo(block), berr.SUCCESS
}

//get block by hash
func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	if len(str) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash common.Uint256
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}

//get block height by transaction hash
func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok || len(str) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if tx == nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = height
	return resp
}

//get block transaction hashes by height
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	hash := bactor.GetBlockHashFromStore(index)
	if hash == common.UINT256_EMPTY {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	resp["Result"] = bcomn.GetBlockTransactions(block)
	return resp
}

//get block by height
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	block, err := bactor.GetBlockByHeight(index)
	if err != nil || block == nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if getTxBytes {
		resp["Result"] = common.ToHexString(block.ToArray())
	} else {
		resp["Result"] = bcomn.GetBlockInfo(block)
	}
	return resp
}

//get transaction by hash
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, tx, err := bactor.GetTxnWithHeightByTxHash(hash)
	if tx == nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if err != nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if tx == nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		resp["Result"] = common.ToHexString(tx.Raw)
		return resp
	}
	tran := bcomn.TransArryByteToHexString(tx)
	tran.Height = height
	resp["Result"] = tran
	return resp
}

//send raw transaction
func SendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Data"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	txn, err := types.TransactionFromRawBytes(bys)
	if err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	var hash common.Uint256
	hash = txn.Hash()
	log.Debugf("SendRawTransaction recv %s", hash.ToHexString())
	if txn.TxType == types.Invoke || txn.TxType == types.Deploy {
		if preExec, ok := cmd["PreExec"].(string); ok && preExec == "1" {
			rst, err := bactor.PreExecuteContract(txn)
			if err != nil {
				log.Infof("PreExec: ", err)
				resp = ResponsePack(berr.SMARTCODE_ERROR)
				resp["Result"] = err.Error()
				return resp
			}
			resp["Result"] = bcomn.ConvertPreExecuteResult(rst)
			return resp
		}
	}
	log.Debugf("SendRawTransaction send to txpool %s", hash.ToHexString())
	if errCode, desc := bcomn.SendTxToPool(txn); errCode != ontErrors.ErrNoError {
		resp["Error"] = int64(errCode)
		resp["Result"] = desc
		log.Warnf("SendRawTransaction verified %s error: %s", hash.ToHexString(), desc)
		return resp
	}
	log.Debugf("SendRawTransaction verified %s", hash.ToHexString())
	resp["Result"] = hash.ToHexString()
	return resp
}

//get smartcontract event by height
func GetSmartCodeEventTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	eventInfos, err := bactor.GetEventNotifyByHeight(index)
	if err != nil {
		if scom.ErrNotFound == err {
			return ResponsePack(berr.SUCCESS)
		}
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	eInfos := make([]*bcomn.ExecuteNotify, 0, len(eventInfos))
	for _, eventInfo := range eventInfos {
		_, notify := bcomn.GetExecuteNotify(eventInfo)
		eInfos = append(eInfos, &notify)
	}
	resp["Result"] = eInfos
	return resp
}

//get smartcontract event by transaction hash
func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {
	if !config.DefConfig.Common.EnableEventLog {
		return ResponsePack(berr.INVALID_METHOD)
	}

	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	eventInfo, err := bactor.GetEventNotifyByTxHash(hash)
	if err != nil {
		if scom.ErrNotFound == err {
			return ResponsePack(berr.SUCCESS)
		}
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if eventInfo == nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	_, notify := bcomn.GetExecuteNotify(eventInfo)
	resp["Result"] = notify
	return resp
}

//get storage from contract
func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	address, err := bcomn.GetAddress(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	key := cmd["Key"].(string)
	item, err := common.HexToBytes(key)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	value, err := bactor.GetStorageItem(address, item)
	if err != nil {
		if err == scom.ErrNotFound {
			return ResponsePack(berr.SUCCESS)
		}
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = common.ToHexString(value)
	return resp
}

//get merkle proof by transaction hash
func GetMerkleProof(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	bhStr, ok := cmd["BlockHeight"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	rhStr, ok := cmd["RootHeight"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(bhStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	rootHeight, err := strconv.ParseInt(rhStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if height >= rootHeight || height == 0 {
		ResponsePack(berr.INVALID_PARAMS)
	}
	proof, err := bactor.GetMerkleProof(uint32(height), uint32(rootHeight))
	if err != nil {
		ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = bcomn.MerkleProof{"MerkleProof", hex.EncodeToString(proof)}
	return resp
}

//get memory pool transaction count
func GetMemPoolTxCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := bactor.GetTxnCount()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//get memory poll transaction state
func GetMemPoolTxState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	txEntry, err := bactor.GetTxFromPool(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	attrs := []bcomn.TXNAttrInfo{}
	for _, t := range txEntry.Attrs {
		attrs = append(attrs, bcomn.TXNAttrInfo{t.Height, int(t.Type), int(t.ErrCode)})
	}
	resp["Result"] = bcomn.TXNEntryInfo{attrs}
	return resp
}
