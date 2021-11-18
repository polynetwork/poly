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
	"bytes"
	_ "bytes"
	"encoding/binary"
	"encoding/hex"
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
	"golang.org/x/crypto/sha3"

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

func Hash(data []byte) []byte {
	concatData := bytes.Buffer{}
	concatData.Write([]byte("BlockHeader"))
	prebytes := sha3.Sum256(concatData.Bytes())
	concat := bytes.Buffer{}
	concat.Write(prebytes[:])
	concat.Write(data)
	hashData := sha3.Sum256(concat.Bytes())
	return hashData[:]
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
		hash, _ := header.GetHash()
		fmt.Println(stc.BytesToHexString(*hash))
		abytes, _ := header.BcsSerialize()
		fmt.Println(stc.BytesToHexString(Hash(abytes)))
		var jsonHeader stc.BlockHeader
		json.Unmarshal(header2810118, &jsonHeader)
		assert.Equal(t, header, jsonHeader.ToTypesHeader())
	}
	{
		//header2810119, _ := hex.DecodeString("0x0a097b0a20202020202022626c6f636b5f68617368223a2022307861333832343734643066643132373066376639386632626462643137646561666662313461363964376261386664303630613033326537323366393937623462222c0a20202020202022706172656e745f68617368223a2022307835366533336232353737353933306534396264356230353338323838313835343063633136373934653232653531616437313333646439336363373533343136222c0a2020202020202274696d657374616d70223a202231363337303633303838313635222c0a202020202020226e756d626572223a202232383130313138222c0a20202020202022617574686f72223a202230783436613164303130316634393131343739303265396530303330353130376564222c0a20202020202022617574686f725f617574685f6b6579223a206e756c6c2c0a2020202020202274786e5f616363756d756c61746f725f726f6f74223a2022307832313138386333346634316237643865383039386666643239313761346664373638613064626466623033643130306166303964376263313038643066363037222c0a20202020202022626c6f636b5f616363756d756c61746f725f726f6f74223a2022307834666532633133306430316234393863643666346232303365633239373865663138393036653132656539326463663664613536346437653534613063363330222c0a2020202020202273746174655f726f6f74223a2022307862653564323332376338666632633831363435623734323661663061343032393739616565336163323136383534313230396633383036633534653464363037222c0a202020202020226761735f75736564223a202230222c0a20202020202022646966666963756c7479223a202230783063653737366237222c0a20202020202022626f64795f68617368223a2022307863303165303332396465366438393933343861386566346264353164623536313735623366613039383865353763336463656338656166313361313634643937222c0a20202020202022636861696e5f6964223a20312c0a202020202020226e6f6e6365223a20313234393930323836352c0a202020202020226578747261223a202230783634336230303030220a0920207d0a09")
		//header2810120, _ := hex.DecodeString("0x0a097b0a20202020202022626c6f636b5f68617368223a2022307832346165363865393234373063396439393339316437393538663534306636653966636439633364306432616438653562303336333638613636366634666662222c0a20202020202022706172656e745f68617368223a2022307830306162393030626332383431656666613461353266663036653661613461303930663234383263633830393062633361336666363531396565643135366461222c0a2020202020202274696d657374616d70223a202231363337303633303936393933222c0a202020202020226e756d626572223a202232383130313230222c0a20202020202022617574686f72223a202230783730376438666330313661636165306131613835393736396164306334666366222c0a20202020202022617574686f725f617574685f6b6579223a206e756c6c2c0a2020202020202274786e5f616363756d756c61746f725f726f6f74223a2022307838326134646664623562343066656132626430393266326233393034343739653134623262373165393132646663623736656265643330656663316335353834222c0a20202020202022626c6f636b5f616363756d756c61746f725f726f6f74223a2022307831623433333361303934393137656366323166313234303037333836376235623130363562663266346264666262316236313465383636616539346439326338222c0a2020202020202273746174655f726f6f74223a2022307836373238366336633464663561633762643866356332613033383636616662363465323839666432306136363165306331363633643961313864333762663861222c0a202020202020226761735f75736564223a202230222c0a20202020202022646966666963756c7479223a202230783065396435626338222c0a20202020202022626f64795f68617368223a2022307863303165303332396465366438393933343861386566346264353164623536313735623366613039383865353763336463656338656166313361313634643937222c0a20202020202022636861696e5f6964223a20312c0a202020202020226e6f6e6365223a203138343535303336362c0a202020202020226578747261223a202230783434303033376163220a0920207d0a09")
		//header2810121, _ := hex.DecodeString("0x0a097b0a20202020202022626c6f636b5f68617368223a2022307832303064353630336236386132366135356363333131323438613365343337306335373438373638663532363936366263343633336565613966663238623261222c0a20202020202022706172656e745f68617368223a2022307832346165363865393234373063396439393339316437393538663534306636653966636439633364306432616438653562303336333638613636366634666662222c0a2020202020202274696d657374616d70223a202231363337303633303938393935222c0a202020202020226e756d626572223a202232383130313231222c0a20202020202022617574686f72223a202230783436613164303130316634393131343739303265396530303330353130376564222c0a20202020202022617574686f725f617574685f6b6579223a206e756c6c2c0a2020202020202274786e5f616363756d756c61746f725f726f6f74223a2022307864653436396636316137613961616464646564303032393761346264343130316464343661363534313738363937306630313137376366653836333065633033222c0a20202020202022626c6f636b5f616363756d756c61746f725f726f6f74223a2022307831613935363132323338666139353434333031623262353164663965386462373235366264383566393634353834303533616162333830303431633931643834222c0a2020202020202273746174655f726f6f74223a2022307839333439653131373637323837323664356666386566363665393034366131383036633262393163623136376133353662393935313535663962326136356434222c0a202020202020226761735f75736564223a202230222c0a20202020202022646966666963756c7479223a202230783065346432633561222c0a20202020202022626f64795f68617368223a2022307863303165303332396465366438393933343861386566346264353164623536313735623366613039383865353763336463656338656166313361313634643937222c0a20202020202022636861696e5f6964223a20312c0a202020202020226e6f6e6365223a2036373131323130352c0a202020202020226578747261223a202230783134653130303030220a0920207d0a09")
		//header2810122, _ := hex.DecodeString("0x0a097b0a20202020202022626c6f636b5f68617368223a2022307836633830346634326165383834363061313864326131653435393935363839326631643438303364313565313539323764396330353633386634306231626333222c0a20202020202022706172656e745f68617368223a2022307832303064353630336236386132366135356363333131323438613365343337306335373438373638663532363936366263343633336565613966663238623261222c0a2020202020202274696d657374616d70223a202231363337303633313033323233222c0a202020202020226e756d626572223a202232383130313232222c0a20202020202022617574686f72223a202230783436613164303130316634393131343739303265396530303330353130376564222c0a20202020202022617574686f725f617574685f6b6579223a206e756c6c2c0a2020202020202274786e5f616363756d756c61746f725f726f6f74223a2022307833396239646665636130353237383639313939616230633938303838333635343762386135653333636336323336623634303737333163333833386231616132222c0a20202020202022626c6f636b5f616363756d756c61746f725f726f6f74223a2022307830323161623563663633353732313839626430323836306166633261663035626637326536306135656233383737616633373863366366633436623262353136222c0a2020202020202273746174655f726f6f74223a2022307866633166613435653639306637636466346137366465653939353362646533313531316363666333333936323266616535343836626437663034383735636530222c0a202020202020226761735f75736564223a202230222c0a20202020202022646966666963756c7479223a202230783066323337363038222c0a20202020202022626f64795f68617368223a2022307863303165303332396465366438393933343861386566346264353164623536313735623366613039383865353763336463656338656166313361313634643937222c0a20202020202022636861696e5f6964223a20312c0a202020202020226e6f6e6365223a2036373131323931382c0a202020202020226578747261223a202230783565373330303030220a0920207d0a09")

		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 1
		param.Address = acct.Address
		param.Headers = append(param.Headers, []byte(Header2810119))
		//param.Headers = append(param.Headers, header2810120)
		//param.Headers = append(param.Headers, header2810121)
		//param.Headers = append(param.Headers, header2810122)
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
