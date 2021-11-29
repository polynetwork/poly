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
	"reflect"

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
		param.Headers = append(param.Headers, []byte(Header2810119))
		param.Headers = append(param.Headers, []byte(Header2810120))
		param.Headers = append(param.Headers, []byte(Header2810121))
		param.Headers = append(param.Headers, []byte(Header2810122))
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

func Test_getNextTarget(t *testing.T) {
	type args struct {
		blocks   []BlockDiffInfo
		timePlan uint64
	}
	diff0, _ := hex.DecodeString("0e74b15e")
	diff1, _ := hex.DecodeString("128dd7af")
	diff2, _ := hex.DecodeString("14a42733")
	diff3, _ := hex.DecodeString("133b86e6")
	diff4, _ := hex.DecodeString("11f54845")
	diff5, _ := hex.DecodeString("1310dccb")
	diff6, _ := hex.DecodeString("1222c9f0")
	diff7, _ := hex.DecodeString("12b92b58")
	diff8, _ := hex.DecodeString("124da70c")
	diff9, _ := hex.DecodeString("11408ac0")
	diff10, _ := hex.DecodeString("15d7bcb2")
	diff11, _ := hex.DecodeString("16317044")
	diff12, _ := hex.DecodeString("168d12d7")
	diff13, _ := hex.DecodeString("1ac09aff")
	diff14, _ := hex.DecodeString("18a2cdbd")
	diff15, _ := hex.DecodeString("184f681a")
	diff16, _ := hex.DecodeString("18115b5e")
	diff17, _ := hex.DecodeString("1afaf67a")
	diff18, _ := hex.DecodeString("19f7fe19")
	diff19, _ := hex.DecodeString("180d92a4")
	diff20, _ := hex.DecodeString("161206f7")
	diff21, _ := hex.DecodeString("15688141")
	diff22, _ := hex.DecodeString("13ec5a26")
	diff23, _ := hex.DecodeString("1431ab06")
	diff24, _ := hex.DecodeString("135b2bde")
	blocks := []BlockDiffInfo{
		BlockDiffInfo{1637911931937, *new(uint256.Int).SetBytes(diff1)},
		BlockDiffInfo{1637911903653, *new(uint256.Int).SetBytes(diff2)},
		BlockDiffInfo{1637911889623, *new(uint256.Int).SetBytes(diff3)},
		BlockDiffInfo{1637911888900, *new(uint256.Int).SetBytes(diff4)},
		BlockDiffInfo{1637911888227, *new(uint256.Int).SetBytes(diff5)},
		BlockDiffInfo{1637911877864, *new(uint256.Int).SetBytes(diff6)},
		BlockDiffInfo{1637911875570, *new(uint256.Int).SetBytes(diff7)},
		BlockDiffInfo{1637911867620, *new(uint256.Int).SetBytes(diff8)},
		BlockDiffInfo{1637911863497, *new(uint256.Int).SetBytes(diff9)},
		BlockDiffInfo{1637911862363, *new(uint256.Int).SetBytes(diff10)},
		BlockDiffInfo{1637911839267, *new(uint256.Int).SetBytes(diff11)},
		BlockDiffInfo{1637911832464, *new(uint256.Int).SetBytes(diff12)},
		BlockDiffInfo{1637911825582, *new(uint256.Int).SetBytes(diff13)},
		BlockDiffInfo{1637911811002, *new(uint256.Int).SetBytes(diff14)},
		BlockDiffInfo{1637911809762, *new(uint256.Int).SetBytes(diff15)},
		BlockDiffInfo{1637911804871, *new(uint256.Int).SetBytes(diff16)},
		BlockDiffInfo{1637911800093, *new(uint256.Int).SetBytes(diff17)},
		BlockDiffInfo{1637911789309, *new(uint256.Int).SetBytes(diff18)},
		BlockDiffInfo{1637911785915, *new(uint256.Int).SetBytes(diff19)},
		BlockDiffInfo{1637911784248, *new(uint256.Int).SetBytes(diff20)},
		BlockDiffInfo{1637911782976, *new(uint256.Int).SetBytes(diff21)},
		BlockDiffInfo{1637911779023, *new(uint256.Int).SetBytes(diff22)},
		BlockDiffInfo{1637911777159, *new(uint256.Int).SetBytes(diff23)},
		BlockDiffInfo{1637911770816, *new(uint256.Int).SetBytes(diff24)},
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
				5962,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNextTarget() got = %v, want %v", got, tt.want)
			}
		})
	}
}
