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

package cosmos

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"math"
)

type KeyHeights struct {
	HeightList []int64
}

func (this *KeyHeights) Serialization(sink *common.ZeroCopySink) {
	//first sort the list  (big -> small)
	sink.WriteVarUint(uint64(len(this.HeightList)))
	for _, v := range this.HeightList {
		sink.WriteInt64(v)
	}
}

func (this *KeyHeights) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize HeightList length error")
	}
	heightList := make([]int64, 0)
	for i := 0; uint64(i) < n; i++ {
		height, eof := source.NextInt64()
		if eof {
			return fmt.Errorf("utils.DecodeVarUint, deserialize height error")
		}
		if height > math.MaxInt64 {
			return fmt.Errorf("deserialize height error: height more than max uint32")
		}
		heightList = append(heightList, height)
	}
	this.HeightList = heightList
	return nil
}

func (this *KeyHeights) AddNewHeight(height int64) {
	this.HeightList = append(this.HeightList, height)
	i := len(this.HeightList) - 1
	for ;i >0;i -- {
		if this.HeightList[i] < this.HeightList[i - 1] {
			this.HeightList[i], this.HeightList[i - 1] = this.HeightList[i - 1], this.HeightList[i]
		} else {
			break
		}
	}
}

func (this *KeyHeights) FindKeyHeight(height int64) (int64, error) {
	i := len(this.HeightList) - 1
	for ;i >= 0;i -- {
		if height > this.HeightList[i] {
			return this.HeightList[i], nil
		}
	}
	return 0, fmt.Errorf("findKeyHeight, can not find key height with height %d", height)
}

func putGenesisHeader(native *native.NativeService, cdc *codec.Codec, header *CosmosHeader, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	headerBs, err := cdc.MarshalBinaryBare(header)
	if err != nil {
		return err
	}
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(utils.GetUint64Bytes(uint64(header.Header.Height))))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(uint64(header.Header.Height))),
		cstates.GenRawStorageItem(headerBs))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(utils.GetUint64Bytes(uint64(header.Header.Height))))
	notifyPutHeader(native, chainID, uint64(header.Header.Height), header.Header.Hash().String())
	return nil
}

func hasGenesis(native *native.NativeService, chainID uint64) (bool, error) {
	contract := utils.HeaderSyncContractAddress
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.GENESIS_HEADER), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return false, err
	}
	if headerStore != nil {
		return true, nil
	}
	return false, nil
}

func getCurrentHeight(native *native.NativeService, chainID uint64) (int64, error) {
	contract := utils.HeaderSyncContractAddress
	heightStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return 0, fmt.Errorf("getCurrentHeight, get blockHashStore error: %v", err)
	}
	if heightStore == nil {
		return 0, fmt.Errorf("getCurrentHeight, genesis header had been initialized")
	}
	heightBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		return 0, fmt.Errorf("getCurrentHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return int64(utils.GetBytesUint64(heightBytes)), nil
}

func GetHeaderByHeight(native *native.NativeService, cdc *codec.Codec, height int64, chainID uint64) (*CosmosHeader, error) {
	contract := utils.HeaderSyncContractAddress
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(uint64(height))))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHeight, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	var header CosmosHeader
	err = cdc.UnmarshalBinaryBare(storeBytes, &header)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize header error: %v", err)
	}
	return &header, nil
}

func getCurrentHeader(native *native.NativeService, cdc *codec.Codec, chainID uint64) (*CosmosHeader, error) {
	height, err := getCurrentHeight(native, chainID)
	if err != nil {
		return nil, err
	}
	return GetHeaderByHeight(native, cdc, height, chainID)
}

func putHeader(native *native.NativeService, cdc *codec.Codec, chainID uint64, header *CosmosHeader) error {
	contract := utils.HeaderSyncContractAddress
	headerBs, err := cdc.MarshalBinaryBare(header)
	if err != nil {
		return err
	}
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(uint64(header.Header.Height))),
		cstates.GenRawStorageItem(headerBs))

	currentHeight, _ := getCurrentHeight(native, chainID)
	if currentHeight < header.Header.Height {
		native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)),
			cstates.GenRawStorageItem(utils.GetUint64Bytes(uint64(header.Header.Height))))
	}
	notifyPutHeader(native, chainID, uint64(header.Header.Height), header.Header.Hash().String())
	return nil
}

func notifyPutHeader(native *native.NativeService, chainID uint64, height uint64, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, blockHash, native.GetHeight()},
		})
}

func FindKeyHeight(native *native.NativeService, height int64, chainID uint64) (int64, error) {
	keyHeights, err := GetKeyHeights(native, chainID)
	if err != nil {
		return 0, fmt.Errorf("findKeyHeight, GetKeyHeights error: %v", err)
	}
	i := len(keyHeights.HeightList) - 1
	for ;i >= 0;i -- {
		if height > keyHeights.HeightList[i] {
			return keyHeights.HeightList[i], nil
		}
	}
	return 0, fmt.Errorf("findKeyHeight, can not find key height with height %d", height)
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
