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
      "block_hash": "0x346e93455aafe78407a5ed9fa443ebf641e3b219420db21b3d091e57e302a086",
      "parent_hash": "0x263f952597ee89514b4c74c9e122244004d2b680b6dd78ab2f898a63d3272fd0",
      "timestamp": "1638331302562",
      "number": "14644",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x105999bd948ee7c962937ab3a466cd463ca30079501003fc8ab157add6356f89",
      "block_accumulator_root": "0xd642211e6c46fadb8bf85fcf5e61ebaffe9c668b623cdc598c798e81e4603bba",
      "state_root": "0xf5b62f11846518e0c2e05b40ca34fc1c32adedb50ba8d3507efdbdcabc754a6b",
      "gas_used": "0",
      "difficulty": "0x0109",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 707677044,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x263f952597ee89514b4c74c9e122244004d2b680b6dd78ab2f898a63d3272fd0",
      "parent_hash": "0x1e7966d2572f5190165a578aee638f26eebc4f673407f4d6908179c93f9932ba",
      "timestamp": "1638331301987",
      "number": "14643",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0x683fe4df25b4fd6b4a067c09770150210fc44da0de15f82e198c38c4602006b3",
      "block_accumulator_root": "0xd9f07ad0299426b27dd7bae0f829d7d9994a2299fbd5a5ad89826699d6b31265",
      "state_root": "0x0264c05e40402aec5c68a29440e3564cd2ae211078335429d90ef6421348865c",
      "gas_used": "0",
      "difficulty": "0xef",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 4228009554,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x1e7966d2572f5190165a578aee638f26eebc4f673407f4d6908179c93f9932ba",
      "parent_hash": "0xf7f68a0b982e0826efba237cd8ac07461c9f14fef59ccade05ca9c94b537f3bf",
      "timestamp": "1638331301564",
      "number": "14642",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0x87af8e823e976a29a5482d013ac8bf244d17459a60c4c489cebc6b901ed926c7",
      "block_accumulator_root": "0xdfdaccc8d6301a15e16d1d4a51256b66a7feea0e1db941bd06d64c5fa71bffa0",
      "state_root": "0x2ac3cd5943b6ecff0116fa985b911dafa51811307a8dbf8ff99afd794e69f1ff",
      "gas_used": "0",
      "difficulty": "0xec",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 2409105602,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0xf7f68a0b982e0826efba237cd8ac07461c9f14fef59ccade05ca9c94b537f3bf",
      "parent_hash": "0x0d7936ffb0423768b53ee19a640e0109fdbb9ea3c2b5a5ed5126889f8de735df",
      "timestamp": "1638331297135",
      "number": "14641",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x3fbe619908d47a68e5bdb88cbb78045e10256d02dff5ebb614c23261d4936ce7",
      "block_accumulator_root": "0x00cf2a92e7470644ef0d178cd5d0438497165049be67bd011428008f7ed71c76",
      "state_root": "0x19ea3bd23e7ff374f67793c63729b784c24e5b183da7128645d7b2b97f45b1e9",
      "gas_used": "0",
      "difficulty": "0x0103",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1318018096,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x0d7936ffb0423768b53ee19a640e0109fdbb9ea3c2b5a5ed5126889f8de735df",
      "parent_hash": "0x50ed527cb04d4099d3a71dcd50f0f7655590b242edd4b17c7b4c27b219dcb770",
      "timestamp": "1638331288742",
      "number": "14640",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x98cc562c5158c7f134e823367da251b496281b5b1ba37d7d7cff237c4c0638d9",
      "block_accumulator_root": "0x7d5ec47d1b4357db18419ec4380a46e32a94a6ad3381fa94bb7254c7fa12f2f1",
      "state_root": "0x9c58bd37234625ca1c62b7da83b209a05b4d280aa9c460722683bb74c72fe456",
      "gas_used": "0",
      "difficulty": "0xea",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 3580730544,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x50ed527cb04d4099d3a71dcd50f0f7655590b242edd4b17c7b4c27b219dcb770",
      "parent_hash": "0x97f2aef544d74602c7a54a70bd5659272d9fe8ba83ff0235d6091208a41e15be",
      "timestamp": "1638331288188",
      "number": "14639",
      "author": "0xfab981cf1ee57d043be6f4f80b557506",
      "author_auth_key": null,
      "txn_accumulator_root": "0xf6633f47f02d3f376ed4d04030661f90b0d9ab17511db50604da234c76a12856",
      "block_accumulator_root": "0x82eb66a89022d3a40237d69d4beb513ce46a8ceda1117793fb7d031b852270d0",
      "state_root": "0x3cd59576dffb9ab055edb10add064ad334e6c48a194830bb0e383253b99d3842",
      "gas_used": "0",
      "difficulty": "0xd4",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1826850589,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x97f2aef544d74602c7a54a70bd5659272d9fe8ba83ff0235d6091208a41e15be",
      "parent_hash": "0x52951e1470eed7a7998eb7eba79a955a4b99dd3b87912e06cbaae20402df1cf6",
      "timestamp": "1638331287706",
      "number": "14638",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x5b4da3653804a8faefa943c712438029b0b313d0622346de1d8d2fecbfd6ecfe",
      "block_accumulator_root": "0xfb13521ddaac3e41ccf08125b587c88a63d2e1f19c5dd076d569f05b9ab3f8ae",
      "state_root": "0xb1f8a1c11f860ea3c2d3c9725ffa674b9c674a4aeedebbb3b8d3c1e5f24d6372",
      "gas_used": "0",
      "difficulty": "0xcf",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1370287593,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x52951e1470eed7a7998eb7eba79a955a4b99dd3b87912e06cbaae20402df1cf6",
      "parent_hash": "0x29b712d407fce15e38dd0e7dc09dceffe1c0afed64ee62912a96ba8f415eb318",
      "timestamp": "1638331283650",
      "number": "14637",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0xa57ef660b3fcafc502bcc4f459ff1b134105d0eff386ed81c0f68cb6fd20c696",
      "block_accumulator_root": "0xcffb52fe32cd55ee763a95631c435439b5f4e20786e235d12c36e93d547a678f",
      "state_root": "0xcf1ca859c53554669d23d4079dd7fd00b91cc2cb5d36f55dc6fa98cdb69db0a5",
      "gas_used": "0",
      "difficulty": "0xc5",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 82749575,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x29b712d407fce15e38dd0e7dc09dceffe1c0afed64ee62912a96ba8f415eb318",
      "parent_hash": "0x9d7bcb8caa37bc743f8196f7ede2c8d9d31c51054217cc51a5132e287cd29fd4",
      "timestamp": "1638331281477",
      "number": "14636",
      "author": "0xfab981cf1ee57d043be6f4f80b557506",
      "author_auth_key": null,
      "txn_accumulator_root": "0x954f9dc32512aa00eab36f23e731d78062df19f58b98525fb1dcb651fd2476d3",
      "block_accumulator_root": "0x0a15e75d55bb485b795c92c10b214c3d55e8a74ca7e84f7a7b1a0ff999c293ba",
      "state_root": "0xd639c76cb3061410a0b859b6cd10967dd607162bc659328e76014fdec58beb7d",
      "gas_used": "0",
      "difficulty": "0xc4",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1917343304,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x9d7bcb8caa37bc743f8196f7ede2c8d9d31c51054217cc51a5132e287cd29fd4",
      "parent_hash": "0x57f3332b908cb47092c50a7945d91eef0fb0a48920ef8cce6d1fd6ddc60b78ea",
      "timestamp": "1638331276488",
      "number": "14635",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0x92dae868a7781d9b529732db62b89dc1190c6adc33a6639c759fa20bff1598ac",
      "block_accumulator_root": "0x224ab95ff059ea96e63a60be519341b564234e40a6d3377333135bd9d4c945fa",
      "state_root": "0xf4e7100af3bad35c286bce970547a8c2da699cd402183ec96a0e56c77c875595",
      "gas_used": "0",
      "difficulty": "0xbc",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 2977870676,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x57f3332b908cb47092c50a7945d91eef0fb0a48920ef8cce6d1fd6ddc60b78ea",
      "parent_hash": "0x7946118b179f473cde99d34c3868574c2d025319d97dff30a8cd5a4b02dbcfb2",
      "timestamp": "1638331273581",
      "number": "14634",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x6466149c92495b533239220c512fe09ad95165ed606de0436fcba66d9da1d65b",
      "block_accumulator_root": "0x1f2605479c1ef277496f833104a2241994c07c703ee036b8a2369251cb24c51f",
      "state_root": "0x87dcf93bc89d770113b27d3df1178f59f14445e1a8533e7813b28ee82bd306c0",
      "gas_used": "0",
      "difficulty": "0xb1",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 2256672915,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x7946118b179f473cde99d34c3868574c2d025319d97dff30a8cd5a4b02dbcfb2",
      "parent_hash": "0x06ad4a0478056ee56f60dfadc4edeb51e760024492ad5c6790ebaa99c70f579d",
      "timestamp": "1638331271782",
      "number": "14633",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x5021e047ac221628dfab25dab0cec90a8bc75df749918edf39f50ace4cb5eac8",
      "block_accumulator_root": "0x9e7e765c475f3a5f9bb58fd0f7584fbafe0d6616941a87cb1ef665c3084b7c82",
      "state_root": "0x55d018ecbeb45b2ecbc69f41e00b923e918abae52a81709aa6b68ede7cbfec16",
      "gas_used": "0",
      "difficulty": "0xa4",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1089744842,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x06ad4a0478056ee56f60dfadc4edeb51e760024492ad5c6790ebaa99c70f579d",
      "parent_hash": "0xf91417d1f9fad83a0eee21a246bbaee25f151764c1ae230eb150357ebff3ff34",
      "timestamp": "1638331270830",
      "number": "14632",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0x86321de24f446474d9eca884631cd14ef6a55c3ef78624bf4b3346e76332a8cd",
      "block_accumulator_root": "0xb4891a7d1237c6d6caa0f41a092b3fd30bfdfc40d1150dec6d37acdb720d00da",
      "state_root": "0x37205300460f4dce0a5e53b09a531dc6135439ae7e3e43de6275d1a8428201a3",
      "gas_used": "0",
      "difficulty": "0x9a",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 2403003260,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0xf91417d1f9fad83a0eee21a246bbaee25f151764c1ae230eb150357ebff3ff34",
      "parent_hash": "0x5b50c998ae865e64e622790dfea81595be1eb025ea0ea2d9647541df89019ac7",
      "timestamp": "1638331269597",
      "number": "14631",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0xc3c2b8c2da5bed9b1a7c54db16d744aca7682fe403b0c9c38a06546b3a1a89b7",
      "block_accumulator_root": "0x74052dff3e2541472cd151c5681c4b24820fa31a91e45c3e408815b9d3f7eeef",
      "state_root": "0x5ef41259cd81fb6bfc3f2fd7e466968b1052aa63f33387748844145cee3fb8e4",
      "gas_used": "0",
      "difficulty": "0x93",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 855207348,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x5b50c998ae865e64e622790dfea81595be1eb025ea0ea2d9647541df89019ac7",
      "parent_hash": "0x9e34dfb0a485677828f4fe619d1fee04922c3fcf766e761a231ae5c32da62923",
      "timestamp": "1638331267351",
      "number": "14630",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x9df2ff42c0476de2658f9febf6ec559abcd22bff12f7ddb630a9be950c665da5",
      "block_accumulator_root": "0xad20e4517841472c009d3c3fc0f67867fe92f4abdf52ec55afd06001bd5b0cdb",
      "state_root": "0xc439d77733133f835ed64881bf261033ca318909f6b38968f9cce960dcc2aedb",
      "gas_used": "0",
      "difficulty": "0x93",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1355781271,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x9e34dfb0a485677828f4fe619d1fee04922c3fcf766e761a231ae5c32da62923",
      "parent_hash": "0x9ccf9ffb18cef52fba86f7e65e291923e430272d70d217695ca44a11af6ca89c",
      "timestamp": "1638331262591",
      "number": "14629",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0x1e7503b6b3989efc3dbfc076c923df0121abc26fc51039ea66726b690a9d1cbf",
      "block_accumulator_root": "0x0fa34decf029f561e1d870a55ac24480a2089e218b6fb0c137f37a89acc41ab3",
      "state_root": "0x60179a2bf6d599c01a5c6763eae635cff523132805d1d436bd22591f90ba7a35",
      "gas_used": "0",
      "difficulty": "0x8d",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 114265595,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x9ccf9ffb18cef52fba86f7e65e291923e430272d70d217695ca44a11af6ca89c",
      "parent_hash": "0x305c417ed7bfa76d2f19816996514dc8de7e7b8daed9699774ee8d0ea3b8b75d",
      "timestamp": "1638331260306",
      "number": "14628",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x226bda25787f396a4c9f36f55a65ef41bc844b9aabf3a625fcf9ea328f4471d8",
      "block_accumulator_root": "0x5b8fe190a787210f6030327c392401a574b399dd3717b75b7875e13a90b633cd",
      "state_root": "0x0d49f5c3688b47e9f72ded60c6019525acc3c9e68775dc697acf70312f41a77a",
      "gas_used": "0",
      "difficulty": "0x8f",
      "body_hash": "0xcc91b44978512fd25b67474c56e94dac0d23969527f4a26e499a6349213543d7",
      "chain_id": 253,
      "nonce": 2342164875,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x305c417ed7bfa76d2f19816996514dc8de7e7b8daed9699774ee8d0ea3b8b75d",
      "parent_hash": "0x68faa019b91ba0462b2a60eafca5ba1e71df47724b2b34e87c3fa350596bc00e",
      "timestamp": "1638331254476",
      "number": "14627",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0xa6c69e42f9d09e9bb2e527cef7ebea0aafdc5e37d6fcd2218201bcb564bbf582",
      "block_accumulator_root": "0xd1ef25f0fe7cb8f5f807c3902a2ad4a1dc266c14603e9e4b9c51b9926f2436cd",
      "state_root": "0x16ea6ad9dd014056338a2f373f81a1fb30c02df225dfc4ce1cc09b82a41edcb9",
      "gas_used": "0",
      "difficulty": "0x90",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 3056293911,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x68faa019b91ba0462b2a60eafca5ba1e71df47724b2b34e87c3fa350596bc00e",
      "parent_hash": "0x62f1b7b1db79acf5ed63eba318717809003a81267c111879a5921547eb829ef4",
      "timestamp": "1638331249272",
      "number": "14626",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x9d1f2ee31065ab017641fdf4131b6d5dbeb525b84495893f76e506d19cdc0063",
      "block_accumulator_root": "0xfbacafbe845a9b9734854b3e4842c04ceaa577c0f97753822214202bc0743b2f",
      "state_root": "0x8165199fe0b63910f168b59a0a8280b8f45ed1f82c43b21330caf833442f2f20",
      "gas_used": "0",
      "difficulty": "0x92",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 2053435971,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x62f1b7b1db79acf5ed63eba318717809003a81267c111879a5921547eb829ef4",
      "parent_hash": "0x2f14e04840b420f7f3b55cc04a815db5507e8d91506b88b697bafd7b8fcd0d69",
      "timestamp": "1638331243260",
      "number": "14625",
      "author": "0xfab981cf1ee57d043be6f4f80b557506",
      "author_auth_key": null,
      "txn_accumulator_root": "0xb808570e51bea03cbcce7df2cfaf3edc310de44b13cb51e1134f2538ef6197c9",
      "block_accumulator_root": "0xc6c539ae9e7be038d6cbf0ab76429c8b0adcf100146397fd3af68203ab433e24",
      "state_root": "0x478f60a5d6ef7fbb20222c3f4f73995395b8175dcec660e6c0320e668553e0cb",
      "gas_used": "0",
      "difficulty": "0x93",
      "body_hash": "0x4c126960fdb250482965184463dff06e8c39f5f1187d8aa5e6d21999bcc90524",
      "chain_id": 253,
      "nonce": 3516353973,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x2f14e04840b420f7f3b55cc04a815db5507e8d91506b88b697bafd7b8fcd0d69",
      "parent_hash": "0x4dadbec47517f8281449d780de578b57a159fe22340d2d854e5f719805320556",
      "timestamp": "1638331237750",
      "number": "14624",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x10b6c85a74c421ff222775294244028df92551c5bb8e1afe11647ffed17ad5e9",
      "block_accumulator_root": "0xa97bd9bc6687952dab1339bd803186f01fdf25a2a1a1bbcf611190614a85c369",
      "state_root": "0x2dc39dca3900bf668a43af862b681618c858bb66c2452efbcad43d427ead4033",
      "gas_used": "0",
      "difficulty": "0x8b",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 326721171,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x4dadbec47517f8281449d780de578b57a159fe22340d2d854e5f719805320556",
      "parent_hash": "0xe71ea2f82a8ff9a301829c1c420dd315d1567115ebae824dd35d4aca33d8127c",
      "timestamp": "1638331236606",
      "number": "14623",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0xf4c7c2a9e2f2e6a7557e3db620a5616d911a93dbb9c97c592186f7a1133f9304",
      "block_accumulator_root": "0x7771677be900d5e44c74f04fce9aff986a09397eb518dc80c48e763801c25de0",
      "state_root": "0x3020ae8dd0ee8ccc7ab54fd79ede07112b1c242402c850e5253c1a6ed9f55f67",
      "gas_used": "0",
      "difficulty": "0x8a",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 210950409,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0xe71ea2f82a8ff9a301829c1c420dd315d1567115ebae824dd35d4aca33d8127c",
      "parent_hash": "0x35600f8a40bcedc18b8da6ec8bb208412ded912c4b6e74020a7a533302326fb0",
      "timestamp": "1638331232507",
      "number": "14622",
      "author": "0xed0f7fcbc522176bf6c8c42f60419718",
      "author_auth_key": null,
      "txn_accumulator_root": "0x109754f79e05f19c9bb68370d76b4f664a93862f3b35006a2924b4a7240c2bb3",
      "block_accumulator_root": "0x37c3c505c4f58ad0001562d91357ee41af92e94806b09a3a5bf78c2d9c130265",
      "state_root": "0x5065c9569db1e8fa16142651a1c29a7ea63cda73d95a7e29804829ac5fa5738f",
      "gas_used": "0",
      "difficulty": "0xaf",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 2490603851,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0x35600f8a40bcedc18b8da6ec8bb208412ded912c4b6e74020a7a533302326fb0",
      "parent_hash": "0xccc9b2e1ef4aac0d4de062ce68c62bdfbae7f7a9ddc5113bb6d25bc9fa78088c",
      "timestamp": "1638331212446",
      "number": "14621",
      "author": "0x00e4ea282432073992bc04ab278ddd60",
      "author_auth_key": null,
      "txn_accumulator_root": "0xb3c6790007b32d7ed0b6ed10b50d0adb2932da9e5ccec11de333e9cdfce019da",
      "block_accumulator_root": "0xee69446adc8931e2151a3bcd78b0a0df233c2e2f94502820a7496872b4d774dc",
      "state_root": "0xd87854af6b1239ec346cdf7c1ef45651719fd3f45d076116f8fe9c59543179af",
      "gas_used": "0",
      "difficulty": "0xb4",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 608097306,
      "extra": "0x00000000"
    },
    {
      "block_hash": "0xccc9b2e1ef4aac0d4de062ce68c62bdfbae7f7a9ddc5113bb6d25bc9fa78088c",
      "parent_hash": "0x95e3671fd68c4a22d1a16bdff62188306b4a84a5368c06a2a4c8fec314940ab6",
      "timestamp": "1638331205918",
      "number": "14620",
      "author": "0xfab981cf1ee57d043be6f4f80b557506",
      "author_auth_key": null,
      "txn_accumulator_root": "0x12e8f742a9afcfa6d8a4c1d5ea3ee1ce33029602781a03db2789f06a98732174",
      "block_accumulator_root": "0x7f90aed098e4f843f10a04abd5c58a720434aa4fd5fd014219dfe8a1cb52adfa",
      "state_root": "0x753bcd8091c943f1a87c1b3b6430d6bb0f8ec93aef6467b404447b8004342a42",
      "gas_used": "0",
      "difficulty": "0xa9",
      "body_hash": "0xc01e0329de6d899348a8ef4bd51db56175b3fa0988e57c3dcec8eaf13a164d97",
      "chain_id": 253,
      "nonce": 1607925320,
      "extra": "0x00000000"
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
		assert.Equal(t, uint64(14620), height)
		headerHash := getHeaderHashByHeight(native, 14620)
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
