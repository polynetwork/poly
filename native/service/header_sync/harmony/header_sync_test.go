/*
 * Copyright (C) 2022 The poly network Authors
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

package harmony

import (
	"encoding/hex"
	"encoding/json"
	"gotest.tools/assert"
	"testing"

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
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
)

var (
	acct     = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
	harmonyChainID = uint64(2)

	service *native.NativeService
	handler = NewHandler()
)

func init() {
	setBKers()
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) (service *native.NativeService, err error) {
	shouldInit := db == nil
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
	service, err = native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		return
	}
	if shouldInit {
		extraInfo := Context{
			NetworkID: 0, // Harmony Mainnet
		}
		extraInfoBytes, _ := json.Marshal(extraInfo)
		err = side_chain_manager.PutSideChain(service, &side_chain_manager.SideChain{
			ChainId:   harmonyChainID,
			ExtraInfo: extraInfoBytes,
		})
		if err != nil {
			return
		}
	}
	return
}

func TestHandler_SyncGenesisHeader(t *testing.T) {

	// Test genesis header
	testCase := func (headerHex string, shouldSucceed bool, msg string) {
		data, err := hex.DecodeString(headerHex)
		assert.NilError(t, err)
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = harmonyChainID
		param.GenesisHeader = data
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		service, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)
		err = handler.SyncGenesisHeader(service)
		if shouldSucceed {
			assert.NilError(t, err, "HarmonyHandler failed to sync genesis header")
		} else {
			assert.Error(t, err, msg)
		}
	}

	// Not epoch last block
	{
		headerHex:= "f90357b902d1f902ce87486d6e79546764827631f902c0a05ce7c168d0ef6b43cbe59c1fa3f95eda43013979a611544e30d374f976bedca19445df588b05a675b5f16e253aa15d90333df9fedfa0a4c634947a180770aa009171dd41281cbe9820ec85fc7186cd76b2d7fa31afd3a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000831abffe8806f05b59d3b2000080845df44ea680a00000000000000000000000000000000000000000000000000000000000000000831ac02e5680b860ab9945cb94578c7013c5dcca42064015af4c04dbb9d542799b35613f09b6149627e728ae445a69b7897e7ad53b28a503819c9d0d59a628acf64909fffe0fa213cde72892b718136895c50a15675307a4efe64a55e2a6186171f0e8f01c901086a07fffdfbf3cbd5ceffff9ffeb6fafffffffffffffffffffffffbfffffbfffff03a00000000000000000000000000000000000000000000000000000000000000000808080b860565949c5dfe0aabd13019848f2c9abd96ac917923b21af2c2272227644f328b4e62ca8d4fa8005fa36d6bc0ec6975202eba1243a42e04b25f75a4aa7828a17c5739d1a626697d38ca3dee242745aead4c6086579d950607cc545118927415f8aa07fffdfbf3cbd5eeffffdffeb6fafffffffffffffffffffffffbfffffbfffff03"
		testCase(
			headerHex,
			false,
			"HarmonyHandler, failed to extract Epoch from header, err: unexpected empty shard state in header",
		)
	}
	// epoch last block 1753087
	{
		testCase(
			headerHex1753087,
			true, "",
		)
	}
}

func TestHandler_SyncBlockHeader(t *testing.T) {
	// Register genesis header first
	TestHandler_SyncGenesisHeader(t)

	// Test epoch sync
	testCase := func (headerHex string, shouldSucceed bool, msg string) {
		data, err := hex.DecodeString(headerHex)
		assert.NilError(t, err)
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = harmonyChainID
		param.Headers = [][]byte{data}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		service, err = NewNative(sink.Bytes(), tx, service.GetCacheDB())
		assert.NilError(t, err)
		err = handler.SyncBlockHeader(service)
		if shouldSucceed {
			assert.NilError(t, err, "HarmonyHandler failed to sync block header")
		} else {
			assert.Error(t, err, msg)
		}
	}

	// Not epoch last block
	{
		testCase(
			headerHex1769470,
			false, "HarmonyHandler, failed to extract Epoch from header, err: unexpected empty shard state in header")
	}

	// last block of next epoch
	{
		testCase(
			headerHex1785855,
			false, "HarmonyHandler failed to validate next epoch, idx 0 block 1785855, err: epoch does not match, current 87, got 88")
	}

	// last block of current epoch
	{
		testCase(
			headerHex1769471,
			true, "")
	}
}

