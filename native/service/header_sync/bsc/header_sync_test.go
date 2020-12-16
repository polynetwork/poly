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
package bsc

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly-io-test/chains/eth"
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
)

func init() {
	setBKers()
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
	// errDesc := e.Error()
	// if strings.Contains(errDesc, "ETHHandler GetHeaderByHeight, genesis header had been initialized") {
	// 	return GENESIS_INITIALIZED
	// } else if strings.Contains(errDesc, "ETHHandler SyncGenesisHeader: getGenesisHeader, deserialize header err:") {
	// 	return GENESIS_PARAM_ERROR
	// } else if strings.Contains(errDesc, "SyncBlockHeader, deserialize header err:") {
	// 	return SYNCBLOCK_PARAM_ERROR
	// } else if strings.Contains(errDesc, "SyncBlockHeader, get the parent block failed. Error:") {
	// 	return SYNCBLOCK_ORPHAN
	// } else if strings.Contains(errDesc, "SyncBlockHeader, invalid difficulty:") {
	// 	return DIFFICULTY_ERROR
	// } else if strings.Contains(errDesc, "SyncBlockHeader, verify header error:") {
	// 	return NONCE_ERROR
	// } else if strings.Contains(errDesc, "SyncGenesisHeader, checkWitness error:") {
	// 	return OPERATOR_ERROR
	// }
	return UNKNOWN
}

func getLatestHeight(native *native.NativeService) uint64 {
	height, err := GetCanonicalHeight(native, BSCChainID)
	if err != nil {
		return 0
	}

	return height
}

func getHeaderHashByHeight(native *native.NativeService, height uint64) ethcommon.Hash {
	hws, err := GetCanonicalHeader(native, BSCChainID, height)
	if err != nil {
		return ethcommon.Hash{}
	}

	return hws.Header.Hash()
}

func getHeaderByHash(native *native.NativeService, hash ethcommon.Hash) []byte {
	hws, err := getHeader(native, hash, BSCChainID)
	if err != nil {
		return nil
	}

	headerOnly, _ := json.Marshal(hws.Header)
	return headerOnly
}

const BSCChainID = 2

func getGenesisHeader(t *testing.T) *GenesisHeader {
	tool := eth.NewEthTools("https://data-seed-prebsc-1-s1.binance.org:8545/")
	height, err := tool.GetNodeHeight()
	if err != nil {
		panic(err)
	}

	epochHeight := height - height%200
	pEpochHeight := epochHeight - 200
	ppEpochHeight := pEpochHeight - 200

	hdr, err := tool.GetBlockHeader(epochHeight)
	if err != nil {
		assert.NilError(t, err)
	}
	phdr, err := tool.GetBlockHeader(pEpochHeight)
	if err != nil {
		assert.NilError(t, err)
	}
	pvalidators, err := ParseValidators(phdr.Extra[32 : len(phdr.Extra)-65])
	if err != nil {
		assert.NilError(t, err)
	}
	pphdr, err := tool.GetBlockHeader(ppEpochHeight)
	if err != nil {
		assert.NilError(t, err)
	}
	ppvalidators, err := ParseValidators(pphdr.Extra[32 : len(pphdr.Extra)-65])
	if err != nil {
		assert.NilError(t, err)
	}

	genesisHeader := GenesisHeader{Header: *hdr, PrevValidators: []HeightAndValidators{
		{Height: big.NewInt(int64(pEpochHeight)), Validators: pvalidators},
		{Height: big.NewInt(int64(ppEpochHeight)), Validators: ppvalidators},
	}}

	return &genesisHeader
}

func TestSyncGenesisHeader(t *testing.T) {
	genesisHeader := getGenesisHeader(t)
	headerOnlyBytes, _ := json.Marshal(genesisHeader.Header)
	genesisHeaderBytes, _ := json.Marshal(genesisHeader)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = BSCChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	ethHandler := NewHandler()
	err = ethHandler.SyncGenesisHeader(native)

	assert.Equal(t, SUCCESS, typeOfError(err), err)
	height := getLatestHeight(native)
	assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
	headerHash := getHeaderHashByHeight(native, height)
	assert.Equal(t, true, bytes.Equal(genesisHeader.Header.Hash().Bytes(), headerHash.Bytes()))
	headerFromStore := getHeaderByHash(native, headerHash)
	assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))
}
