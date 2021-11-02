/*
 * Copyright (C) 2021 The poly network Authors
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

package ont

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"

	ocommon "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	otypes "github.com/ontio/ontology/core/types"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/consensus/vbft/config"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutCrossChainMsg(native *native.NativeService, chainID uint64, crossChainMsg *otypes.CrossChainMsg) error {
	contract := utils.HeaderSyncContractAddress
	sink := ocommon.NewZeroCopySink(nil)
	crossChainMsg.Serialization(sink)
	chainIDBytes := utils.GetUint64Bytes(chainID)
	heightBytes := utils.GetUint32Bytes(crossChainMsg.Height)

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CROSS_CHAIN_MSG), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CURRENT_MSG_HEIGHT), chainIDBytes),
		cstates.GenRawStorageItem(heightBytes))
	hscommon.NotifyPutCrossChainMsg(native, chainID, crossChainMsg.Height)
	return nil
}

func GetCrossChainMsg(native *native.NativeService, chainID uint64, height uint32) (*otypes.CrossChainMsg, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	heightBytes := utils.GetUint32Bytes(height)

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
	crossChainMsg := new(otypes.CrossChainMsg)
	if err := crossChainMsg.Deserialization(ocommon.NewZeroCopySource(crossChainMsgBytes)); err != nil {
		return nil, fmt.Errorf("GetCrossChainMsg, deserialize header error: %v", err)
	}
	return crossChainMsg, nil
}

func PutBlockHeader(native *native.NativeService, chainID uint64, blockHeader *otypes.Header) error {
	contract := utils.HeaderSyncContractAddress
	sink := ocommon.NewZeroCopySink(nil)
	blockHeader.Serialization(sink)
	chainIDBytes := utils.GetUint64Bytes(chainID)
	heightBytes := utils.GetUint32Bytes(blockHeader.Height)

	blockHash := blockHeader.Hash()
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.BLOCK_HEADER), chainIDBytes, blockHash.ToArray()),
		cstates.GenRawStorageItem(sink.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.HEADER_INDEX), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(blockHash.ToArray()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CURRENT_HEADER_HEIGHT), chainIDBytes),
		cstates.GenRawStorageItem(heightBytes))
	hscommon.NotifyPutHeader(native, chainID, uint64(blockHeader.Height), blockHash.ToHexString())
	return nil
}

func GetHeaderByHeight(native *native.NativeService, chainID uint64, height uint32) (*otypes.Header, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	heightBytes := utils.GetUint32Bytes(height)

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
	header := new(otypes.Header)
	if err := header.Deserialization(ocommon.NewZeroCopySource(headerBytes)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize header error: %v", err)
	}
	return header, nil
}

func GetHeaderByHash(native *native.NativeService, chainID uint64, hash common.Uint256) (*otypes.Header, error) {
	contract := utils.HeaderSyncContractAddress

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.BLOCK_HEADER),
		utils.GetUint64Bytes(chainID), hash.ToArray()))
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
	header := new(otypes.Header)
	if err := header.Deserialization(ocommon.NewZeroCopySource(headerBytes)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return header, nil
}

func VerifyCrossChainMsg(native *native.NativeService, chainID uint64, crossChainMsg *otypes.CrossChainMsg,
	bookkeepers []keypair.PublicKey) error {
	height := crossChainMsg.Height
	//search consensus peer
	keyHeight, err := FindKeyHeight(native, height, chainID)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, findKeyHeight error:%v", err)
	}

	consensusPeer, err := getConsensusPeersByHeight(native, chainID, keyHeight)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, get ConsensusPeer error:%v", err)
	}
	if len(bookkeepers)*3 < len(consensusPeer.PeerMap) {
		return fmt.Errorf("verifyCrossChainMsg, header Bookkeepers num %d must more than 2/3 consensus node num %d",
			len(bookkeepers), len(consensusPeer.PeerMap))
	}
	for _, bookkeeper := range bookkeepers {
		pubkey := vconfig.PubkeyID(bookkeeper)
		_, present := consensusPeer.PeerMap[pubkey]
		if !present {
			return fmt.Errorf("verifyCrossChainMsg, invalid pubkey error:%v", pubkey)
		}
	}
	hash := crossChainMsg.Hash()
	err = signature.VerifyMultiSignature(hash[:], bookkeepers, len(bookkeepers),
		crossChainMsg.SigData)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, VerifyMultiSignature error:%s, heigh:%d", err,
			crossChainMsg.Height)
	}
	return nil
}

//verify header of any height
//find key height and get consensus peer first, then check the sign
func verifyHeader(native *native.NativeService, chainID uint64, header *otypes.Header) error {
	height := header.Height
	//search consensus peer
	keyHeight, err := FindKeyHeight(native, height, chainID)
	if err != nil {
		return fmt.Errorf("verifyHeader, findKeyHeight error:%v", err)
	}

	consensusPeer, err := getConsensusPeersByHeight(native, chainID, keyHeight)
	if err != nil {
		return fmt.Errorf("verifyHeader, get ConsensusPeer error:%v", err)
	}
	if len(header.Bookkeepers)*3 < len(consensusPeer.PeerMap) {
		return fmt.Errorf("verifyHeader, header Bookkeepers num %d must more than 2/3 consensus node num %d", len(header.Bookkeepers), len(consensusPeer.PeerMap))
	}
	for _, bookkeeper := range header.Bookkeepers {
		pubkey := vconfig.PubkeyID(bookkeeper)
		_, present := consensusPeer.PeerMap[pubkey]
		if !present {
			return fmt.Errorf("verifyHeader, invalid pubkey error:%v", pubkey)
		}
	}
	hash := header.Hash()
	err = signature.VerifyMultiSignature(hash[:], header.Bookkeepers, len(header.Bookkeepers), header.SigData)
	if err != nil {
		return fmt.Errorf("verifyHeader, VerifyMultiSignature error:%s, heigh:%d", err, header.Height)
	}
	return nil
}

func GetKeyHeights(native *native.NativeService, chainID uint64) (*KeyHeights, error) {
	contract := utils.HeaderSyncContractAddress
	value, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.KEY_HEIGHTS), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return nil, fmt.Errorf("GetKeyHeights, get keyHeights value error: %v", err)
	}
	keyHeights := new(KeyHeights)
	if value != nil {
		keyHeightsBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return nil, fmt.Errorf("GetKeyHeights, deserialize from raw storage item err:%v", err)
		}
		err = keyHeights.Deserialization(common.NewZeroCopySource(keyHeightsBytes))
		if err != nil {
			return nil, fmt.Errorf("GetKeyHeights, deserialize keyHeights err:%v", err)
		}
	}
	return keyHeights, nil
}

func PutKeyHeights(native *native.NativeService, chainID uint64, keyHeights *KeyHeights) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	keyHeights.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.KEY_HEIGHTS), utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getConsensusPeersByHeight(native *native.NativeService, chainID uint64, height uint32) (*ConsensusPeers, error) {
	contract := utils.HeaderSyncContractAddress
	heightBytes := utils.GetUint32Bytes(height)
	chainIDBytes := utils.GetUint64Bytes(chainID)
	consensusPeerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER), chainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, get consensusPeerStore error: %v", err)
	}

	if consensusPeerStore == nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, can not find any record")
	}
	consensusPeerBytes, err := cstates.GetValueFromRawStorageItem(consensusPeerStore)
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize from raw storage item err:%v", err)
	}
	consensusPeers := new(ConsensusPeers)
	if err := consensusPeers.Deserialization(common.NewZeroCopySource(consensusPeerBytes)); err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize consensusPeer error: %v", err)
	}
	return consensusPeers, nil
}

func putConsensusPeers(native *native.NativeService, consensusPeers *ConsensusPeers) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	consensusPeers.Serialization(sink)
	chainIDBytes := utils.GetUint64Bytes(consensusPeers.ChainID)
	heightBytes := utils.GetUint32Bytes(consensusPeers.Height)
	blockHeightBytes := utils.GetUint32Bytes(native.GetHeight())

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER), chainIDBytes, heightBytes), cstates.GenRawStorageItem(sink.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER_BLOCK_HEIGHT), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(blockHeightBytes))

	//update key heights
	keyHeights, err := GetKeyHeights(native, consensusPeers.ChainID)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, GetKeyHeights error: %v", err)
	}

	keyHeights.HeightList = append(keyHeights.HeightList, consensusPeers.Height)
	err = PutKeyHeights(native, consensusPeers.ChainID, keyHeights)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, putKeyHeights error: %v", err)
	}
	return nil
}

func UpdateConsensusPeer(native *native.NativeService, chainID uint64, header *otypes.Header) error {
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(header.ConsensusPayload, blkInfo); err != nil {
		return fmt.Errorf("updateConsensusPeer, unmarshal blockInfo error: %s", err)
	}
	if blkInfo.NewChainConfig != nil {
		consensusPeers := &ConsensusPeers{
			ChainID: chainID,
			Height:  header.Height,
			PeerMap: make(map[string]*Peer),
		}
		for _, p := range blkInfo.NewChainConfig.Peers {
			consensusPeers.PeerMap[p.ID] = &Peer{Index: p.Index, PeerPubkey: p.ID}
		}
		err := putConsensusPeers(native, consensusPeers)
		if err != nil {
			return fmt.Errorf("updateConsensusPeer, put ConsensusPeer error: %s", err)
		}
	}
	return nil
}

func FindKeyHeight(native *native.NativeService, height uint32, chainID uint64) (uint32, error) {
	keyHeights, err := GetKeyHeights(native, chainID)
	if err != nil {
		return 0, fmt.Errorf("findKeyHeight, GetKeyHeights error: %v", err)
	}
	for _, v := range keyHeights.HeightList {
		if height > v {
			return v, nil
		}
	}
	return 0, fmt.Errorf("findKeyHeight, can not find key height with height %d", height)
}
