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

package neo

import (
	"encoding/hex"
	"fmt"
	"github.com/joeqian10/neo-gogogo/tx"
	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	hscommon "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/header_sync/ont"
	"github.com/ontio/multi-chain/native/service/utils"
)

//verify header of any height
//find key height and get neoconsensus first, then check the witness
func verifyHeader(native *native.NativeService, chainID uint64, header *NeoBlockHeader) error {
	height := header.Index
	//search consensus peer
	keyHeight, err := ont.FindKeyHeight(native, height, chainID)
	if err != nil {
		return fmt.Errorf("verifyHeader, findKeyHeight error:%s", err)
	}

	neoConsensus, err := getNextConsensusByHeight(native, chainID, keyHeight)
	if err != nil {
		return fmt.Errorf("verifyHeader, get Consensus error:%s", err)
	}
	if neoConsensus.NextConsensus != header.Witness.GetScriptHash() {
		return fmt.Errorf("verifyHeader, invalid script hash in header error, expected:%s, got:%s", neoConsensus.NextConsensus.String(), header.Witness.GetScriptHash().String())
	}

	msg, err := header.GetMessage()
	if err != nil {
		return fmt.Errorf("verifyHeader, unable to get hash data of header")
	}
	if verified := tx.VerifyMultiSignatureWitness(msg, header.Witness); !verified {
		return fmt.Errorf("verifyHeader, VerifyMultiSignatureWitness error:%s, height:%d", err, header.Index)
	}
	return nil
}

func VerifyCrossChainMsg(native *native.NativeService, chainID uint64, crossChainMsg *NeoCrossChainMsg) error {
	height := crossChainMsg.Index
	//search consensus peer
	keyHeight, err := ont.FindKeyHeight(native, height, chainID)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, findKeyHeight error:%v", err)
	}

	neoConsensus, err := getNextConsensusByHeight(native, chainID, keyHeight)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, get ConsensusPeer error:%v", err)
	}
	crossChainMsgConsensus, err := crossChainMsg.GetScriptHash()
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, getScripthash error:%v", err)
	}
	if neoConsensus.NextConsensus != crossChainMsgConsensus {
		return fmt.Errorf("verifyCrossChainMsg, invalid script hash in NeoCrossChainMsg error, expected:%s, got:%s", neoConsensus.NextConsensus.String(), crossChainMsgConsensus.String())
	}
	msg, err := crossChainMsg.GetMessage()
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, unable to get unsigned message of neo crossChainMsg")
	}
	invScript, _ := hex.DecodeString(crossChainMsg.Witness.InvocationScript)
	verScript, _ := hex.DecodeString(crossChainMsg.Witness.VerificationScript)
	witness := &tx.Witness{
		InvocationScript:   invScript,
		VerificationScript: verScript,
	}
	if verified := tx.VerifyMultiSignatureWitness(msg, witness); !verified {
		return fmt.Errorf("verifyCrossChainMsg, VerifyMultiSignatureWitness error:%s, height:%d", "verified failed", crossChainMsg.Index)
	}
	return nil
}

func PutBlockHeader(native *native.NativeService, chainID uint64, blockHeader *NeoBlockHeader) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	if err := blockHeader.Serialization(sink); err != nil {
		return fmt.Errorf("PubBlockHeader, NeoBlockHeaderToBytes error:%s", err)
	}
	height := blockHeader.Index
	heightBytes := utils.GetUint32Bytes(height)
	chainIDBytes := utils.GetUint64Bytes(chainID)

	blockHash := blockHeader.Hash()

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.BLOCK_HEADER), chainIDBytes, blockHash.Bytes()),
		cstates.GenRawStorageItem(sink.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.HEADER_INDEX), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(blockHash.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CURRENT_HEADER_HEIGHT), chainIDBytes),
		cstates.GenRawStorageItem(heightBytes))
	hscommon.NotifyPutHeader(native, chainID, height, blockHash.String())
	return nil
}

func GetHeaderByHeight(native *native.NativeService, chainID uint64, height uint32) (*NeoBlockHeader, error) {
	contract := utils.HeaderSyncContractAddress
	heightBytes := utils.GetUint32Bytes(height)
	chainIDBytes := utils.GetUint64Bytes(chainID)

	blockHashStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.HEADER_INDEX),
		chainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if blockHashStore == nil {
		return nil, fmt.Errorf("GetHeaderByHeight, can not find any index records")
	}
	blockHashBytes, err := cstates.GetValueFromRawStorageItem(blockHashStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize blockHashBytes from raw storage item err:%v", err)
	}
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.BLOCK_HEADER),
		chainIDBytes, blockHashBytes))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get headerStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHeight, can not find any header records")
	}
	headerBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	header := new(NeoBlockHeader)
	if err := header.Deserialization(common.NewZeroCopySource(headerBytes)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize bytes to RpcBlockHeader error:%v", err)
	}
	return header, nil
}

func GetHeaderByHash(native *native.NativeService, chainID uint64, hash common.Uint256) (*NeoBlockHeader, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.BLOCK_HEADER), chainIDBytes, hash.ToArray()))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get headerStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHash, can not find any records")
	}
	headerBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize from raw storage item err:%v", err)
	}
	header := new(NeoBlockHeader)
	if err := header.Deserialization(common.NewZeroCopySource(headerBytes)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize bytes to RpcBlockHeader error:%v", err)
	}
	return header, nil
}

func PutCrossChainMsg(native *native.NativeService, chainID uint64, crossChainMsg *NeoCrossChainMsg) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	crossChainMsg.Serialization(sink)
	chainIDBytes := utils.GetUint64Bytes(chainID)
	heightBytes := utils.GetUint32Bytes(crossChainMsg.Index)

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CROSS_CHAIN_MSG), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CURRENT_MSG_HEIGHT), chainIDBytes),
		cstates.GenRawStorageItem(heightBytes))
	hscommon.NotifyPutCrossChainMsg(native, chainID, crossChainMsg.Index)
	return nil
}

func GetCrossChainMsg(native *native.NativeService, chainID uint64, height uint32) (*NeoCrossChainMsg, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	heightBytes := utils.GetUint32Bytes(height)

	crossChainMsg := new(NeoCrossChainMsg)
	crossChainMsgStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.CROSS_CHAIN_MSG),
		chainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("GetCrossChainMsg, get headerStore error: %v", err)
	}
	if crossChainMsgStore == nil {
		return nil, fmt.Errorf("GetCrossChainMsg, can not find any header records")
	}
	crossChainMsgBytes, err := cstates.GetValueFromRawStorageItem(crossChainMsgStore)
	if err != nil {
		return nil, fmt.Errorf("GetCrossChainMsg, deserialize headerBytes from raw storage item err:%v", err)
	}
	if err := crossChainMsg.Deserialization(common.NewZeroCopySource(crossChainMsgBytes)); err != nil {
		return nil, fmt.Errorf("GetCrossChainMsg, deserialize header error: %v", err)
	}
	return crossChainMsg, nil
}

func UpdateConsensusPeer(native *native.NativeService, chainID uint64, header *NeoBlockHeader) error {
	//search consensus peer
	keyHeight, err := ont.FindKeyHeight(native, header.Index, chainID)
	if err != nil {
		return fmt.Errorf("UpdateConsensusPeer, findKeyHeight error:%s", err)
	}
	prevNeoConsensus, err := getNextConsensusByHeight(native, chainID, keyHeight)
	if err != nil {
		return fmt.Errorf("getNextConsensusByHeight error:%s", err)
	}
	if prevNeoConsensus.NextConsensus != header.NextConsensus {
		neoConsensus := &NeoConsensus{
			ChainID:       chainID,
			Height:        header.Index,
			NextConsensus: header.NextConsensus,
		}

		err := putNextConsensusByHeight(native, neoConsensus)
		if err != nil {
			return fmt.Errorf("updateConsensusPeer, putNextConsensusByHeight eerror: %s", err)
		}
	}
	return nil
}

func getNextConsensusByHeight(native *native.NativeService, chainID uint64, height uint32) (*NeoConsensus, error) {
	contract := utils.HeaderSyncContractAddress
	heightBytes := utils.GetUint32Bytes(height)
	chainIDBytes := utils.GetUint64Bytes(chainID)
	neoConsensusStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER), chainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("getNextConsensusByHeight, get nextConsensusStore error: %v", err)
	}
	if neoConsensusStore == nil {
		return nil, fmt.Errorf("getNextConsensusByHeight, can not find any record")
	}
	neoConsensusBytes, err := cstates.GetValueFromRawStorageItem(neoConsensusStore)
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize from raw storage item err:%v", err)
	}
	neoConsensus := new(NeoConsensus)
	if err := neoConsensus.Deserialization(common.NewZeroCopySource(neoConsensusBytes)); err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize consensusPeer error: %v", err)
	}
	return neoConsensus, nil
}

func putNextConsensusByHeight(native *native.NativeService, neoConsensus *NeoConsensus) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	neoConsensus.Serialization(sink)
	chainIDBytes := utils.GetUint64Bytes(neoConsensus.ChainID)
	heightBytes := utils.GetUint32Bytes(neoConsensus.Height)
	blockHeightBytes := utils.GetUint32Bytes(native.GetHeight())

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER), chainIDBytes, heightBytes), cstates.GenRawStorageItem(sink.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER_BLOCK_HEIGHT), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(blockHeightBytes))

	//update key heights
	keyHeights, err := ont.GetKeyHeights(native, neoConsensus.ChainID)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, GetKeyHeights error: %v", err)
	}
	keyHeights.HeightList = append(keyHeights.HeightList, neoConsensus.Height)
	err = ont.PutKeyHeights(native, neoConsensus.ChainID, keyHeights)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, putKeyHeights error: %v", err)
	}
	return nil
}
