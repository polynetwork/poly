package pixiechain

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/states"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/genesis"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	ccmcom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	eth2 "github.com/polynetwork/poly/native/service/cross_chain_manager/eth"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	synccom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/eth"
	"github.com/polynetwork/poly/native/service/header_sync/pixiechain"
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

type proofReq struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint          `json:"id"`
}

type ProofRep struct {
	JsonRPC string `json:"jsonrpc"`
	Result  Proof  `json:"proof"`
	Id      uint   `json:"id"`
}

var (
	acct     = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
)

const (
	SUCCESS = iota
	HEADER_NOT_EXIST
	PROOF_FORMAT_ERROR
	VERIFY_PROOT_ERROR
	TX_HAS_COMMIT
	TRANSTION_NOT_CONFIRMED
	UNKNOWN
)

const (
	// PixieChain Testnet chainId
	pixieTestnetChainID uint64 = 666

	// PixieChain Testnet RPC
	pixieTestnetRPC = "https://http-testnet.chain.pixie.xyz"

	BlocksToWait           = 21
	CrossChainDataContract = "" // TODO: to be filled
)

const (
	CrossChainTxHeight = 1 // TODO: to be filled
	KeyIndex           = 2
	RawValue           = "" // TODO: to be filled
	testKey            = "" // TODO: to be filled
	testBlockHeight    = "" // TODO: to be filled
	testProofStr       = "" // TODO: to be filled
)

func init() {
	setBKers()
}

func typeOfError(e error) int {
	if e == nil {
		return SUCCESS
	}
	errDesc := e.Error()
	if strings.Contains(errDesc, "GetHeaderByHeight, height is too big") {
		return HEADER_NOT_EXIST
	} else if strings.Contains(errDesc, "unmarshal proof error:") {
		return PROOF_FORMAT_ERROR
	} else if strings.Contains(errDesc, "verify proof value hash failed") {
		return VERIFY_PROOT_ERROR
	} else if strings.Contains(errDesc, "check done transaction error:checkDoneTx, tx already done") {
		return TX_HAS_COMMIT
	} else if strings.Contains(errDesc, "transaction is not confirmed") {
		return TRANSTION_NOT_CONFIRMED
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
	ns, err := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		panic(fmt.Sprintf("NewNativeService error: %+v", err))
	}

	// add sidechain info
	extra := pixiechain.ExtraInfo{
		ChainID: big.NewInt(int64(pixieTestnetChainID)),
		Period:  3,
	}
	extraBytes, _ := json.Marshal(extra)
	contaractAddr, _ := hex.DecodeString(ccmcom.Replace0x(CrossChainDataContract))
	side := &side_chain_manager.SideChain{
		Name:         "PixieChain",
		ChainId:      pixieTestnetChainID,
		BlocksToWait: BlocksToWait,
		Router:       utils.PIXIECHAIN_ROUTER,
		CCMCAddress:  contaractAddr,
		ExtraInfo:    extraBytes,
	}
	sideInfo, err := ns.GetCacheDB().Get(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(side.ChainId)))
	if err != nil {
		panic(fmt.Sprintf("NewNativeService GetSideChainInfo error: %+v", err))
	}
	if sideInfo == nil {
		sink := common.NewZeroCopySink(nil)
		_ = side.Serialization(sink)
		ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(side.ChainId)), cstates.GenRawStorageItem(sink.Bytes()))
	}
	return ns
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

func (client *PixieChainClient) GetProofFromAchieveNode(contractAddress string, key string, blockheight string) ([]byte, error) {
	req := &proofReq{
		JsonRPC: "2.0",
		Method:  "eth_getProof",
		Params:  []interface{}{contractAddress, []string{key}, blockheight},
		Id:      1,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("get_ethproof: marshal req err: %s", err)
	}

	fmt.Printf("req is %s\n", data)
	resp, err := client.SendRestRequest(data)
	if err != nil {
		return nil, fmt.Errorf("GetProof: send request err: %s", err)
	}
	proofRep := &ProofRep{}
	err = json.Unmarshal(resp, proofRep)
	if err != nil {
		return nil, fmt.Errorf("GetProof, unmarshal resp err: %s", err)
	}

	fmt.Printf("proof res is:%v\n", proofRep)

	result, err := json.Marshal(proofRep.Result)
	if err != nil {
		return nil, fmt.Errorf("GetProof, Marshal result err: %s", err)
	}
	proof := new(Proof)
	if err = json.Unmarshal(result, proof); err != nil {
		fmt.Printf("json.Unmarshal result to Proof struct err: %v", err)
	}
	return result, nil
}

func (client *PixieChainClient) GetProof(contractAddress string, key string, blockheight string) ([]byte, error) {
	result := make([]byte, 0)

	if contractAddress == CrossChainDataContract &&
		key == testKey &&
		blockheight == testBlockHeight {
		result = []byte(testProofStr)
	}

	return result, nil
}

func getGenesisHeader(t *testing.T) *pixiechain.GenesisHeader {
	c := newPixieChainClient()
	height, err := c.GetNodeHeight()
	if err != nil {
		panic(err)
	}

	height = uint64(CrossChainTxHeight)

	var backOffHeight uint64 = 0

	epochHeight := height - height%200 - backOffHeight
	pEpochHeight := epochHeight - 200 - backOffHeight

	hdr, err := c.GetBlockHeader(epochHeight)
	assert.NilError(t, err)
	phdr, err := c.GetBlockHeader(pEpochHeight)
	assert.NilError(t, err)
	pvalidators, err := pixiechain.ParseValidators(phdr.Extra[32 : len(phdr.Extra)-65])
	assert.NilError(t, err)

	genesisHeader := pixiechain.GenesisHeader{Header: *hdr, PrevValidators: []pixiechain.HeightAndValidators{
		{Height: big.NewInt(int64(pEpochHeight)), Validators: pvalidators},
	}}

	return &genesisHeader
}

func getLatestHeight(native *native.NativeService) (height uint64) {
	heightStore, err := native.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(synccom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(pixieTestnetChainID)))
	if err != nil {
		err = fmt.Errorf("hgetLatestHeight err:%v", err)
		return
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		err = fmt.Errorf("getLatestHeight, GetValueFromRawStorageItem err:%v", err)
		return
	}
	height = utils.GetBytesUint64(storeBytes)
	return
}

func untilGetBlockHeader(t *testing.T, c *PixieChainClient, height uint64) *eth.Header {
	for {
		hdr, err := c.GetBlockHeader(height)
		if err == nil {
			return hdr
		}
		// time.Sleep(time.Second)
		t.Error("GetBlockHeader", height, err)
	}
}

func TestProofHandle(t *testing.T) {
	var (
		client      *PixieChainClient
		testNative  *native.NativeService
		syncHandler *pixiechain.Handler
		handler     *PixieHandler
		height      uint64
	)

	syncHandler = pixiechain.NewPixieHandler()
	handler = NewPixieHandler()
	{
		genesisHeader := getGenesisHeader(t)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(synccom.SyncGenesisHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acct.Address},
		}
		testNative = NewNative(sink.Bytes(), tx, nil)
		err := syncHandler.SyncGenesisHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err))
		height := getLatestHeight(testNative)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)

	}

	{
		client = newPixieChainClient()
		param := new(synccom.SyncBlockHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.Address = acct.Address
		height = getLatestHeight(testNative)

		var headerNumber = uint64(CrossChainTxHeight) + 2*BlocksToWait - 2 - height
		for i := 1; i <= int(headerNumber); i++ {
			headerBs, _ := json.Marshal(untilGetBlockHeader(t, client, height+uint64(i)))
			param.Headers = append(param.Headers, headerBs)
		}

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		testNative = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())

		err := syncHandler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(testNative)
		assert.Equal(t, latestHeight, height+headerNumber)
	}

	var dparam *ccmcom.EntranceParam
	{
		key := hex.EncodeToString(big.NewInt(int64(KeyIndex)).Bytes())
		proofHeight := uint32(CrossChainTxHeight + BlocksToWait)
		proofHeightHex := hexutil.EncodeBig(big.NewInt(int64(proofHeight)))
		keyBytes, err := eth2.MappingKeyAt(key, "01")
		if err != nil {
			fmt.Printf("handleLockDepositEvents - MappingKeyAt error:%s\n", err.Error())
			return
		}
		proofKey := hexutil.Encode(keyBytes)
		proof, err := client.GetProof(CrossChainDataContract, proofKey, proofHeightHex)
		// proof, err := client.GetProofFromAchieveNode(CrossChainDataContract, proofKey, proofHeightHex)

		if err != nil {
			fmt.Printf("GetProof, err: %v", err)
			return
		}
		fmt.Printf("proof is %x\n", proof)

		value, _ := hex.DecodeString(RawValue)

		dparam = new(ccmcom.EntranceParam)
		dparam.SourceChainID = pixieTestnetChainID
		dparam.Height = proofHeight
		dparam.Proof = proof
		dparam.RelayerAddress = acct.Address[:]
		dparam.Extra = value
		dparam.HeaderOrCrossChainMsg = []byte{}
		sink := common.NewZeroCopySink(nil)
		dparam.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		testNative = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())
		_, err = handler.MakeDepositProposal(testNative)
		assert.Equal(t, TRANSTION_NOT_CONFIRMED, typeOfError(err))
	}

	{
		param := new(synccom.SyncBlockHeaderParam)
		param.ChainID = pixieTestnetChainID
		param.Address = acct.Address
		height = getLatestHeight(testNative)
		var headerNumber = uint64(CrossChainTxHeight) + 2*BlocksToWait - 1 - height
		for i := 1; i <= int(headerNumber); i++ {
			headerBs, _ := json.Marshal(untilGetBlockHeader(t, client, height+uint64(i)))
			param.Headers = append(param.Headers, headerBs)
		}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		testNative = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())
		err := syncHandler.SyncBlockHeader(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err))

		sink = common.NewZeroCopySink(nil)
		dparam.Serialization(sink)
		tx = &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		testNative = NewNative(sink.Bytes(), tx, testNative.GetCacheDB())
		_, err = handler.MakeDepositProposal(testNative)
		assert.Equal(t, SUCCESS, typeOfError(err))

		// cross chain tx happened at height : CrossChainTxHeight
		// relayers try to submit tx proof generating from height : CrossChainTxHeight + BlocksToWait
		assert.Equal(t, dparam.Height, uint32(CrossChainTxHeight+BlocksToWait))
		// poly chain says: now I trust the header at height: dParam.Height when poly chain gets header at height: dParam.Height + BlocksToWait - 1
		height = getLatestHeight(testNative)
		assert.Equal(t, height, uint64(dparam.Height+BlocksToWait-1))
	}
}
