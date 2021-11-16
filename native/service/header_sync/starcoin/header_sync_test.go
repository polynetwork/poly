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

package starcoin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/genesis"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	stc "github.com/starcoinorg/starcoin-go/client"
	stctypes "github.com/starcoinorg/starcoin-go/types"
	//stcutils "github.com/starcoinorg/starcoin-go/utils"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const (
	SUCCESS = iota
	GENESIS_PARAM_ERROR
	GENESIS_INITIALIZED
	SYNCBLOCK_PARAM_ERROR
	SYNCBLOCK_ORPHAN
	DIFFICULTY_ERROR
	NONCE_ERROR
	OPERATOR_ERROR
	UNKNOWN
)

const MainHeaderJson = `
	{
      "block_hash": "0x80848150abee7e9a3bfe9542a019eb0b8b01f124b63b011f9c338fdb935c417d",
      "parent_hash": "0xb82a2c11f2df62bf87c2933d0281e5fe47ea94d5f0049eec1485b682df29529a",
      "timestamp": "1621311100863",
      "number": "0",
      "author": "0x00000000000000000000000000000001",
      "author_auth_key": null,
      "txn_accumulator_root": "0x43609d52fdf8e4a253c62dfe127d33c77e1fb4afdefb306d46ec42e21b9103ae",
      "block_accumulator_root": "0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000",
      "state_root": "0x61125a3ab755b993d72accfea741f8537104db8e022098154f3a66d5c23e828d",
      "gas_used": "0",
      "difficulty": "0xb1ec37",
      "body_hash": "0x7564db97ee270a6c1f2f73fbf517dc0777a6119b7460b7eae2890d1ce504537b",
      "chain_id": 1,
      "nonce": 0,
      "extra": "0x00000000"
	  }
	`

var (
	acct     = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
)

func init() {
	setBKers()
}

func typeOfError(e error) int {
	if e == nil {
		return SUCCESS
	}
	errDesc := e.Error()
	if strings.Contains(errDesc, "STCHandler GetHeaderByHeight, genesis header had been initialized") {
		return GENESIS_INITIALIZED
	} else if strings.Contains(errDesc, "STCHandler SyncGenesisHeader: getGenesisHeader, deserialize header err:") {
		return GENESIS_PARAM_ERROR
	} else if strings.Contains(errDesc, "SyncBlockHeader, deserialize header err:") {
		return SYNCBLOCK_PARAM_ERROR
	} else if strings.Contains(errDesc, "SyncBlockHeader, get the parent block failed. Error:") {
		return SYNCBLOCK_ORPHAN
	} else if strings.Contains(errDesc, "SyncBlockHeader, invalid difficulty:") {
		return DIFFICULTY_ERROR
	} else if strings.Contains(errDesc, "SyncBlockHeader, verify header error:") {
		return NONCE_ERROR
	} else if strings.Contains(errDesc, "SyncGenesisHeader, checkWitness error:") {
		return OPERATOR_ERROR
	}
	return UNKNOWN
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		sink := common.NewZeroCopySink(nil)
		view := &node_manager.GovernanceView{
			TxHash: common.UINT256_EMPTY,
			Height: 0,
			View:   0,
		}
		view.Serialization(sink)
		db.Put(utils.ConcatKey(utils.NodeManagerContractAddress, []byte(node_manager.GOVERNANCE_VIEW)), states.GenRawStorageItem(sink.Bytes()))

		peerPoolMap := &node_manager.PeerPoolMap{
			PeerPoolMap: map[string]*node_manager.PeerPoolItem{
				vconfig.PubkeyID(acct.PublicKey): {
					Address:    acct.Address,
					Status:     node_manager.ConsensusStatus,
					PeerPubkey: vconfig.PubkeyID(acct.PublicKey),
					Index:      0,
				},
			},
		}
		sink.Reset()
		peerPoolMap.Serialization(sink)
		db.Put(utils.ConcatKey(utils.NodeManagerContractAddress,
			[]byte(node_manager.PEER_POOL), utils.GetUint32Bytes(0)), states.GenRawStorageItem(sink.Bytes()))

	}
	ret, _ := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	return ret
}

func getLatestHeight(native *native.NativeService) uint64 {
	contractAddress := utils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(0)...)
	// try to get storage
	result, err := native.GetCacheDB().Get(utils.ConcatKey(contractAddress, key))
	if err != nil {
		return 0
	}
	if result == nil || len(result) == 0 {
		return 0
	} else {
		heightBytes, _ := states.GetValueFromRawStorageItem(result)
		return binary.LittleEndian.Uint64(heightBytes)
	}
}

func getHeaderHashByHeight(native *native.NativeService, height uint64) stctypes.HashValue {
	headerStore, _ := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(1), utils.GetUint64Bytes(height)))
	hashBytes, _ := states.GetValueFromRawStorageItem(headerStore)
	hash, _ := stctypes.BcsDeserializeHashValue(hashBytes)
	return hash
}

func getHeaderByHash(native *native.NativeService, headHash stctypes.HashValue) []byte {
	headHashBytes, _ := headHash.BcsSerialize()
	headerStore, _ := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(1), headHashBytes))
	headerBytes, err := states.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return headerBytes
}

func TestSyncGenesisHeader(t *testing.T) {
	var headerBytes = []byte(MainHeaderJson)
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = 1
	param.GenesisHeader = headerBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native := NewNative(sink.Bytes(), tx, nil)
	STCHandler := NewSTCHandler()
	err := STCHandler.SyncGenesisHeader(native)
	assert.Equal(t, SUCCESS, typeOfError(err))
	height := getLatestHeight(native)
	assert.Equal(t, uint64(0), height)
	headerHash := getHeaderHashByHeight(native, 0)
	headerFormStore := getHeaderByHash(native, headerHash)
	header, _ := stctypes.BcsDeserializeBlockHeader(headerFormStore)
	var jsonHeader stc.BlockHeader
	json.Unmarshal(headerBytes, &jsonHeader)
	assert.Equal(t, header, jsonHeader.ToTypesHeader())
}

func TestSyncGenesisHeaderTwice(t *testing.T) {
	STCHandler := NewSTCHandler()
	var native *native.NativeService
	{
		var headerBytes = []byte(MainHeaderJson)
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 1
		param.GenesisHeader = headerBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, nil)
		err := STCHandler.SyncGenesisHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
	{
		var headerBytes = []byte(MainHeaderJson)
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 1
		param.GenesisHeader = headerBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		err := STCHandler.SyncGenesisHeader(native)
		assert.Equal(t, GENESIS_INITIALIZED, typeOfError(err))
	}
}
