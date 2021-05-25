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
package kai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	kaiclient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/bytes"
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
	"gotest.tools/assert"
)

var (
	acct     *account.Account = account.NewAccount("")
	setBKers                  = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
	tool kaiclient.Node
)

func newKaiClient() kaiclient.Node {
	url := "https://dev-6.kardiachain.io"
	node, err := kaiclient.NewNode(url, nil)
	if err != nil {
		panic(err)
	}
	return node
}

func init() {
	setBKers()
	tool = newKaiClient()
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) (*native.NativeService, error) {
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
	return native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
}

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

func typeOfError(e error) int {
	if e == nil {
		return SUCCESS
	}
	errDesc := e.Error()
	if strings.Contains(errDesc, "genesis had been initialized") {
		return GENESIS_INITIALIZED
	} else if strings.Contains(errDesc, "deserialize GenesisHeader err") {
		return GENESIS_PARAM_ERROR
		// } else if strings.Contains(errDesc, "SyncBlockHeader, deserialize header err:") {
		// 	return SYNCBLOCK_PARAM_ERROR
		// } else if strings.Contains(errDesc, "SyncBlockHeader, get the parent block failed. Error:") {
		// 	return SYNCBLOCK_ORPHAN
		// } else if strings.Contains(errDesc, "SyncBlockHeader, invalid difficulty:") {
		// 	return DIFFICULTY_ERROR
		// } else if strings.Contains(errDesc, "SyncBlockHeader, verify header error:") {
		// 	return NONCE_ERROR
	} else if strings.Contains(errDesc, "SyncGenesisHeader, checkWitness error:") {
		return OPERATOR_ERROR
	}
	return UNKNOWN
}

const ChainID = 138

func getGenesisHeaderByHeight(t *testing.T, epochHeight uint64) *kaiclient.FullHeader {
	h, err := tool.FullHeaderByNumber(context.Background(), epochHeight)
	assert.NilError(t, err)
	return h
}

func getGenesisHeader(t *testing.T) *kaiclient.FullHeader {
	return getGenesisHeaderByHeight(t, 5)
}

func TestSyncGenesisHeader(t *testing.T) {

	genesisHeader := getGenesisHeader(t)
	genesisHeaderBytes, err := json.Marshal(genesisHeader)
	assert.NilError(t, err)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	handler := NewHandler()
	err = handler.SyncGenesisHeader(native)
	assert.NilError(t, err)

	epochInfo, err := GetEpochSwitchInfo(native, 138)
	assert.NilError(t, err)
	assert.Equal(t, epochInfo.Height, int64(5))
	assert.Equal(t, epochInfo.NextValidatorsHash.String(), bytes.HexBytes(genesisHeader.Header.NextValidatorsHash.Bytes()).String())
	assert.Equal(t, epochInfo.ChainID, "138")
	assert.Equal(t, epochInfo.BlockHash.String(), bytes.HexBytes(genesisHeader.Header.Hash().Bytes()).String())
}

func syncGenesisBlockHeader(t *testing.T) *native.NativeService {
	genesisHeader := getGenesisHeader(t)
	genesisHeaderBytes, err := json.Marshal(genesisHeader)
	assert.NilError(t, err)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	handler := NewHandler()
	err = handler.SyncGenesisHeader(native)
	assert.NilError(t, err)
	return native
}

func TestSyncBlockHeader(t *testing.T) {
	native := syncGenesisBlockHeader(t)

	param := new(scom.SyncBlockHeaderParam)
	param.ChainID = ChainID
	param.Address = acct.Address

	h1 := getGenesisHeaderByHeight(t, 6)
	hb1, err := json.Marshal(h1)
	assert.NilError(t, err)

	h2 := getGenesisHeaderByHeight(t, 7)
	hb2, err := json.Marshal(h2)
	assert.NilError(t, err)

	param.Headers = append(param.Headers, hb1)
	param.Headers = append(param.Headers, hb2)

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, err = NewNative(sink.Bytes(), tx, native.GetCacheDB())
	assert.NilError(t, err)
	handler := NewHandler()
	_ = handler.SyncBlockHeader(native)
}

func TestVerifyHeader(t *testing.T) {
	native := syncGenesisBlockHeader(t)
	epochInfo, err := GetEpochSwitchInfo(native, 138)
	assert.NilError(t, err)
	h6 := getGenesisHeaderByHeight(t, 6)
	err = VerifyHeader(h6, epochInfo)
	assert.NilError(t, err)
}
