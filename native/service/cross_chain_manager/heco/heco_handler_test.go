package heco

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	etypes "github.com/ethereum/go-ethereum/core/types"
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
	"github.com/polynetwork/poly/native/service/cross_chain_manager/eth"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	synccom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/heco"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"gotest.tools/assert"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

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
	HecoChainID    uint64 = 7
	HecoTestnetRpc        = "https://http-testnet.hecochain.com"
	BlocksToWait          = 21
	HecoCCDC              = "0x7b78A940b5C9A5186035543893e377EbEEa1EDfD" // huobi eco chain ccdc (cross chain data contract)
)

var (
	CrossChainTxHeight = 808477
	KeyIndex           = 2
	RawValue           = "20000000000000000000000000000000000000000000000000000000000000000220fc1afb9c106376dae4a2b04b6532f5cec0eb27df04352b64c7cd24f8f7e6aadd14820a47272484d11e9ddcfcfe4a94d3ccdb37f394070000000000000014820a47272484d11e9ddcfcfe4a94d3ccdb37f39406756e6c6f636b4a14000000000000000000000000000000000000000014344cfc3b8635f72f14200aaf2168d9f75df86fd30200000000000000000000000000000000000000000000000000000000000000"
)

func init() {
	setBKers()
	//client := newHecoClient()

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
	extra := heco.ExtraInfo{
		// test id 256, main id 128
		ChainID: big.NewInt(128),
		Period:  3,
	}
	extraBytes, _ := json.Marshal(extra)
	contaractAddr, _ := hex.DecodeString(ccmcom.Replace0x(HecoCCDC))
	side := &side_chain_manager.SideChain{
		Name:         "heco",
		ChainId:      HecoChainID,
		BlocksToWait: BlocksToWait,
		Router:       utils.HECO_ROUTER,
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

type HecoClient struct {
	addr       string
	restClient *http.Client
}

func newHecoClient() *HecoClient {

	c := &HecoClient{
		addr: HecoTestnetRpc,
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

func (this *HecoClient) SendRestRequest(data []byte) ([]byte, error) {
	resp, err := this.restClient.Post(this.addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("http post request:%s error:%s", data, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rest response body error:%s", err)
	}
	return body, nil
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

func (this *HecoClient) GetNodeHeight() (uint64, error) {
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
	resp, err := this.SendRestRequest(data)
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

type BlockReq struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint          `json:"id"`
}

type BlockRep struct {
	JsonRPC string         `json:"jsonrpc"`
	Result  *etypes.Header `json:"result"`
	Id      uint           `json:"id"`
}

func (this *HecoClient) GetBlockHeader(height uint64) (*etypes.Header, error) {
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
	resp, err := this.SendRestRequest(data)
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

func (this *HecoClient) GetProofFromAchieveNode(contractAddress string, key string, blockheight string) ([]byte, error) {
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
	resp, err := this.SendRestRequest(data)
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
func (this *HecoClient) GetProof(contractAddress string, key string, blockheight string) ([]byte, error) {
	result := make([]byte, 0)

	proofStr1 := "{\"address\":\"0x7b78a940b5c9a5186035543893e377ebeea1edfd\",\"accountProof\":[\"0xf90211a000444dea13727b76365e8790efb17c1adc6751ad1e5124d84bd459e7f6e52744a02899c51278969ee83dbf07cb8a712906e14e43967894b623557546160108c474a06aaf0389f0bafb9638254a0955e498ea9840df0dfb258e83b71ddc007efd70f3a02e950a8a1c1af6bb3b48cef39a1ac1562aeab722cb0d1f0cc81d2198e22a32b8a02796ce8bb429ad3f61b9f53ff01c135c17c53d02145a9f76b532821f4f31dcfca0e055dbde457259b35fdb1e1637982199fabecf5a10aa0f6736e6c310026f1344a038a1e050ab6cd8e1c4a1ed3184c1e803e1687bfaec941b59cd6ab9ff064abaf4a029dc0e7acc8443176da178de13768fae642de479ebe0720236c7a07ec0dd4f1ca0beb36ef202ffc9803df7d1593243e2d309dedaf2323c7f455da1c5bc4df5b5e6a0af208966ddf600a501d03a5e89a639e823f6a3f2c9e86427c6dc0e327561aeeda0ae781c50f6ffd9ae890d1b8320fe34a54530f6cceee36530da65c4612671fee5a0c1e81a20bd97a03bd4fc67ecb3d7ce53652c3c83ec4fd7d305ff9ebfd33360d4a0b86bceabd0fd2549257a18702d72a17041092c43944ebd0caf0e14dfb6ad8737a0bf2a4e9e7601b29b223b073ed1f8f7320ac2ad7555907b5eef29d4125aba27bda000ffa8fa428969c6dfbe36c3908d128545dee3422dd741a3b0e5f30efeee45dba0581dafb1f3f584e45c40e2c1291f56365f9d832e366c45448055a3662f62452480\",\"0xf90211a066a2304a0baf5c97fee499a3cb8a020d2564cbd1627c88dc69f89b4c9d8f4649a01fcb9acf9c281cafbca6d20d449323fc5a72a77701ea0c25804b14e39969496ca0a44d039b545990a4671cb06af21f28debb147965cec54246f8b9ca9872d199b0a00a3ca9d674ae3cd6ccd1ca8a985ad626343a58392104aaf617361c38efd932e2a0b942fd2b12f28900952ba2ac523725fbe1cbf2ed34e286fc2c2c4206f82baaa7a0fa9ce14e844e8f211eba9bb4e017678c89527b4a263e1904159f149d520f1a25a0286110013f12683a8247959559cd2256b111b6c641e9bb8e10e29a8c88ac9642a0dd5013060aa18beb8490607509609c8618e5b5d41f1b48cee3ed2b1f4c12a1e6a02c9ccf387f01605f2d58a97ea62961d8335b2e66823046a2ffb40095f6012e24a0bbe5f6933c3f3b3e12dd1952bce2cfae77654fdafefeb733f189aa143ebb7dcca0520f3f2bcf428b33d287f3ad0d86d58adfd70261a8b5b75ec0821a2a60858ec4a0fc0863923da4f1739ca471b8af2ad4f9806e9fb0763736e667fdb2b2cb24b63da088cbd7be60d0b6f3d497623e13c5c3c45886af49da5bbed6d9d7a6f0d3209073a00aa6fcbce7f58934c6bb89cd3213f83ae8404e34a32a5efbd6c1a3681c400e7ea0f49f230b4f4f3253a076bffb5ea73f7c46b8a3e9baa4a5d8f31796f05dcafea9a0e783e25cf8b98d6b7614fa7724d6e125e3c187808aeb4263a6b6522c2be2469580\",\"0xf901d1a07e7f80f16faffc32f24871f6a9dac35cc1646d2eab3e8b15318492d6b49126e5a0e3e2241422def5f6b139f6cf02c75cec74d90922d574690537a7cbe341e25f58a0d5c6ba154b89909b0a3758baeeb431557cca9d68c9d47166117a12a2cb32ba9e80a0c0db04e3285d049fc073e571ead5133c2087d1f3468ffe49644912231ed6b8cba09e77ddccc9ee0a49a136b90af1c0c74783484ce186e5d56073316c95e0ad663ea03ac2e912eaf92b22810d5d992746f87db6af483125372a8c7dcacb3c6d33da8ba07a8d168f283636d738f2292ee948466a1332fd2d4472ead373e97a6e4185ca41a02cadc5523cbfd765aa3507cd415bba7dc6bb45b89b8e3806ae14c6c9ae5fc39ba015a6c6e205c3ce8ec362638577718079d522dc248496eefe79a51d263e963fdba024499cf0ccb872f0fafdcab09fd324a195d947ec242e45cbe82b5cf763739a5e80a0b732d276c2c01d2b00bc1993dfa0b58480e7d917b40fc3f42bd0cd92007204b3a0c17f4ebdca55bfbbdddaf0dfa221953c7f26f7a514c22ca54fc9067510aff75ba0511ebb41d9f85f9006c3b3963676b2a20543a36b08c89467a4ba475cf73dfcdea03a3faf64ef0901d74a836857eb872fbc1d613e32efc5647854d7b31803428fe080\",\"0xf8689f306e84ac0b76baa156eeb0bba73ce3504729c5930077fdd0fac31eca1ccec2b846f8440180a034884751b402e8c27fc712eccdb508ce45e8bd4e66541e39fb74abaae93f1e61a045b501b894b983d55f8ecf35d15c48ea02031f475cdaf7cfb7727bf017f649a6\"],\"balance\":\"0x0\",\"codeHash\":\"0x45b501b894b983d55f8ecf35d15c48ea02031f475cdaf7cfb7727bf017f649a6\",\"nonce\":\"0x1\",\"storageHash\":\"0x34884751b402e8c27fc712eccdb508ce45e8bd4e66541e39fb74abaae93f1e61\",\"storageProof\":[{\"key\":\"0xd9d16d34ffb15ba3a3d852f0d403e2ce1d691fb54de27ac87cd2f993f3ec330f\",\"value\":\"0x449636f31269f9b6a2120b0ff907a50bf0a198be0fe8df02f2783dfbcadf28ff\",\"proof\":[\"0xf8b1a036ea3f3f2f92ebb1209c2fa558a58a2925f6c01505954d653e14e54d6a7537b880a065b408c98535c2eaedec14bf7e0821b07e7d0450ef9e306344962ead3cc85eaf80a0cd457259696115235e64c7822334d62129e2f1604425a7da6494f35fc45be51880a0d10b44f162456327f06a26d293f4351d88d7b350c159b3c8dc219aba7561e5238080a091b50d8ba69c6c1faa1fc369a05dfd85a426e65fa013a63e1a7fc8c972ba913e80808080808080\",\"0xf843a03feccf6caa602894c8105bdda7f81b2a7bb7de7dba1f18af92d8d057b708cb41a1a0449636f31269f9b6a2120b0ff907a50bf0a198be0fe8df02f2783dfbcadf28ff\"]}]}"
	if contractAddress == "0x7b78A940b5C9A5186035543893e377EbEEa1EDfD" &&
		key == "0xd9d16d34ffb15ba3a3d852f0d403e2ce1d691fb54de27ac87cd2f993f3ec330f" &&
		blockheight == "0xc5632" {
		result = []byte(proofStr1)
	}

	return result, nil
}

func getGenesisHeader(t *testing.T) *heco.GenesisHeader {
	c := newHecoClient()
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
	pvalidators, err := heco.ParseValidators(phdr.Extra[32 : len(phdr.Extra)-65])
	assert.NilError(t, err)

	genesisHeader := heco.GenesisHeader{Header: *hdr, PrevValidators: []heco.HeightAndValidators{
		{Height: big.NewInt(int64(pEpochHeight)), Validators: pvalidators},
	}}

	return &genesisHeader
}

func getLatestHeight(native *native.NativeService) (height uint64) {
	heightStore, err := native.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(synccom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(HecoChainID)))
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

func untilGetBlockHeader(t *testing.T, c *HecoClient, height uint64) *etypes.Header {
	for {
		hdr, err := c.GetBlockHeader(height)
		if err == nil {
			return hdr
		}
		//time.Sleep(time.Second)
		fmt.Println("GetBlockHeader", height, err)
	}
}
func TestProofHandle(t *testing.T) {
	var (
		client      *HecoClient
		native      *native.NativeService
		syncHandler *heco.Handler
		handler     *HecoHandler
		height      uint64
	)
	syncHandler = heco.NewHecoHandler()
	handler = NewHecoHandler()
	{
		genesisHeader := getGenesisHeader(t)
		genesisHeaderBytes, _ := json.Marshal(genesisHeader)

		param := new(synccom.SyncGenesisHeaderParam)
		param.ChainID = HecoChainID
		param.GenesisHeader = genesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acct.Address},
		}
		native = NewNative(sink.Bytes(), tx, nil)
		err := syncHandler.SyncGenesisHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err))
		height := getLatestHeight(native)
		assert.Equal(t, genesisHeader.Header.Number.Uint64(), height)

	}
	{
		client = newHecoClient()
		param := new(synccom.SyncBlockHeaderParam)
		param.ChainID = HecoChainID
		param.Address = acct.Address
		height = getLatestHeight(native)

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
		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())

		err := syncHandler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)
		latestHeight := getLatestHeight(native)
		assert.Equal(t, latestHeight, height+headerNumber)
	}

	var dparam *ccmcom.EntranceParam
	{

		key := hex.EncodeToString(big.NewInt(int64(KeyIndex)).Bytes())
		proofHeight := uint32(CrossChainTxHeight + BlocksToWait)
		proofHeightHex := hexutil.EncodeBig(big.NewInt(int64(proofHeight)))
		keyBytes, err := eth.MappingKeyAt(key, "01")
		if err != nil {
			fmt.Printf("handleLockDepositEvents - MappingKeyAt error:%s\n", err.Error())
			return
		}
		proofKey := hexutil.Encode(keyBytes)
		proof, err := client.GetProof(HecoCCDC, proofKey, proofHeightHex)
		//proof, err := client.GetProofFromAchieveNode(HecoCCDC, proofKey, proofHeightHex)

		if err != nil {
			fmt.Printf("GetProof, err: %v", err)
			return
		}
		fmt.Printf("proof is %x\n", proof)

		value, _ := hex.DecodeString(RawValue)

		dparam = new(ccmcom.EntranceParam)
		dparam.SourceChainID = HecoChainID
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
		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		_, err = handler.MakeDepositProposal(native)
		assert.Equal(t, TRANSTION_NOT_CONFIRMED, typeOfError(err))

	}

	{
		param := new(synccom.SyncBlockHeaderParam)
		param.ChainID = HecoChainID
		param.Address = acct.Address
		height = getLatestHeight(native)
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
		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		err := syncHandler.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err))

		sink = common.NewZeroCopySink(nil)
		dparam.Serialization(sink)
		tx = &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		_, err = handler.MakeDepositProposal(native)
		assert.Equal(t, SUCCESS, typeOfError(err))

		// cross chain tx happened at height : CrossChainTxHeight
		// relayers try to submit tx proof generating from height : CrossChainTxHeight + BlocksToWait
		assert.Equal(t, dparam.Height, uint32(CrossChainTxHeight+BlocksToWait))
		// poly chain says: now I trust the header at height: dParam.Height when poly chain gets header at height: dParam.Height + BlocksToWait - 1
		height = getLatestHeight(native)
		assert.Equal(t, height, uint64(dparam.Height+BlocksToWait-1))
	}
}
