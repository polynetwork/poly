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
	_ "bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/holiman/uint256"
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
const Header2810119 = `
	{
      "block_hash": "0x00ab900bc2841effa4a52ff06e6aa4a090f2482cc8090bc3a3ff6519eed156da",
      "parent_hash": "0xa382474d0fd1270f7f98f2bdbd17deaffb14a69d7ba8fd060a032e723f997b4b",
      "timestamp": "1637063089399",
      "number": "2810119",
      "author": "0x3b8ebb9e889f8df0b603d8d9f3f05524",
      "author_auth_key": null,
      "txn_accumulator_root": "0x57736acacaeca3c1f391b9d1a2965191099e8e9b4533d8d9e6fe97915a746ad1",
      "block_accumulator_root": "0x282d6399a2581f3319207c17bdeeefdd3066a908a7c0c0c81541b3527c4a7f47",
      "state_root": "0x96a472a42d0b62fd4daa48e71b06e61637bfd6561b10c5864351cd6d3ef42273",
      "gas_used": "0",
      "difficulty": "0x0daecc86",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 1,
      "nonce": 255857088,
      "extra": "0x163a0000"
	  }
	`
const Header2810120 = `
	{
      "block_hash": "0x24ae68e92470c9d99391d7958f540f6e9fcd9c3d0d2ad8e5b036368a666f4ffb",
      "parent_hash": "0x00ab900bc2841effa4a52ff06e6aa4a090f2482cc8090bc3a3ff6519eed156da",
      "timestamp": "1637063096993",
      "number": "2810120",
      "author": "0x707d8fc016acae0a1a859769ad0c4fcf",
      "author_auth_key": null,
      "txn_accumulator_root": "0x82a4dfdb5b40fea2bd092f2b3904479e14b2b71e912dfcb76ebed30efc1c5584",
      "block_accumulator_root": "0x1b4333a094917ecf21f1240073867b5b1065bf2f4bdfbb1b614e866ae94d92c8",
      "state_root": "0x67286c6c4df5ac7bd8f5c2a03866afb64e289fd20a661e0c1663d9a18d37bf8a",
      "gas_used": "0",
      "difficulty": "0x0e9d5bc8",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 1,
      "nonce": 184550366,
      "extra": "0x440037ac"
	  }
	`
const Header2810121 = `
	{
      "block_hash": "0x200d5603b68a26a55cc311248a3e4370c5748768f526966bc4633eea9ff28b2a",
      "parent_hash": "0x24ae68e92470c9d99391d7958f540f6e9fcd9c3d0d2ad8e5b036368a666f4ffb",
      "timestamp": "1637063098995",
      "number": "2810121",
      "author": "0x46a1d0101f491147902e9e00305107ed",
      "author_auth_key": null,
      "txn_accumulator_root": "0xde469f61a7a9aaddded00297a4bd4101dd46a6541786970f01177cfe8630ec03",
      "block_accumulator_root": "0x1a95612238fa9544301b2b51df9e8db7256bd85f964584053aab380041c91d84",
      "state_root": "0x9349e1176728726d5ff8ef66e9046a1806c2b91cb167a356b995155f9b2a65d4",
      "gas_used": "0",
      "difficulty": "0x0e4d2c5a",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 1,
      "nonce": 67112105,
      "extra": "0x14e10000"
	  }
	`
const Header2810122 = `
	{
      "block_hash": "0x6c804f42ae88460a18d2a1e459956892f1d4803d15e15927d9c05638f40b1bc3",
      "parent_hash": "0x200d5603b68a26a55cc311248a3e4370c5748768f526966bc4633eea9ff28b2a",
      "timestamp": "1637063103223",
      "number": "2810122",
      "author": "0x46a1d0101f491147902e9e00305107ed",
      "author_auth_key": null,
      "txn_accumulator_root": "0x39b9dfeca0527869199ab0c9808836547b8a5e33cc6236b6407731c3838b1aa2",
      "block_accumulator_root": "0x021ab5cf63572189bd02860afc2af05bf72e60a5eb3877af378c6cfc46b2b516",
      "state_root": "0xfc1fa45e690f7cdf4a76dee9953bde31511ccfc339622fae5486bd7f04875ce0",
      "gas_used": "0",
      "difficulty": "0x0f237608",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 1,
      "nonce": 67112918,
      "extra": "0x5e730000"
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

func packageHeader(header string, timeTarget uint64) string {
	var build strings.Builder
	build.WriteString("{\"header\":")
	build.WriteString(header)
	build.WriteString(", \"block_time_target\":")
	build.WriteString(fmt.Sprint(timeTarget))
	build.WriteString(",\"block_difficulty_window\":24}")
	return build.String()
}

func getLatestHeight(native *native.NativeService) uint64 {
	contractAddress := utils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(1)...)
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
	return hashBytes
}

func getHeaderByHash(native *native.NativeService, headHash *stctypes.HashValue) []byte {
	headerStore, _ := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(1), *headHash))
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
	headerFormStore := getHeaderByHash(native, &headerHash)
	header, _ := stctypes.BcsDeserializeBlockHeader(headerFormStore)
	var jsonHeader stc.BlockHeader
	json.Unmarshal(headerBytes, &jsonHeader)
	headerNew, _ := jsonHeader.ToTypesHeader()
	assert.Equal(t, header, *headerNew)
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

func TestSyncHeader(t *testing.T) {
	STCHandler := NewSTCHandler()
	var native *native.NativeService
	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	{
		header2810118, _ := hex.DecodeString("0a097b0a20202020202022626c6f636b5f68617368223a2022307861333832343734643066643132373066376639386632626462643137646561666662313461363964376261386664303630613033326537323366393937623462222c0a20202020202022706172656e745f68617368223a2022307835366533336232353737353933306534396264356230353338323838313835343063633136373934653232653531616437313333646439336363373533343136222c0a2020202020202274696d657374616d70223a202231363337303633303838313635222c0a202020202020226e756d626572223a202232383130313138222c0a20202020202022617574686f72223a202230783436613164303130316634393131343739303265396530303330353130376564222c0a20202020202022617574686f725f617574685f6b6579223a206e756c6c2c0a2020202020202274786e5f616363756d756c61746f725f726f6f74223a2022307832313138386333346634316237643865383039386666643239313761346664373638613064626466623033643130306166303964376263313038643066363037222c0a20202020202022626c6f636b5f616363756d756c61746f725f726f6f74223a2022307834666532633133306430316234393863643666346232303365633239373865663138393036653132656539326463663664613536346437653534613063363330222c0a2020202020202273746174655f726f6f74223a2022307862653564323332376338666632633831363435623734323661663061343032393739616565336163323136383534313230396633383036633534653464363037222c0a202020202020226761735f75736564223a202230222c0a20202020202022646966666963756c7479223a202230783063653737366237222c0a20202020202022626f64795f68617368223a2022307863303165303332396465366438393933343861386566346264353164623536313735623366613039383865353763336463656338656166313361313634643937222c0a20202020202022636861696e5f6964223a20312c0a202020202020226e6f6e6365223a20313234393930323836352c0a202020202020226578747261223a202230783634336230303030220a0920207d0a09")
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 1
		param.GenesisHeader = header2810118
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), tx, nil)
		err := STCHandler.SyncGenesisHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err))

		height := getLatestHeight(native)
		assert.Equal(t, uint64(2810118), height)
		headerHash := getHeaderHashByHeight(native, 2810118)
		headerFormStore := getHeaderByHash(native, &headerHash)
		header, _ := stctypes.BcsDeserializeBlockHeader(headerFormStore)
		var jsonHeader stc.BlockHeader
		json.Unmarshal(header2810118, &jsonHeader)
		headerNew, _ := jsonHeader.ToTypesHeader()
		assert.Equal(t, header, *headerNew)
	}
	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 1
		param.Address = acct.Address
		param.Headers = append(param.Headers, []byte(packageHeader(Header2810119, 5918)))
		param.Headers = append(param.Headers, []byte(packageHeader(Header2810120, 5918)))
		param.Headers = append(param.Headers, []byte(packageHeader(Header2810121, 5918)))
		param.Headers = append(param.Headers, []byte(packageHeader(Header2810122, 5918)))
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		err := STCHandler.SyncBlockHeader(native)
		if err != nil {
			t.Fatal("SyncBlockHeader", err)
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
}

const HalleyHeaders = `
[
  {
    "timestamp": "1639375200198",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0xfa55091e7f19023cd70d55bc147c194d09649585ac90cade4898302530c50bda",
    "block_hash": "0xb6c0a3c14df4133e5ce8b89f7adff3add41e1df10b818da39c8eab54f26225cb",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x80",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3108099670,
    "number": "222625",
    "parent_hash": "0xf976fea99030c3442508b6deac2596b338d9dc9d3a2bcc886ebed1bcd70b1fce",
    "state_root": "0xa0f7a539ecaeabe08e47ba2a11e698684f75db18e623cacbd4dd83724bf4a945",
    "txn_accumulator_root": "0x0b4bbaefcb7a509b32ae41681b39ad6e4917e79220aa2883d6b995b7f94b55c0"
  },
  {
    "timestamp": "1639375190723",
    "author": "0x57aa381a5d7c0141da3965393eed9958",
    "author_auth_key": null,
    "block_accumulator_root": "0xb654635a9435e9c3526a9edc7cd6904173b5d8942c3ba521ee3595077aa9f961",
    "block_hash": "0xf976fea99030c3442508b6deac2596b338d9dc9d3a2bcc886ebed1bcd70b1fce",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x7f",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 405931573,
    "number": "222624",
    "parent_hash": "0xe9a60ae37dbdd9853127fa4009caa629062db56db7756f41a337302d1cd7b0a0",
    "state_root": "0x4b6d85eb6f97758234ac8dbad49d8c7f41864a645c1afbc190a9c7a8fa140a2c",
    "txn_accumulator_root": "0x59d489f529ae157669d48ce63f2af54d3d758bfaf299b1e2a23d991e24d9dd59"
  },
  {
    "timestamp": "1639375186680",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x3119257f80a50d54da8c9caa9037f3ab36b6e8e9c0417bb9129383e445f67304",
    "block_hash": "0xe9a60ae37dbdd9853127fa4009caa629062db56db7756f41a337302d1cd7b0a0",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x7e",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3117778102,
    "number": "222623",
    "parent_hash": "0xdcac7d9317aebc10b18e011d2050ea768b98c9ea552855c46535a38af165a81e",
    "state_root": "0x6d16b2e6b2c48b38da9f3072c06e5063e19fa0e8fbdc3313b338594161c31172",
    "txn_accumulator_root": "0x108c3a2240bca50e818d5cd28b4659468628d02fa8089a8ed6033771d52e9d1b"
  },
  {
    "timestamp": "1639375182033",
    "author": "0x57aa381a5d7c0141da3965393eed9958",
    "author_auth_key": null,
    "block_accumulator_root": "0x562b9f7a2f8a6101e034f5be3efab4d7b907b046816f5d3dee679fc8b6512543",
    "block_hash": "0xdcac7d9317aebc10b18e011d2050ea768b98c9ea552855c46535a38af165a81e",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x82",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3929424765,
    "number": "222622",
    "parent_hash": "0x87318b8fa9507f4069dac0a090c44bb7c75278a105108d674cdd73b0736249d0",
    "state_root": "0xd8dea7200f3204147e68810f033d0d2496261cd510244b4056b67fac4fa85258",
    "txn_accumulator_root": "0xd59d7849e84832c3a7e0386f38dcb97ab85d9ddba99a088c8da914756cafa48e"
  },
  {
    "timestamp": "1639375175417",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x48e2711239bc8f2e233734becc494d48e536a5552978cce975321ff9fb940b48",
    "block_hash": "0x87318b8fa9507f4069dac0a090c44bb7c75278a105108d674cdd73b0736249d0",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x8a",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3589097564,
    "number": "222621",
    "parent_hash": "0x2ea3f9b56cf25516b05ef8c81080a8787f198ea52a3f2d6b8bbc7f7df9484de1",
    "state_root": "0x6a80917148af7b1f97fce1476de4529d28b2bbed173646d94d55b5ee8db9d7bb",
    "txn_accumulator_root": "0x58fbffaa10d0753769b36ccf81a708947d44f798d282c5da5a5ab8202e1e5405"
  },
  {
    "timestamp": "1639375166231",
    "author": "0x57aa381a5d7c0141da3965393eed9958",
    "author_auth_key": null,
    "block_accumulator_root": "0x38d89cd983151a19b789615d1d77bb83b15b11641af6636e18359820ea375c42",
    "block_hash": "0x2ea3f9b56cf25516b05ef8c81080a8787f198ea52a3f2d6b8bbc7f7df9484de1",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xb6",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3995232715,
    "number": "222620",
    "parent_hash": "0x7609c99847446eb5adb81cb71066b11d53bdbb1ceb0b010ade23db6ffe9a9761",
    "state_root": "0xa53f85a258204d699ef86d4ded28fd0cff49e6c26b1f4753c1994deac40b9943",
    "txn_accumulator_root": "0x3540f761e76af81fbc524c44ba86d38d5b54fadcc4df631ff283dbe123224909"
  },
  {
    "timestamp": "1639375144937",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0xf92f166c0b6d96d407ea6038d8c09b1f753811bf642cfb5fed18efe1b058998b",
    "block_hash": "0x7609c99847446eb5adb81cb71066b11d53bdbb1ceb0b010ade23db6ffe9a9761",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xc0",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 257311134,
    "number": "222619",
    "parent_hash": "0x6fbbfd5b05417b4d8a4f1f0bf5e78c54a7772389d0de350a873259e60e68d1f4",
    "state_root": "0x3742a5b4025bdb6f6730ae0dff448aa893317a1e065383e6f842f1bc5ed6cd55",
    "txn_accumulator_root": "0x059e53fec0fbb8de2d9d88ec6c3c6031afc26cb47b453cf48723cc0d1b316200"
  },
  {
    "timestamp": "1639375137558",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x67ecc0d31cbaaf03502922f108621d8e9081926a5ba7edcabd4df798f0a49dc0",
    "block_hash": "0x6fbbfd5b05417b4d8a4f1f0bf5e78c54a7772389d0de350a873259e60e68d1f4",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xcd",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 371246835,
    "number": "222618",
    "parent_hash": "0x5171747b92d12c774b5a59f2cf4e7ee20a74fbb6c07d6d768a7cf8b2bdfea15b",
    "state_root": "0x9fd0030095e1ac2b3b581fee4db027a0fe24070b42b357f6287f26b9dab8a775",
    "txn_accumulator_root": "0x728ba8be7e4e5f716aa1aa50b69947085cae727f2b1700387f2c30e17a594cc6"
  },
  {
    "timestamp": "1639375129589",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0x325407fddbcfa599dc053a71582a30f6490c6a0a6d991b765d8ca9a7e9389797",
    "block_hash": "0x5171747b92d12c774b5a59f2cf4e7ee20a74fbb6c07d6d768a7cf8b2bdfea15b",
    "body_hash": "0x94f4be06edbb008010ada171280a7c9033e3f9575eb04ca12425fbdf14073195",
    "chain_id": 253,
    "difficulty": "0xc2",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3822228938,
    "number": "222617",
    "parent_hash": "0xe31fc863649f14038540011ec1a11197e5aea0c0fdd96a6aa2ab776f5b84aa26",
    "state_root": "0xeb2f7cd7f95ca2d56c665690959ca45560ebed3a88f37c77733de21bc8a67463",
    "txn_accumulator_root": "0xd40a660232eca511c3720c20046cdd556f821255b45b2acd8958617baa0e78d7"
  },
  {
    "timestamp": "1639375126993",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x520b666e8db1f5698e0a3361e6d1971812add9e3fe01e9cb638749b60e9fb166",
    "block_hash": "0xe31fc863649f14038540011ec1a11197e5aea0c0fdd96a6aa2ab776f5b84aa26",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xc7",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 1505730553,
    "number": "222616",
    "parent_hash": "0x2962e0b78133927214142792fad95964efbdc90bec74d16c827044b26f0cdea2",
    "state_root": "0xf0e1adb4e52af061f38534bfd7b795a0e5d257c90d2ad39620b63916120fa743",
    "txn_accumulator_root": "0x6e8d04ee7c90f0f62cb83f489a990f93203746a04f639961bb6791ba456a55f2"
  },
  {
    "timestamp": "1639375120947",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0x68daa7ef9f491e3727283563dfaafac5cb3257f7f18c624ec56c4350e0ad0160",
    "block_hash": "0x2962e0b78133927214142792fad95964efbdc90bec74d16c827044b26f0cdea2",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xb7",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 2004121221,
    "number": "222615",
    "parent_hash": "0x67fa6612cd950ee17bb54774ccdba721a08894e26d4919b3fcc86a56e78b77a4",
    "state_root": "0x01234b4cc613a66fd955449212eb239b7c4905d5bd02234af1b248fdff245b27",
    "txn_accumulator_root": "0xdcf698ee2d31c0833c5ff32a52ffbb23f1c123711bfb8f4a090486b978ed26c0"
  },
  {
    "timestamp": "1639375119910",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0xa0786819527743baf188097fb42a8761f16219f874c9971a5e094aa57a63a7a3",
    "block_hash": "0x67fa6612cd950ee17bb54774ccdba721a08894e26d4919b3fcc86a56e78b77a4",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xb6",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 1683927075,
    "number": "222614",
    "parent_hash": "0x593f7c20d57d4aca9c79d653386074681f2833360c7b8644afcabac7390f85c3",
    "state_root": "0x4d6c2e3870afcdf53c8756017386a875ef27335da8ab321ad1c0bf48ce4ec6d0",
    "txn_accumulator_root": "0xb7a79864daa4a23c701c2d5cd14dbcbf9c54384fb66f3fe2ebd5714edefb02a6"
  },
  {
    "timestamp": "1639375115007",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0xe9ec2469dff17c02bfbcce9cc36c097ca37158f6f44571fe3e4e6474824ad087",
    "block_hash": "0x593f7c20d57d4aca9c79d653386074681f2833360c7b8644afcabac7390f85c3",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xab",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3344799031,
    "number": "222613",
    "parent_hash": "0x192ee812d71deeb75e9e408c09d2d520ecbdd2708b273d1cf91f6d58a688f88f",
    "state_root": "0xb89eea543f0b7c9be85f9617a6f790884624c5605ad0dde322e0ee6fddfe2afe",
    "txn_accumulator_root": "0x7862d543480cc8e1c6c07f15b773ba0b27171c60f5243da852b394b20da8d4b6"
  },
  {
    "timestamp": "1639375113374",
    "author": "0x57aa381a5d7c0141da3965393eed9958",
    "author_auth_key": null,
    "block_accumulator_root": "0xec2ededc7c136a2a5ccae0e2229cb8bba1f3266171a654f2c5129c729ae583f3",
    "block_hash": "0x192ee812d71deeb75e9e408c09d2d520ecbdd2708b273d1cf91f6d58a688f88f",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0xa2",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 1737789817,
    "number": "222612",
    "parent_hash": "0x9eef8621742f87cd1f7652571d906374f1575a5677e5fc696aa8c671ab9eb988",
    "state_root": "0xd8297e63cfbe4dcfc136e885e82e3f71609557a5d7a9667b10b1efe436a6caf6",
    "txn_accumulator_root": "0x5901464f95d014649d2748228ae05bf0ac9b079f5c5856e309c274e7e78f15fa"
  },
  {
    "timestamp": "1639375111048",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0x9140a23be1c8d9da1af6e170a205b882cec0437584a7895835cbcd33782e4df2",
    "block_hash": "0x9eef8621742f87cd1f7652571d906374f1575a5677e5fc696aa8c671ab9eb988",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x96",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3100801417,
    "number": "222611",
    "parent_hash": "0xd676f6b003ea710ec530866a986ded318d9d0f6ce94817a69bc1c14960f9bc92",
    "state_root": "0x57492bc896f4e4a2b2a2a4d3d9b7b35d7ab79f823604780e6509a95a6f2a2a37",
    "txn_accumulator_root": "0x30e0f6b0d45997c6c2bda7c4e92bfe18abe7178f43bc303f0952e1e33c59f4d0"
  },
  {
    "timestamp": "1639375110533",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x545380b73c03905ac6bd2d969922963216f92432def9ba1610cf07c8401d3bfa",
    "block_hash": "0xd676f6b003ea710ec530866a986ded318d9d0f6ce94817a69bc1c14960f9bc92",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x92",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 2478486882,
    "number": "222610",
    "parent_hash": "0x218ccbee42c8cbe110f856127dfe1138d00f4274dc112aeac7f6496e11548e16",
    "state_root": "0x238d5d83c99ce0ead518cba907bdc32adeec7916b4e63342c91de8937c3b7ee4",
    "txn_accumulator_root": "0xb1a42aeb8bd66ab01ba215f1260d7f83c5a90febd5cff17a15aa3e02268eccb1"
  },
  {
    "timestamp": "1639375107398",
    "author": "0x57aa381a5d7c0141da3965393eed9958",
    "author_auth_key": null,
    "block_accumulator_root": "0xa0be27fa000207e714185d859eafe47535784067a45c2496994ad7ed78264fbc",
    "block_hash": "0x218ccbee42c8cbe110f856127dfe1138d00f4274dc112aeac7f6496e11548e16",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x96",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 96394581,
    "number": "222609",
    "parent_hash": "0x3b1c57e13cd123c17a69f45d5054a8e111cbc0e37063c0db269237194a135aa6",
    "state_root": "0x01241b6f942f93022875087395ec2af74404a1d803ab160ef4ab968143471ce5",
    "txn_accumulator_root": "0x8325579b81e5091dbf3ea630caff62e45de82d833e92703df02145ff35fc9f8c"
  },
  {
    "timestamp": "1639375100507",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x5bd929e964aa0ad2db5d49bf421f93ebcf5043ac0e738770cf34725ae159381f",
    "block_hash": "0x3b1c57e13cd123c17a69f45d5054a8e111cbc0e37063c0db269237194a135aa6",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x8e",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 2570314818,
    "number": "222608",
    "parent_hash": "0x48998783dc3663562d29a32ea8d1da31c1c58668dca4c7aa9bb7d689216c6d03",
    "state_root": "0x22c0438f695548805d387516ef7483504e343b7f14b6cb4146c387e9814b443c",
    "txn_accumulator_root": "0xbe505e9d770dbf51f2afe63b9a49edc46c54187d76f309e73097111be10b0ce6"
  },
  {
    "timestamp": "1639375099200",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0x58fa84e521fa4796271c29a02721485e43002e4fb08ce327337f2e80092bd047",
    "block_hash": "0x48998783dc3663562d29a32ea8d1da31c1c58668dca4c7aa9bb7d689216c6d03",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x91",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 2080679222,
    "number": "222607",
    "parent_hash": "0x809494e63abe1f88bd853d48d52dfb5c1e55de628946564f40e92fcd26dfac6c",
    "state_root": "0x8e2939f7c948cb159ee0cc5cb3848bb453b54a78772e3172aa6f5604955df916",
    "txn_accumulator_root": "0x867e377f0d61e91750bfaef42a3eafce2388097bf09b7f630478e7c6775871ff"
  },
  {
    "timestamp": "1639375093065",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0xd22bece4f79cda8a5312e060a5dee16c30484392d34c38d41bb6d601cab17db2",
    "block_hash": "0x809494e63abe1f88bd853d48d52dfb5c1e55de628946564f40e92fcd26dfac6c",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x98",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 389243955,
    "number": "222606",
    "parent_hash": "0x24d4c2e622bed8ac215e06736c02df6ddbb9cbed54655ed9daed111dff814f63",
    "state_root": "0xcc0dea1701fd02e3c986c858c53bc7db1b15a543d547e3284a37adc938580609",
    "txn_accumulator_root": "0x6d026419d9ffc13a5f4c0b38e27aaabc31dd2c9a14dc7c892f3ca680f2071199"
  },
  {
    "timestamp": "1639375085160",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0xb22f26b79894e029b3de9cf397eaecb2c7761ba5e51b5b6b79e7a696833a8993",
    "block_hash": "0x24d4c2e622bed8ac215e06736c02df6ddbb9cbed54655ed9daed111dff814f63",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x90",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3483997107,
    "number": "222605",
    "parent_hash": "0x5f8484f8865c1d5f51662477fb97450f99d73e45d18d6be6107a3e36296ceb0f",
    "state_root": "0xd9934cc66ee99955d79873ea54f12b89e9d8266ccede5869024fe8fd673d8df1",
    "txn_accumulator_root": "0x44c4519da3b43f4f373ad2c51bda9fc2ab89342b97598ee5b11eaf0e50585bab"
  },
  {
    "timestamp": "1639375083174",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0xec8c8520395ca6b82e41232c94fbea1411128858520fe7b11c435f7543ecb13e",
    "block_hash": "0x5f8484f8865c1d5f51662477fb97450f99d73e45d18d6be6107a3e36296ceb0f",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x93",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 2687647505,
    "number": "222604",
    "parent_hash": "0x15a078ecfe7d1d5c5172198cd7a648c6df73d1d90c52da4f7b5dbdb3859c604f",
    "state_root": "0xbdb0269992111a05b00bfde5209b9687b9a7e570165d75d8b9eef5e8cdc5893d",
    "txn_accumulator_root": "0x656edb693fc28f5936319046e746422e35c19273c8767d1cabf4a4d15c2850e6"
  },
  {
    "timestamp": "1639375076986",
    "author": "0x00e4ea282432073992bc04ab278ddd60",
    "author_auth_key": null,
    "block_accumulator_root": "0xa2a8334cadfb2730e3a877111cb7f628f1001d224a3b38b39f18963cdffc6edb",
    "block_hash": "0x15a078ecfe7d1d5c5172198cd7a648c6df73d1d90c52da4f7b5dbdb3859c604f",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x97",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 3729981630,
    "number": "222603",
    "parent_hash": "0x983d3a2f794e6778b071043cb1fb8cafcd0238dd8a63a8e4ef600a522a30a422",
    "state_root": "0xeb87d0a72670fb8146507767e5157b58b90a0f7bb52246ae0a49f40bd1171e9d",
    "txn_accumulator_root": "0xcfede1ba91634b2d7519192817611a556f2f4db717d386a9e390504429cdde2e"
  },
  {
    "timestamp": "1639375070408",
    "author": "0xfab981cf1ee57d043be6f4f80b557506",
    "author_auth_key": null,
    "block_accumulator_root": "0x924f6e6b7bea9c7ea77f2a61f9ecc6de901d889268a5892773470fbf9879647a",
    "block_hash": "0x983d3a2f794e6778b071043cb1fb8cafcd0238dd8a63a8e4ef600a522a30a422",
    "body_hash": "0x672eacb8b2d150c4e9b114ac97fead9ed663e58b4296ed8645f5e2f1a65a2915",
    "chain_id": 253,
    "difficulty": "0x8f",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 1608253494,
    "number": "222602",
    "parent_hash": "0x1ca3398b78aa7a59beaab4d9b041f35ab37659592cc8fbe2ad4bf88d9c5f892c",
    "state_root": "0x72df7a40acfcf31de353e7b925a024a9794dc96bb307f5d3b39ec6ac99883119",
    "txn_accumulator_root": "0xd18bbe8a06df89786fb0df8b812f84a538f8c8a20485b53edd1e0331faee4489"
  },
  {
    "timestamp": "1639375068873",
    "author": "0xed0f7fcbc522176bf6c8c42f60419718",
    "author_auth_key": null,
    "block_accumulator_root": "0x537f10ad3eaaafa8ff9b70a2489d86b9720ed01705da57efda0bd9cbd0b23068",
    "block_hash": "0x1ca3398b78aa7a59beaab4d9b041f35ab37659592cc8fbe2ad4bf88d9c5f892c",
    "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
    "chain_id": 253,
    "difficulty": "0x87",
    "difficulty_number": 0,
    "extra": "0x00000000",
    "gas_used": "0",
    "Nonce": 2754674653,
    "number": "222601",
    "parent_hash": "0x44c1eb5f207f8a2e35d669449e9d677a350829472925481852d9d282c7ca8108",
    "state_root": "0x417c1394ea24c7a3ca1088e514319113dd7ddb2c612209fb53da70a14002c7f8",
    "txn_accumulator_root": "0xcee628a7e1deaae1bc30acb3500929a31c858ab7f38f7bf7538a35a8b89b47cb"
  }
]
`

func getWithDifficultyHeader(header stc.BlockHeader) stc.BlockHeaderWithDifficutyInfo {
	info := stc.BlockHeaderWithDifficutyInfo{
		header,
		5000,
		24,
	}
	return info
}

func TestSyncHalleyHeader(t *testing.T) {
	STCHandler := NewSTCHandler()
	var native *native.NativeService
	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	var jsonHeaders []stc.BlockHeader
	json.Unmarshal([]byte(HalleyHeaders), &jsonHeaders)

	{
		genesisHeader, _ := json.Marshal(jsonHeaders[24])
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 1
		param.GenesisHeader = genesisHeader
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), tx, nil)
		err := STCHandler.SyncGenesisHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err))

		height := getLatestHeight(native)
		assert.Equal(t, uint64(222601), height)
		headerHash := getHeaderHashByHeight(native, 222601)
		headerFormStore := getHeaderByHash(native, &headerHash)
		header, _ := stctypes.BcsDeserializeBlockHeader(headerFormStore)
		newHeader, _ := jsonHeaders[24].ToTypesHeader()
		assert.Equal(t, header, *newHeader)
	}
	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 1
		param.Address = acct.Address
		for i := 23; i >= 0; i-- {
			header, _ := json.Marshal(getWithDifficultyHeader(jsonHeaders[i]))
			param.Headers = append(param.Headers, header)
		}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		err := STCHandler.SyncBlockHeader(native)
		if err != nil {
			t.Fatal("SyncBlockHeader", err)
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
}
func TestGetNextTarget(t *testing.T) {
	type args struct {
		blocks   []BlockDiffInfo
		timePlan uint64
	}
	diff0, _ := hex.DecodeString("0109")
	diff1, _ := hex.DecodeString("ef")
	diff2, _ := hex.DecodeString("ec")
	diff3, _ := hex.DecodeString("0103")
	diff4, _ := hex.DecodeString("ea")
	diff5, _ := hex.DecodeString("d4")
	diff6, _ := hex.DecodeString("cf")
	diff7, _ := hex.DecodeString("c5")
	diff8, _ := hex.DecodeString("c4")
	diff9, _ := hex.DecodeString("bc")
	diff10, _ := hex.DecodeString("b1")
	diff11, _ := hex.DecodeString("a4")
	diff12, _ := hex.DecodeString("9a")
	diff13, _ := hex.DecodeString("93")
	diff14, _ := hex.DecodeString("93")
	diff15, _ := hex.DecodeString("8d")
	diff16, _ := hex.DecodeString("8f")
	diff17, _ := hex.DecodeString("90")
	diff18, _ := hex.DecodeString("92")
	diff19, _ := hex.DecodeString("93")
	diff20, _ := hex.DecodeString("8b")
	diff21, _ := hex.DecodeString("8a")
	diff22, _ := hex.DecodeString("af")
	diff23, _ := hex.DecodeString("b4")
	diff24, _ := hex.DecodeString("a9")
	blocks := []BlockDiffInfo{
		BlockDiffInfo{1638331301987, *targetToDiff(new(uint256.Int).SetBytes(diff1))},
		BlockDiffInfo{1638331301564, *targetToDiff(new(uint256.Int).SetBytes(diff2))},
		BlockDiffInfo{1638331297135, *targetToDiff(new(uint256.Int).SetBytes(diff3))},
		BlockDiffInfo{1638331288742, *targetToDiff(new(uint256.Int).SetBytes(diff4))},
		BlockDiffInfo{1638331288188, *targetToDiff(new(uint256.Int).SetBytes(diff5))},
		BlockDiffInfo{1638331287706, *targetToDiff(new(uint256.Int).SetBytes(diff6))},
		BlockDiffInfo{1638331283650, *targetToDiff(new(uint256.Int).SetBytes(diff7))},
		BlockDiffInfo{1638331281477, *targetToDiff(new(uint256.Int).SetBytes(diff8))},
		BlockDiffInfo{1638331276488, *targetToDiff(new(uint256.Int).SetBytes(diff9))},
		BlockDiffInfo{1638331273581, *targetToDiff(new(uint256.Int).SetBytes(diff10))},
		BlockDiffInfo{1638331271782, *targetToDiff(new(uint256.Int).SetBytes(diff11))},
		BlockDiffInfo{1638331270830, *targetToDiff(new(uint256.Int).SetBytes(diff12))},
		BlockDiffInfo{1638331269597, *targetToDiff(new(uint256.Int).SetBytes(diff13))},
		BlockDiffInfo{1638331267351, *targetToDiff(new(uint256.Int).SetBytes(diff14))},
		BlockDiffInfo{1638331262591, *targetToDiff(new(uint256.Int).SetBytes(diff15))},
		BlockDiffInfo{1638331260306, *targetToDiff(new(uint256.Int).SetBytes(diff16))},
		BlockDiffInfo{1638331254476, *targetToDiff(new(uint256.Int).SetBytes(diff17))},
		BlockDiffInfo{1638331249272, *targetToDiff(new(uint256.Int).SetBytes(diff18))},
		BlockDiffInfo{1638331243260, *targetToDiff(new(uint256.Int).SetBytes(diff19))},
		BlockDiffInfo{1638331237750, *targetToDiff(new(uint256.Int).SetBytes(diff20))},
		BlockDiffInfo{1638331236606, *targetToDiff(new(uint256.Int).SetBytes(diff21))},
		BlockDiffInfo{1638331232507, *targetToDiff(new(uint256.Int).SetBytes(diff22))},
		BlockDiffInfo{1638331212446, *targetToDiff(new(uint256.Int).SetBytes(diff23))},
		BlockDiffInfo{1638331205918, *targetToDiff(new(uint256.Int).SetBytes(diff24))},
	}
	tests := []struct {
		name    string
		args    args
		want    uint256.Int
		wantErr bool
	}{
		{"test difficulty",
			args{
				blocks,
				5000,
			},
			*new(uint256.Int).SetBytes(diff0),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNextTarget(tt.args.blocks, tt.args.timePlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNextTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(targetToDiff(&got).ToBig(), tt.want.ToBig()) {
				t.Errorf("getNextTarget() got = %v, want %v", got, tt.want)
			}
		})
	}
}
