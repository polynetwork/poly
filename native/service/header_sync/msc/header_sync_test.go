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
package msc

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
	height, err := GetCanonicalHeight(native, MSCChainID)
	if err != nil {
		return 0
	}

	return height
}

func getHeaderHashByHeight(native *native.NativeService, height uint64) ethcommon.Hash {
	hws, err := GetCanonicalHeader(native, MSCChainID, height)
	if err != nil {
		return ethcommon.Hash{}
	}

	return hws.Header.Hash()
}

func getHeaderByHash(native *native.NativeService, hash ethcommon.Hash) []byte {
	hws, err := getHeader(native, hash, MSCChainID)
	if err != nil {
		return nil
	}

	headerOnly, _ := json.Marshal(hws.Header)
	return headerOnly
}

const MSCChainID = 2

func getTool() *eth.ETHTools {
	tool := eth.NewEthTools("http://49.234.157.164:8545/")
	return tool
}

func getBlockHeaderByHash(t *testing.T, hash ethcommon.Hash) *etypes.Header {
	tool := getTool()
	hdr, err := tool.GetBlockHeaderByHash(hash)
	assert.NilError(t, err)
	return hdr
}

func getBlockHeader(t *testing.T, height uint64) *etypes.Header {
	tool := getTool()
	hdr, err := tool.GetBlockHeader(height)
	assert.NilError(t, err)

	return hdr
}

func getGenesisHeaderByHeight(t *testing.T, epochHeight uint64) *etypes.Header {

	tool := getTool()
	hdr, err := tool.GetBlockHeader(epochHeight)
	assert.NilError(t, err)

	_, err = ecrecover(hdr)
	assert.NilError(t, err)

	return hdr
}

func getGenesisHeader(t *testing.T) *etypes.Header {
	tool := getTool()
	height, err := tool.GetNodeHeight()
	if err != nil {
		panic(err)
	}

	epochHeight := height - height%30000
	return getGenesisHeaderByHeight(t, epochHeight)
}

func TestSyncGenesisHeader(t *testing.T) {

	genesisHeader := getGenesisHeader(t)
	genesisHeaderBytes, _ := json.Marshal(genesisHeader)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = MSCChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	{
		// add sidechain info
		extra := ExtraInfo{
			// test id 97
			ChainID: big.NewInt(97),
			Period:  1,
			Epoch:   200,
		}
		extraBytes, _ := json.Marshal(extra)
		side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
			ExtraInfo: extraBytes,
			ChainId:   MSCChainID,
		})
	}
	handler := NewHandler()
	err = handler.SyncGenesisHeader(native)

	assert.Equal(t, SUCCESS, typeOfError(err), err)
	height := getLatestHeight(native)
	assert.Equal(t, genesisHeader.Number.Uint64(), height)
	headerHash := getHeaderHashByHeight(native, height)
	assert.Equal(t, true, (*etypes.Header)(genesisHeader).Hash() == headerHash)
	headerFromStore := getHeaderByHash(native, headerHash)
	assert.Equal(t, true, bytes.Equal(headerFromStore, genesisHeaderBytes))

}

func TestSyncGenesisHeaderNoOperator(t *testing.T) {

	genesisHeader := getGenesisHeader(t)
	genesisHeaderBytes, _ := json.Marshal(genesisHeader)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = MSCChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{}

	native, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	{
		// add sidechain info
		extra := ExtraInfo{
			// test id 97
			ChainID: big.NewInt(97),
			Period:  1,
			Epoch:   200,
		}
		extraBytes, _ := json.Marshal(extra)
		side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
			ExtraInfo: extraBytes,
			ChainId:   MSCChainID,
		})
	}
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
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = MSCChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, nil)
		{
			// add sidechain info
			extra := ExtraInfo{
				// test id 97
				ChainID: big.NewInt(97),
				Period:  1,
				Epoch:   200,
			}
			extraBytes, _ := json.Marshal(extra)
			side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
				ExtraInfo: extraBytes,
				ChainId:   MSCChainID,
			})
		}
		assert.NilError(t, err)
		handler := NewHandler()
		err = handler.SyncGenesisHeader(native)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(native)
		assert.Equal(t, genesisHeader.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(native, height)
		assert.Equal(t, true, (*etypes.Header)(genesisHeader).Hash() == headerHash)
		headerFromStore := getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, genesisHeaderBytes))
	}

	{
		genesisHeader := getGenesisHeader(t)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = MSCChainID
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
	param.ChainID = MSCChainID
	param.GenesisHeader = nil
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	native, _ := NewNative(sink.Bytes(), tx, nil)
	{
		// add sidechain info
		extra := ExtraInfo{
			// test id 97
			ChainID: big.NewInt(97),
			Period:  1,
			Epoch:   200,
		}
		extraBytes, _ := json.Marshal(extra)
		side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
			ExtraInfo: extraBytes,
			ChainId:   MSCChainID,
		})
	}
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
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = MSCChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)

		{
			// add sidechain info
			extra := ExtraInfo{
				// test id 97
				ChainID: big.NewInt(97),
				Period:  1,
				Epoch:   200,
			}
			extraBytes, _ := json.Marshal(extra)
			side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
				ExtraInfo: extraBytes,
				ChainId:   MSCChainID,
			})
		}

		err = handler.SyncGenesisHeader(native)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(native)
		assert.Equal(t, genesisHeader.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(native, height)
		assert.Equal(t, true, (*etypes.Header)(genesisHeader).Hash() == headerHash)
		headerFromStore := getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, genesisHeaderBytes), fmt.Sprintf("%s vs %s", string(headerFromStore), string(genesisHeaderBytes)))
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
		param.ChainID = MSCChainID
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

func TestSyncBlockHeaderForked(t *testing.T) {
	var (
		native *native.NativeService
		height uint64
		err    error
	)

	handler := NewHandler()

	{
		genesisHeader := getGenesisHeader(t)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = MSCChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)

		{
			// add sidechain info
			extra := ExtraInfo{
				// test id 97
				ChainID: big.NewInt(97),
				Period:  1,
				Epoch:   200,
			}
			extraBytes, _ := json.Marshal(extra)
			side_chain_manager.PutSideChain(native, &side_chain_manager.SideChain{
				ExtraInfo: extraBytes,
				ChainId:   MSCChainID,
			})
		}

		err = handler.SyncGenesisHeader(native)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(native)
		assert.Equal(t, genesisHeader.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(native, height)
		assert.Equal(t, true, (*etypes.Header)(genesisHeader).Hash() == headerHash)
		headerFromStore := getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, genesisHeaderBytes))
	}

	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = MSCChainID
		param.Address = acct.Address
		for i := 1; i < 10; i++ {
			header := untilGetBlockHeader(t, height+uint64(i))
			headerBytes, _ := json.Marshal(header)
			param.Headers = append(param.Headers, headerBytes)
		}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, _ = NewNative(sink.Bytes(), tx, native.GetCacheDB())

		// fmt.Println("gHeight", height)
		err := handler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(native)
		assert.Equal(t, latestHeight, height+9)
	}

	interestedHeight := height + 10

	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = MSCChainID
		param.Address = acct.Address

		header := getBlockHeader(t, interestedHeight)
		signer, err := ecrecover(header)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		mockSigner = signer
		oldHash := header.Hash()
		header.ReceiptHash = ethcommon.Hash{}
		assert.Equal(t, true, oldHash != header.Hash())
		headerBytes, _ := json.Marshal(header)
		param.Headers = append(param.Headers, headerBytes)

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, _ = NewNative(sink.Bytes(), tx, native.GetCacheDB())

		err = handler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(native)
		assert.Equal(t, latestHeight, interestedHeight)
		headerHash := getHeaderHashByHeight(native, interestedHeight)
		assert.Equal(t, true, header.Hash() == headerHash)
		headerFromStore := getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, headerBytes))

		mockSigner = ethcommon.Address{}

		// add real header
		param = new(scom.SyncBlockHeaderParam)
		param.ChainID = MSCChainID
		param.Address = acct.Address

		realHeader := getBlockHeader(t, interestedHeight)
		realHeaderBytes, _ := json.Marshal(realHeader)
		param.Headers = append(param.Headers, realHeaderBytes)

		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx = &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, _ = NewNative(sink.Bytes(), tx, native.GetCacheDB())

		err = handler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight = getLatestHeight(native)
		assert.Equal(t, latestHeight, interestedHeight)
		headerHash = getHeaderHashByHeight(native, interestedHeight)
		// still the old header
		assert.Equal(t, true, header.Hash() == headerHash)
		headerFromStore = getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, headerBytes))

		// add next header
		param = new(scom.SyncBlockHeaderParam)
		param.ChainID = MSCChainID
		param.Address = acct.Address

		nextHeader := getBlockHeader(t, interestedHeight+1)
		nextHeaderBytes, _ := json.Marshal(nextHeader)
		param.Headers = append(param.Headers, nextHeaderBytes)

		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx = &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, _ = NewNative(sink.Bytes(), tx, native.GetCacheDB())

		err = handler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight = getLatestHeight(native)
		assert.Equal(t, latestHeight, interestedHeight+1)
		headerHash = getHeaderHashByHeight(native, interestedHeight)
		// change to the new header
		assert.Equal(t, true, realHeader.Hash() == headerHash)
		headerFromStore = getHeaderByHash(native, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, realHeaderBytes))
	}

}
