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
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
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
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
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

func getTool() *eth.ETHTools {
	tool := eth.NewEthTools("https://data-seed-prebsc-1-s1.binance.org:8545/")
	return tool
}

func getBlockHeader(t *testing.T, height uint64) *etypes.Header {
	tool := getTool()
	hdr, err := tool.GetBlockHeader(height)
	assert.NilError(t, err)
	return hdr
}

func getGenesisHeader(t *testing.T) *GenesisHeader {
	tool := getTool()
	height, err := tool.GetNodeHeight()
	if err != nil {
		panic(err)
	}

	epochHeight := height - height%200
	pEpochHeight := epochHeight - 200

	hdr, err := tool.GetBlockHeader(epochHeight)
	assert.NilError(t, err)
	phdr, err := tool.GetBlockHeader(pEpochHeight)
	assert.NilError(t, err)
	pvalidators, err := ParseValidators(phdr.Extra[32 : len(phdr.Extra)-65])
	assert.NilError(t, err)

	genesisHeader := GenesisHeader{Header: *hdr, PrevValidators: []HeightAndValidators{
		{Height: big.NewInt(int64(pEpochHeight)), Validators: pvalidators},
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
	handler := NewHandler()
	err = handler.SyncGenesisHeader(native)

	assert.Equal(t, SUCCESS, typeOfError(err), err)
	height := getLatestHeight(native)
	assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
	headerHash := getHeaderHashByHeight(native, height)
	assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
	headerFromStore := getHeaderByHash(native, headerHash)
	assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))

}

func TestSyncGenesisHeaderNoOperator(t *testing.T) {

	genesisHeader := getGenesisHeader(t)
	genesisHeaderBytes, _ := json.Marshal(genesisHeader)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = BSCChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{}

	native, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	handler := NewHandler()
	err = handler.SyncGenesisHeader(native)

	assert.Equal(t, OPERATOR_ERROR, typeOfError(err), err)
	height := getLatestHeight(native)
	assert.Equal(t, uint64(0), height)

}

func TestSyncGenesisHeaderTwice(t *testing.T) {

	var (
		native *native.NativeService
		height uint64
		err    error
	)

	{
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

		native, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)
		handler := NewHandler()
		err = handler.SyncGenesisHeader(native)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(native)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(native, height)
		assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
		headerFromStore := getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))
	}

	{
		genesisHeader := getGenesisHeader(t)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = BSCChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		assert.NilError(t, err)
		handler := NewHandler()
		err = handler.SyncGenesisHeader(native)

		assert.Equal(t, GENESIS_INITIALIZED, typeOfError(err), err)
		assert.Equal(t, getLatestHeight(native), height)
	}

}

func TestSyncGenesisHeader_ParamError(t *testing.T) {

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = BSCChainID
	param.GenesisHeader = nil
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, _ := NewNative(sink.Bytes(), tx, nil)
	handler := NewHandler()
	err := handler.SyncGenesisHeader(native)
	assert.Equal(t, GENESIS_PARAM_ERROR, typeOfError(err), err)

}

func untilGetBlockHeader(t *testing.T, height uint64) *etypes.Header {
	tool := getTool()
	for {
		hdr, err := tool.GetBlockHeader(height)
		if err == nil {
			return hdr
		}
		time.Sleep(time.Second)
		fmt.Println("GetBlockHeader", height, err)
	}
}

func TestSyncBlockHeader(t *testing.T) {

	var (
		native *native.NativeService
		height uint64
		err    error
	)

	handler := NewHandler()

	{
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

		native, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)
		err = handler.SyncGenesisHeader(native)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(native)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(native, height)
		assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
		headerFromStore := getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))
	}

	{

		n1 := untilGetBlockHeader(t, height+1)
		n1Bytes, _ := json.Marshal(n1)
		n2 := untilGetBlockHeader(t, height+2)
		n2Bytes, _ := json.Marshal(n2)
		n3 := untilGetBlockHeader(t, height+3)
		n3Bytes, _ := json.Marshal(n3)
		n4 := untilGetBlockHeader(t, height+4)
		n4Bytes, _ := json.Marshal(n4)
		h2h := map[uint64]*etypes.Header{
			height + 1: n1,
			height + 2: n2,
			height + 3: n3,
			height + 4: n4,
		}

		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = BSCChainID
		param.Address = acct.Address
		param.Headers = append(param.Headers, n1Bytes)
		param.Headers = append(param.Headers, n2Bytes)
		param.Headers = append(param.Headers, n3Bytes)
		param.Headers = append(param.Headers, n4Bytes)
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, _ = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		{
			// add sidechain info
			extra := ExtraInfo{
				// test id 97
				ChainID: big.NewInt(97),
			}
			extraBytes, _ := json.Marshal(extra)
			side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
				ExtraInfo: extraBytes,
				ChainId:   BSCChainID,
			})
		}

		// fmt.Println("gHeight", height)
		err := handler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(native)
		assert.Equal(t, latestHeight, height+4)

		for h := height + 1; h <= height+4; h++ {
			headerHash := getHeaderHashByHeight(native, h)
			assert.Equal(t, true, headerHash == h2h[h].Hash())
			headerBytesFromStore := getHeaderByHash(native, headerHash)
			headerBytes, _ := json.Marshal(h2h[h])
			assert.Equal(t, true, bytes.Equal(headerBytesFromStore, headerBytes))
		}
	}
}
