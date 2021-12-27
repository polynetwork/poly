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

package pixiechain

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"

	ethcommon "github.com/ethereum/go-ethereum/common"
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
	"github.com/polynetwork/poly/native/service/header_sync/eth"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"gotest.tools/assert"
)

type PixieChainClient struct {
	addr       string
	restClient *http.Client
}

type heightReq struct {
	JsonRpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      uint     `json:"id"`
}

type heightRep struct {
	JsonRpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      uint   `json:"id"`
}

type BlockReq struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint          `json:"id"`
}

type BlockRep struct {
	JsonRPC string      `json:"jsonrpc"`
	Result  *eth.Header `json:"result"`
	Id      uint        `json:"id"`
}

const (
	// PixieChain Testnet chainId
	pixieTestnetChainID uint64 = 666

	// PixieChain Testnet RPC
	pixieTestnetRPC = "https://http-testnet.chain.pixie.xyz"
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

var (
	acct     = account.NewAccount("")
	setBKers = func() {
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

func (client *PixieChainClient) SendRestRequest(data []byte) ([]byte, error) {
	resp, err := client.restClient.Post(client.addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("http post request:%s error:%s", data, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rest response body error:%s", err)
	}
	return body, nil
}

func (client *PixieChainClient) GetNodeHeight() (uint64, error) {
	req := &heightReq{
		JsonRpc: "2.0",
		Method:  "eth_blockNumber",
		Params:  make([]string, 0),
		Id:      1,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight: marshal req err: %s", err)
	}
	resp, err := client.SendRestRequest(data)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight err: %s", err)
	}
	rep := &heightRep{}
	err = json.Unmarshal(resp, rep)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight, unmarshal resp err: %s", err)
	}
	height, err := strconv.ParseUint(rep.Result, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight, parse resp height %s failed", rep.Result)
	} else {
		return height, nil
	}
}

func (client *PixieChainClient) GetBlockHeader(height uint64) (*eth.Header, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), true}
	req := &BlockReq{
		JsonRpc: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  params,
		Id:      1,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("GetNodeHeight: marshal req err: %s", err)
	}
	resp, err := client.SendRestRequest(data)
	if err != nil {
		return nil, fmt.Errorf("GetNodeHeight err: %s", err)
	}
	rsp := &BlockRep{}
	err = json.Unmarshal(resp, rsp)
	if err != nil {
		return nil, fmt.Errorf("GetNodeHeight, unmarshal resp err: %s", err)
	}

	return rsp.Result, nil
}

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
	height, err := GetCanonicalHeight(native, pixieTestnetChainID)
	if err != nil {
		return 0
	}

	return height
}

func getHeaderHashByHeight(native *native.NativeService, height uint64) ethcommon.Hash {
	hws, err := GetCanonicalHeader(native, pixieTestnetChainID, height)
	if err != nil {
		return ethcommon.Hash{}
	}

	return hws.Header.Hash()
}

func getHeaderByHash(native *native.NativeService, hash ethcommon.Hash) []byte {
	hws, err := getHeader(native, hash, pixieTestnetChainID)
	if err != nil {
		return nil
	}

	headerOnly, _ := json.Marshal(hws.Header)
	return headerOnly
}

func newPixieChainClient() *PixieChainClient {
	c := &PixieChainClient{
		addr: pixieTestnetRPC,
		restClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   5,
				DisableKeepAlives:     false,
				IdleConnTimeout:       time.Second * 300,
				ResponseHeaderTimeout: time.Second * 300,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 300,
		},
	}
	return c
}

func getBlockHeader(t *testing.T, height uint64) *eth.Header {
	c := newPixieChainClient()
	hdr, err := c.GetBlockHeader(height)
	assert.NilError(t, err)
	return hdr
}

func getGenesisHeader(t *testing.T) *GenesisHeader {
	c := newPixieChainClient()
	height, err := c.GetNodeHeight()
	if err != nil {
		panic(err)
	}

	var backOffHeight uint64 = 200 * 5

	epochHeight := height - height%200 - backOffHeight
	pEpochHeight := epochHeight - 200 - backOffHeight

	hdr, err := c.GetBlockHeader(epochHeight)
	assert.NilError(t, err)
	phdr, err := c.GetBlockHeader(pEpochHeight)
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
	param.ChainID = pixieTestnetChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	testNative, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	handler := NewPixieHandler()
	err = handler.SyncGenesisHeader(testNative)

	assert.Equal(t, SUCCESS, typeOfError(err), err)
	height := getLatestHeight(testNative)
	assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
	headerHash := getHeaderHashByHeight(testNative, height)
	assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
	headerFromStore := getHeaderByHash(testNative, headerHash)
	assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))

}

func TestSyncGenesisHeaderNoOperator(t *testing.T) {

	genesisHeader := getGenesisHeader(t)
	genesisHeaderBytes, _ := json.Marshal(genesisHeader)

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = pixieTestnetChainID
	param.GenesisHeader = genesisHeaderBytes
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{}

	testNative, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	handler := NewPixieHandler()
	err = handler.SyncGenesisHeader(testNative)

	assert.Equal(t, OPERATOR_ERROR, typeOfError(err), err)
	height := getLatestHeight(testNative)
	assert.Equal(t, uint64(0), height)

}

func TestSyncGenesisHeaderTwice(t *testing.T) {

	var (
		testNative *native.NativeService
		height     uint64
		err        error
	)

	{
		genesisHeader := getGenesisHeader(t)
		headerOnlyBytes, _ := json.Marshal(genesisHeader.Header)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		testNative, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)
		handler := NewPixieHandler()
		err = handler.SyncGenesisHeader(testNative)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(testNative)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(testNative, height)
		assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
		headerFromStore := getHeaderByHash(testNative, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))
	}

	{
		genesisHeader := getGenesisHeader(t)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		testNative, err = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())
		assert.NilError(t, err)
		handler := NewPixieHandler()
		err = handler.SyncGenesisHeader(testNative)

		assert.Equal(t, GENESIS_INITIALIZED, typeOfError(err), err)
		assert.Equal(t, getLatestHeight(testNative), height)
	}

}

func TestSyncGenesisHeader_ParamError(t *testing.T) {

	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = pixieTestnetChainID
	param.GenesisHeader = nil
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	testNative, _ := NewNative(sink.Bytes(), tx, nil)
	handler := NewPixieHandler()
	err := handler.SyncGenesisHeader(testNative)
	assert.Equal(t, GENESIS_PARAM_ERROR, typeOfError(err), err)

}

func untilGetBlockHeader(t *testing.T, height uint64) *eth.Header {
	c := newPixieChainClient()
	for {
		hdr, err := c.GetBlockHeader(height)
		if err == nil {
			return hdr
		}
		// time.Sleep(time.Second)
		t.Error("GetBlockHeader", height, err)
	}
}

func TestSyncBlockHeader(t *testing.T) {
	var (
		testNative *native.NativeService
		height     uint64
		err        error
	)

	handler := NewPixieHandler()

	{
		testNative, err = NewNative(nil, &types.Transaction{}, nil)
		// add sidechain info
		extra := ExtraInfo{
			ChainID: big.NewInt(int64(pixieTestnetChainID)),
			Period:  3,
		}
		extraBytes, _ := json.Marshal(extra)
		err = side_chain_manager.PutSideChain(testNative, &side_chain_manager.SideChain{
			ExtraInfo: extraBytes,
			ChainId:   pixieTestnetChainID,
		})
		assert.NilError(t, err)

		sideInfo, err := side_chain_manager.GetSideChain(testNative, pixieTestnetChainID)
		assert.NilError(t, err)
		assert.Equal(t, true, bytes.Equal(sideInfo.ExtraInfo, extraBytes))
		assert.Equal(t, sideInfo.ChainId, pixieTestnetChainID)
	}

	{
		genesisHeader := getGenesisHeader(t)
		headerOnlyBytes, _ := json.Marshal(genesisHeader.Header)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		testNative, err = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())
		assert.NilError(t, err)
		err = handler.SyncGenesisHeader(testNative)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(testNative)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(testNative, height)
		assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
		headerFromStore := getHeaderByHash(testNative, headerHash)
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
		h2h := map[uint64]*eth.Header{
			height + 1: n1,
			height + 2: n2,
			height + 3: n3,
			height + 4: n4,
		}

		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = pixieTestnetChainID
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
		testNative, _ = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())

		// fmt.Println("gHeight", height)
		err = handler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(testNative)
		assert.Equal(t, latestHeight, height+4)

		for h := height + 1; h <= height+4; h++ {
			headerHash := getHeaderHashByHeight(testNative, h)
			assert.Equal(t, true, headerHash == h2h[h].Hash())
			headerBytesFromStore := getHeaderByHash(testNative, headerHash)
			headerBytes, _ := json.Marshal(h2h[h])
			assert.Equal(t, true, bytes.Equal(headerBytesFromStore, headerBytes))
		}
	}
	// check find previous validators and verify header process
	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.Address = acct.Address
		height = getLatestHeight(testNative)

		var headerNumber uint64 = 500
		for i := 1; i <= int(headerNumber); i++ {
			headerBs, _ := json.Marshal(untilGetBlockHeader(t, height+uint64(i)))
			param.Headers = append(param.Headers, headerBs)
		}

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		testNative, _ = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())

		err := handler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(testNative)
		assert.Equal(t, latestHeight, height+headerNumber)

		for h := height + 1; h <= height+headerNumber; h++ {
			headerHash := getHeaderHashByHeight(testNative, h)
			headerBs := param.Headers[h-height-1]
			var oheader eth.Header
			if err := json.Unmarshal(headerBs, &oheader); err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, true, headerHash == oheader.Hash())
			headerBytesFromStore := getHeaderByHash(testNative, headerHash)
			assert.Equal(t, true, bytes.Equal(headerBytesFromStore, headerBs))
		}
	}
}

func TestSyncForkBlockHeader(t *testing.T) {

	var (
		testNative *native.NativeService
		height     uint64
		err        error
	)

	handler := NewPixieHandler()

	{
		testNative, err = NewNative(nil, &types.Transaction{}, nil)
		// add sidechain info
		extra := ExtraInfo{
			ChainID: big.NewInt(int64(pixieTestnetChainID)),
			Period:  3,
		}
		extraBytes, _ := json.Marshal(extra)
		err = side_chain_manager.PutSideChain(testNative, &side_chain_manager.SideChain{
			ExtraInfo: extraBytes,
			ChainId:   pixieTestnetChainID,
		})
		assert.NilError(t, err)

		sideInfo, err := side_chain_manager.GetSideChain(testNative, pixieTestnetChainID)
		assert.NilError(t, err)
		assert.Equal(t, true, bytes.Equal(sideInfo.ExtraInfo, extraBytes))
		assert.Equal(t, sideInfo.ChainId, pixieTestnetChainID)
	}

	{
		genesisHeader := getGenesisHeader(t)
		headerOnlyBytes, _ := json.Marshal(genesisHeader.Header)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		testNative, err = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())
		assert.NilError(t, err)
		err = handler.SyncGenesisHeader(testNative)

		assert.Equal(t, SUCCESS, typeOfError(err), err)
		height = getLatestHeight(testNative)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)
		headerHash := getHeaderHashByHeight(testNative, height)
		assert.Equal(t, true, genesisHeader.Header.Hash() == headerHash)
		headerFromStore := getHeaderByHash(testNative, headerHash)
		assert.Equal(t, true, bytes.Equal(headerFromStore, headerOnlyBytes))
	}

	// check find previous validators and verify header process
	{
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.Address = acct.Address
		height = getLatestHeight(testNative)

		var headerNumber uint64 = 250
		for i := 1; i <= int(headerNumber); i++ {
			headerBs, _ := json.Marshal(untilGetBlockHeader(t, height+uint64(i)))
			param.Headers = append(param.Headers, headerBs)
		}

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		testNative, _ = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())

		err := handler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(testNative)
		assert.Equal(t, latestHeight, height+headerNumber)

		for h := height + 1; h <= height+headerNumber; h++ {
			headerHash := getHeaderHashByHeight(testNative, h)
			headerBs := param.Headers[h-height-1]
			var oheader eth.Header
			if err := json.Unmarshal(headerBs, &oheader); err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, true, headerHash == oheader.Hash())
			headerBytesFromStore := getHeaderByHash(testNative, headerHash)
			assert.Equal(t, true, bytes.Equal(headerBytesFromStore, headerBs))
		}
	}
	// check forked chain can come back to normal
	TestFlagNoCheckPixieHeaderSig = true

	{
		// first sync forked headers
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.Address = acct.Address
		height = getLatestHeight(testNative)

		var headerNumber uint64 = 5
		realHeaders := make([]*eth.Header, 0)
		forkedHeaders := make([]*eth.Header, 0)
		for i := 1; i <= int(headerNumber); i++ {
			headerI := untilGetBlockHeader(t, height+uint64(i))
			realHeaders = append(realHeaders, headerI)

			forkHeader := copyHeader(headerI)
			forkHeader.ReceiptHash = ethcommon.Hash{}
			forkedHeaders = append(forkedHeaders, forkHeader)

		}
		for i := 1; i < int(headerNumber); i++ {
			forkedHeaders[i].ParentHash = forkedHeaders[i-1].Hash()
		}
		for _, v := range forkedHeaders {
			forkedHeaderBs, _ := json.Marshal(v)
			param.Headers = append(param.Headers, forkedHeaderBs)
		}

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		testNative, _ = NewNative(sink.Bytes(), &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}, testNative.GetCacheDB())

		err := handler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(testNative)
		assert.Equal(t, latestHeight, height+headerNumber)

		for h := height + 1; h <= height+headerNumber; h++ {
			headerHash := getHeaderHashByHeight(testNative, h)
			forkheader := forkedHeaders[h-height-1]
			index := h - height - 1
			assert.Equal(t, true, headerHash == forkedHeaders[index].Hash())
			headerBytesFromStore := getHeaderByHash(testNative, headerHash)
			forkHeaderBs, _ := json.Marshal(forkheader)
			assert.Equal(t, true, bytes.Equal(headerBytesFromStore, forkHeaderBs))
		}
		// second sync normal header
		param.Headers = make([][]byte, 0)
		for _, v := range realHeaders {
			realHeaderBs, _ := json.Marshal(v)
			param.Headers = append(param.Headers, realHeaderBs)
		}
		var newHeaderNum uint64 = 1
		oldHeaderLen := uint64(len(param.Headers))
		for i := 1; i <= int(newHeaderNum); i++ {
			headerI := untilGetBlockHeader(t, height+oldHeaderLen+uint64(i))
			realHeaders = append(realHeaders, headerI)
			realHeaderBs, _ := json.Marshal(headerI)
			param.Headers = append(param.Headers, realHeaderBs)
		}

		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		testNative, _ = NewNative(sink.Bytes(), &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}, testNative.GetCacheDB())

		err = handler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight = getLatestHeight(testNative)
		assert.Equal(t, latestHeight, height+headerNumber+newHeaderNum)

		for h := height + 1; h <= height+headerNumber+newHeaderNum; h++ {
			headerHash := getHeaderHashByHeight(testNative, h)
			index := h - height - 1
			realheader := realHeaders[index]
			assert.Equal(t, true, headerHash == realHeaders[index].Hash())
			headerBytesFromStore := getHeaderByHash(testNative, headerHash)
			realHeaderBs, _ := json.Marshal(realheader)
			assert.Equal(t, true, bytes.Equal(headerBytesFromStore, realHeaderBs))
		}

	}

}

func copyHeader(h *eth.Header) *eth.Header {
	cpy := *h
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}
