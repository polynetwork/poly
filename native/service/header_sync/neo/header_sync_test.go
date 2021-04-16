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

package neo

import (
	"encoding/hex"
	"testing"

	"encoding/binary"
	"github.com/joeqian10/neo-gogogo/block"
	"github.com/joeqian10/neo-gogogo/helper"
	tx2 "github.com/joeqian10/neo-gogogo/tx"
	"github.com/joeqian10/neo-gogogo/wallet"
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
)

var (
	neoAcct = func() *wallet.Account {
		acct, _ := wallet.NewAccount()
		return acct
	}()
	acct          *account.Account = account.NewAccount("")
	getNativeFunc                  = func() *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service, _ := native.NewNativeService(cacheDB, new(types.Transaction), 0, 200, common.Uint256{}, 0, nil, false)
		return service
	}
	getNeoHanderFunc = func() *NEOHandler {
		return NewNEOHandler()
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
	merKleRoot, _ := helper.UInt256FromString("0x803ff4abe3ea6533bcc0be574efa02f83ae8fdc651c879056b0d9be336c01bf4")
	nextConsensus, _ := helper.AddressToScriptHash("APyEx5f4Zm4oCHwFWiSTaph1fPBxZacYVR")
	consensusData := binary.BigEndian.Uint64(helper.HexToBytes("000000007c2bac1d"))
	genesisHeader := &NeoBlockHeader{
		&block.BlockHeader{
			Version:       0,
			PrevHash:      prevHash,
			MerkleRoot:    merKleRoot,
			Timestamp:     1468595301,
			Index:         0,
			NextConsensus: nextConsensus,
			ConsensusData: consensusData,
			Witness: &tx2.Witness{
				InvocationScript:   []byte{0},
				VerificationScript: []byte{81},
			},
		},
	}
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

	native := NewNative(sink.Bytes(), tx, nil)
	neoHandler := NewNEOHandler()
	err = neoHandler.SyncGenesisHeader(native)
	assert.NoError(t, err)
}

func TestSyncBlockHeader(t *testing.T) {

	prevHash, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	merKleRoot, _ := helper.UInt256FromString("0x803ff4abe3ea6533bcc0be574efa02f83ae8fdc651c879056b0d9be336c01bf4")
	nextConsensus, _ := helper.AddressToScriptHash("APyEx5f4Zm4oCHwFWiSTaph1fPBxZacYVR")
	consensusData := binary.BigEndian.Uint64(helper.HexToBytes("000000007c2bac1d"))
	genesisHeader := &NeoBlockHeader{
		&block.BlockHeader{
			Version:       0,
			PrevHash:      prevHash,
			MerkleRoot:    merKleRoot,
			Timestamp:     1468595301,
			Index:         0,
			NextConsensus: nextConsensus,
			ConsensusData: consensusData,
			Witness: &tx2.Witness{
				InvocationScript:   []byte{0},
				VerificationScript: []byte{81},
			},
		},
	}
	sink := common.NewZeroCopySink(nil)
	if err := genesisHeader.Serialization(sink); err != nil {
		t.Errorf("NeoBlockHeaderToBytes error:%v", err)
	}
	neoGenesisHeaderbs := sink.Bytes()

	prevHash, _ = helper.UInt256FromString("0x07e53bcbe93de88784e01401f8dae061934d34747e76b6a18b63e4213f3de220")
	merKleRoot, _ = helper.UInt256FromString("0xbf6a0b78bede3ad758e09479b8712f60cb225a4047265847169b3864e23a3bc8")
	nextConsensus, _ = helper.AddressToScriptHash("APyEx5f4Zm4oCHwFWiSTaph1fPBxZacYVR")
	consensusData = binary.BigEndian.Uint64(helper.HexToBytes("d9d3b5ccf2eedde1"))
	is100, _ := hex.DecodeString("40424db765bc1e92e530292ec04ff8ddffb79bec13f04fd9f85c00163328aa9d64f0b40b74ca8c4b56445c9048c50e6a67df57ab221593612c6165251d9770f7e140465f1d1d3b532fcaa8a98633316e24a07358c857a3565f7cc9a1b87dd3e6dcbb191a7c78c1b57889924e813a0daacea5281884ce814d10469560f43c9d567cf440fd7252d9607389e9b61c577a8705b1d74165979dd9440c4a71d47443fc1014e46957b0a537e1244fd9b4363aefb2df5971749daf9073cfd014aecb7dba2b13ab40c141f6c63267ad12ebadb154a83a3444eccff046de534cda6f29059e531de58bfce6287ca68a62b45766df5522dfed449b3d1bdc0a319ab07d21cf8839f5b59240fee381887b2dc82447fbe9e6db6c1aa9adff8f7a7d2998cea4f901c002098115d7ba7e6218275c8690f86b92e8b641d59152243f2253ff86fa9c2b6413a52256")
	vs100, _ := hex.DecodeString("552102486fd15702c4490a26703112a5cc1d0923fd697a33406bd5a1c00e0013b09a7021024c7b7fb6c310fccf1ba33b082519d82964ea93868d676662d4a59ad548df0e7d2102aaec38470f6aad0042c6e877cfd8087d2676b0f516fddd362801b9bd3936399e2103b209fd4f53a7170ea4444e0cb0a6bb6a53c2bd016926989cf85f9b0fba17a70c2103b8d9d5771d8f513aa0869b9cc8d50986403b78c6da36890638c3d46a5adce04a2102ca0e27697b9c248f6f16e085fd0061e26f44da85b58ee835c110caa5ec3ba5542102df48f60e8f3e01c48ff40b9b7f1310d7a8b2a193188befe1c2e3df740e89509357ae")
	neoHeader100 := &NeoBlockHeader{
		&block.BlockHeader{
			Version:       0,
			PrevHash:      prevHash,
			MerkleRoot:    merKleRoot,
			Timestamp:     1476649243,
			Index:         100,
			NextConsensus: nextConsensus,
			ConsensusData: consensusData,
			Witness: &tx2.Witness{
				InvocationScript:   is100,
				VerificationScript: vs100,
			},
		},
	}
	sink = common.NewZeroCopySink(nil)
	if err := neoHeader100.Serialization(sink); err != nil {
		t.Errorf("NeoBlockHeaderToBytes error:%v", err)
	}
	neoHeader100bs := sink.Bytes()

	prevHash, _ = helper.UInt256FromString("0xb4cabbcde5e5d5ecf0429cb4726f7a4d857e195e12bdc568cb1df2097c2d918d")
	merKleRoot, _ = helper.UInt256FromString("0xb230e8bc2e0f35eff5279de5d6f9a6b1c26c5757247c4f33d744676919bfb3d1")
	nextConsensus, _ = helper.AddressToScriptHash("APyEx5f4Zm4oCHwFWiSTaph1fPBxZacYVR")
	consensusData = binary.BigEndian.Uint64(helper.HexToBytes("ab01a2420960665a"))
	is200, _ := hex.DecodeString("409de7102817ffda29e13e96e3ff541ea1c7498805504d39d2145990c9862f2aacdb748b631233c14a63cc1483f730843eced2abab80a9c5bbe4c7b5a552569cac40d3fde603e228e2ca3ee25ebf7651692be1bfa50cbc0230242719cd2d045d21c2fe44dc80e48dd8fd8ab741be0ce527b8aeada80f8c09f86af83ae6ed4f5e1e3940382bd5b00bf1778c09d402165b36756d6cbc8e7970fed080ba9805b96cc7ba5ff0606b7824585e6d4f318cb7b69dd1d7b5dbd82644aa9d329e73828225521f7e401d568aaef9d790358fc8475154d0c541ac7cef4f4fb51a18ddc5c1fab4f49bb138d4a7dad2f259579e7686819e668c28f024be117cc3aaa7a2dcad77336a0427404ab053cae9db932d988884f9e5c9026e18f762aad9d0bb143edc9e6c808f1f5585ed951103c251dcc331a6db3865faf22b778ce05e68e351608528c8fb1b172b")
	vs200, _ := hex.DecodeString("552102486fd15702c4490a26703112a5cc1d0923fd697a33406bd5a1c00e0013b09a7021024c7b7fb6c310fccf1ba33b082519d82964ea93868d676662d4a59ad548df0e7d2102aaec38470f6aad0042c6e877cfd8087d2676b0f516fddd362801b9bd3936399e2103b209fd4f53a7170ea4444e0cb0a6bb6a53c2bd016926989cf85f9b0fba17a70c2103b8d9d5771d8f513aa0869b9cc8d50986403b78c6da36890638c3d46a5adce04a2102ca0e27697b9c248f6f16e085fd0061e26f44da85b58ee835c110caa5ec3ba5542102df48f60e8f3e01c48ff40b9b7f1310d7a8b2a193188befe1c2e3df740e89509357ae")
	neoHeader200 := &NeoBlockHeader{
		&block.BlockHeader{
			Version:       0,
			PrevHash:      prevHash,
			MerkleRoot:    merKleRoot,
			Timestamp:     1476651132,
			Index:         200,
			NextConsensus: nextConsensus,
			ConsensusData: consensusData,
			Witness: &tx2.Witness{
				InvocationScript:   is200,
				VerificationScript: vs200,
			},
		},
	}
	sink = common.NewZeroCopySink(nil)
	if err := neoHeader200.Serialization(sink); err != nil {
		t.Errorf("NeoBlockHeaderToBytes error:%v", err)
	}
	neoHeader200bs := sink.Bytes()

	neoHandler := NewNEOHandler()
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
