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

package neo3

import (
	"github.com/joeqian10/neo3-gogogo/block"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/helper"
	tx2 "github.com/joeqian10/neo3-gogogo/tx"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/genesis"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	//neoAcct = func() *wallet.Account {
	//	acct, _ := wallet.NewAccount()
	//	return acct
	//}()
	acct          *account.Account = account.NewAccount("")
	getNativeFunc                  = func() *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service, _ := native.NewNativeService(cacheDB, new(types.Transaction), 0, 200, common.Uint256{}, 0, nil, false)
		return service
	}
	getNeoHanderFunc = func() *Neo3Handler {
		return NewNeo3Handler()
	}
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
)

func init() {
	setBKers()
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns, _ := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	return ns
}

func TestSyncGenesisHeader(t *testing.T) {
	prevHash, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	merkleRoot, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	nextConsensus, _ := crypto.AddressToScriptHash("NVg7LjGcUSrgxgjX3zEgqaksfMaiS8Z6e1", helper.DefaultAddressVersion)
	vs, _ := crypto.Base64Decode("EQ==")
	witness := tx2.Witness{
		InvocationScript: []byte{},
		VerificationScript: vs,
	}
	genesisHeader := &NeoBlockHeader{Header: block.NewBlockHeader()}
	genesisHeader.SetVersion(0)
	genesisHeader.SetPrevHash(prevHash)
	genesisHeader.SetMerkleRoot(merkleRoot)
	genesisHeader.SetTimeStamp(1468595301000)
	genesisHeader.SetIndex(0)
	genesisHeader.SetPrimaryIndex(0x00)
	genesisHeader.SetNextConsensus(nextConsensus)
	genesisHeader.SetWitnesses([]tx2.Witness{witness})

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = 4
	sink := common.NewZeroCopySink(nil)
	err := genesisHeader.Serialization(sink)
	param.GenesisHeader = sink.Bytes()
	assert.Nil(t, err)

	sink = common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	//cdb := new(storage.CacheDB)
	//cdb.Put()
	native := NewNative(sink.Bytes(), tx, nil)
	neoHandler := NewNeo3Handler()
	err = neoHandler.SyncGenesisHeader(native)
	assert.NoError(t, err)
}

func TestSyncBlockHeader(t *testing.T) {
	prevHash, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	merkleRoot, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	nextConsensus, _ := crypto.AddressToScriptHash("NVg7LjGcUSrgxgjX3zEgqaksfMaiS8Z6e1", helper.DefaultAddressVersion)
	is := []byte{}
	vs, _ := crypto.Base64Decode("EQ==")
	witness := tx2.Witness{
		InvocationScript: is,
		VerificationScript: vs,
	}
	genesisHeader := &NeoBlockHeader{Header: block.NewBlockHeader()}
	genesisHeader.SetVersion(0)
	genesisHeader.SetPrevHash(prevHash)
	genesisHeader.SetMerkleRoot(merkleRoot)
	genesisHeader.SetTimeStamp(1468595301000)
	genesisHeader.SetIndex(0)
	genesisHeader.SetPrimaryIndex(0x00)
	genesisHeader.SetNextConsensus(nextConsensus)
	genesisHeader.SetWitnesses([]tx2.Witness{witness})
	sink := common.NewZeroCopySink(nil)
	if err := genesisHeader.Serialization(sink); err != nil {
		t.Errorf("NeoBlockHeaderToBytes error:%v", err)
	}
	neoGenesisHeaderbs := sink.Bytes()
	// block 100
	prevHash, _ = helper.UInt256FromString("0xcee650e843a8f8cf78f7c62a2bc2108375df93bfa9a912de64bea2d8948fec31")
	merkleRoot, _ = helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	nextConsensus, _ = crypto.AddressToScriptHash("NYmMriQPYAiNxHC7tziV4ABJku5Yqe79N4", helper.DefaultAddressVersion)
	is, _ = crypto.Base64Decode("DEAF2siXzBq5rCzpaNvAxTPuebSgsn3XX7bKMvuf1RzQx1QqJDLBn/XxMyCzAOnsolp67X8eLZ8xuc4bovqSMf4lDECxfs8iV5/5yLs2hVW0tB1d1n7R1J3HEoJ8vNCJr/xrqt3bJRoOFL+eObBmjyo+ZSk5M6GrGH+UqpglZqr0upu8DEAX9lh0LKdbFTOg5GfZ8gP9UdGURu/xbM288BKFUBXhTH1p/2Y4hqZzJoXes+DdlRCwWzhToCMa468OnmTHkxDoDEB4etR6RX+B69qB5cv7QjihqnTowYWzU3Zhec+yz+2wgETkD0aD4uUafiSGpCK7xNB7aknDbFgMJXWSK7cM+c3NDEDw4z4PxskKUfJ1cmXKXxhtdzo/05iEi6c/n2rfZHPLd/YA0aBPQuWf3QSWizQYsmsYyA2uriKR2PA7asqYYB60")
	vs, _ = crypto.Base64Decode("FQwhAwCbdUDhDyVi5f2PrJ6uwlFmpYsm5BI0j/WoaSe/rCKiDCEDAgXpzvrqWh38WAryDI1aokaLsBSPGl5GBfxiLIDmBLoMIQIUuvDO6jpm8X5+HoOeol/YvtbNgua7bmglAYkGX0T/AQwhAj6bMuqJuU0GbmSbEk/VDjlu6RNp6OKmrhsRwXDQIiVtDCEDQI3NQWOW9keDrFh+oeFZPFfZ/qiAyKahkg6SollHeAYMIQKng0vpsy4pgdFXy1u9OstCz9EepcOxAiTXpE6YxZEPGwwhAroscPWZbzV6QxmHBYWfriz+oT4RcpYoAHcrPViKnUq9F0F7zmyl")
	witness = tx2.Witness{
		InvocationScript: is,
		VerificationScript: vs,
	}
	neoHeader100 := &NeoBlockHeader{Header: block.NewBlockHeader()}
	neoHeader100.SetVersion(0)
	neoHeader100.SetPrevHash(prevHash)
	neoHeader100.SetMerkleRoot(merkleRoot)
	neoHeader100.SetTimeStamp(1616577294488)
	neoHeader100.SetIndex(100)
	neoHeader100.SetPrimaryIndex(0x00)
	neoHeader100.SetNextConsensus(nextConsensus)
	neoHeader100.SetWitnesses([]tx2.Witness{witness})
	sink = common.NewZeroCopySink(nil)
	if err := neoHeader100.Serialization(sink); err != nil {
		t.Errorf("NeoBlockHeaderToBytes error:%v", err)
	}
	neoHeader100bs := sink.Bytes()

	// block 200
	prevHash, _ = helper.UInt256FromString("0xe22cd5bb35d832e59554dc2d4165a48bed56b2d7f9df379639078cca35ecc770")
	merkleRoot, _ = helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	nextConsensus, _ = crypto.AddressToScriptHash("NYmMriQPYAiNxHC7tziV4ABJku5Yqe79N4", helper.DefaultAddressVersion)
	is, _ = crypto.Base64Decode("DEBGAk7oudrv67j+GFWetgK8W0Hu6m/15ceMxaq3sl6aKfjTQckmQfqzzGgzJ/rrdWMCyCsjYkgc14MdFvI/GnopDEBj9dRQtcKyu9K2qOyvqqkUoqK/9A32kdj0LbXpYPw+WDrz0DlvjNQF8dc2EuvmwTrYuQ5fUTrHvaPXKTmOPl3IDEAKtZf8AaZmY+onQejV8jqkEN5DGKKWthYGpVza5jueTvx4Hi1B5Uh7k6jW5Z6Y7mUVGuIUAGLbeULsp/MAQiyjDEBibp/Gy0rg5h1Hm3TokJi12KfYSMizn973+rkMDjKiWW1ySq6Sif3BEqlHi1prbuFPYSQf7xiJgs3+P0aFXfPzDEBdqKNocAkjWPBbxrHCHq0DRJoXBgmrXA9BSmytymp/pdU4xt2Y/Gxb/GBBXOAorOumjZ46DxWwnqfVpJ9adHMj")
	vs, _ = crypto.Base64Decode("FQwhAwCbdUDhDyVi5f2PrJ6uwlFmpYsm5BI0j/WoaSe/rCKiDCEDAgXpzvrqWh38WAryDI1aokaLsBSPGl5GBfxiLIDmBLoMIQIUuvDO6jpm8X5+HoOeol/YvtbNgua7bmglAYkGX0T/AQwhAj6bMuqJuU0GbmSbEk/VDjlu6RNp6OKmrhsRwXDQIiVtDCEDQI3NQWOW9keDrFh+oeFZPFfZ/qiAyKahkg6SollHeAYMIQKng0vpsy4pgdFXy1u9OstCz9EepcOxAiTXpE6YxZEPGwwhAroscPWZbzV6QxmHBYWfriz+oT4RcpYoAHcrPViKnUq9F0F7zmyl")
	witness = tx2.Witness{
		InvocationScript: is,
		VerificationScript: vs,
	}
	neoHeader200 := &NeoBlockHeader{Header: block.NewBlockHeader()}
	neoHeader200.SetVersion(0)
	neoHeader200.SetPrevHash(prevHash)
	neoHeader200.SetMerkleRoot(merkleRoot)
	neoHeader200.SetTimeStamp(1616578903762)
	neoHeader200.SetIndex(200)
	neoHeader200.SetPrimaryIndex(0x00)
	neoHeader200.SetNextConsensus(nextConsensus)
	neoHeader200.SetWitnesses([]tx2.Witness{witness})
	sink = common.NewZeroCopySink(nil)
	if err := neoHeader200.Serialization(sink); err != nil {
		t.Errorf("NeoBlockHeaderToBytes error:%v", err)
	}
	neoHeader200bs := sink.Bytes()

	neoHandler := NewNeo3Handler()
	var native *native.NativeService
	{
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 4
		param.GenesisHeader = neoGenesisHeaderbs
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, nil)
		err := neoHandler.SyncGenesisHeader(native)
		assert.NoError(t, err)
	}

	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 4
		param.Address = acct.Address
		param.Headers = append(param.Headers, neoHeader100bs)
		param.Headers = append(param.Headers, neoHeader200bs)
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		err := neoHandler.SyncBlockHeader(native)
		assert.NoError(t, err)
	}
}
