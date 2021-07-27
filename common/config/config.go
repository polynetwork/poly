/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package config

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/constants"
	"github.com/polynetwork/poly/common/log"
)

var Version = "" //Set value when build project

const (
	DEFAULT_CONFIG_FILE_NAME = "./config.json"
	DEFAULT_WALLET_FILE_NAME = "./wallet.dat"
	MIN_GEN_BLOCK_TIME       = 2
	DEFAULT_GEN_BLOCK_TIME   = 6
	DBFT_MIN_NODE_NUM        = 4 //min node number of dbft consensus
	SOLO_MIN_NODE_NUM        = 1 //min node number of solo consensus
	VBFT_MIN_NODE_NUM        = 4 //min node number of vbft consensus

	CONSENSUS_TYPE_DBFT = "dbft"
	CONSENSUS_TYPE_SOLO = "solo"
	CONSENSUS_TYPE_VBFT = "vbft"

	DEFAULT_LOG_LEVEL                       = log.InfoLog
	DEFAULT_MAX_LOG_SIZE                    = 100 //MByte
	DEFAULT_NODE_PORT                       = uint(20338)
	DEFAULT_CONSENSUS_PORT                  = uint(20339)
	DEFAULT_RPC_PORT                        = uint(20336)
	DEFAULT_RPC_LOCAL_PORT                  = uint(20337)
	DEFAULT_REST_PORT                       = uint(20334)
	DEFAULT_WS_PORT                         = uint(20335)
	DEFAULT_REST_MAX_CONN                   = uint(1024)
	DEFAULT_MAX_CONN_IN_BOUND               = uint(1024)
	DEFAULT_MAX_CONN_OUT_BOUND              = uint(1024)
	DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP = uint(16)
	DEFAULT_HTTP_INFO_PORT                  = uint(0)
	DEFAULT_MAX_TX_IN_BLOCK                 = 60000
	DEFAULT_MAX_SYNC_HEADER                 = 500
	DEFAULT_ENABLE_CONSENSUS                = true
	DEFAULT_ENABLE_EVENT_LOG                = true
	DEFAULT_CLI_RPC_PORT                    = uint(20000)
	DEFUALT_CLI_RPC_ADDRESS                 = "127.0.0.1"
	DEFAULT_GAS_LIMIT                       = 20000
	DEFAULT_GAS_PRICE                       = 500

	DEFAULT_DATA_DIR      = "./Chain"
	DEFAULT_RESERVED_FILE = "./peers.rsv"
)

const (
	NETWORK_ID_MAIN_NET   = 1
	NETWORK_ID_TEST_NET   = 2
	NETWORK_ID_SOLO_NET   = 3
	NETWORK_NAME_MAIN_NET = "main"
	NETWORK_NAME_TEST_NET = "test"
	NETWORK_NAME_SOLO_NET = "testmode"
	MAINNET_CHAIN_ID      = 0
	TESTNET_CHAIN_ID      = common.MAX_INT64
)

var NETWORK_MAGIC = map[uint32]uint32{
	NETWORK_ID_MAIN_NET: constants.NETWORK_MAGIC_MAINNET, //Network main
	NETWORK_ID_TEST_NET: constants.NETWORK_MAGIC_TESTNET, //Network test
	NETWORK_ID_SOLO_NET: 0,                               //Network solo
}

var NETWORK_NAME = map[uint32]string{
	NETWORK_ID_MAIN_NET: NETWORK_NAME_MAIN_NET,
	NETWORK_ID_TEST_NET: NETWORK_NAME_TEST_NET,
	NETWORK_ID_SOLO_NET: NETWORK_NAME_SOLO_NET,
}
var CHAIN_ID = map[uint32]uint64{
	NETWORK_ID_MAIN_NET: MAINNET_CHAIN_ID,
	NETWORK_ID_TEST_NET: TESTNET_CHAIN_ID,
}

var EXTRA_INFO_HEIGHT = map[uint32]uint32{
	NETWORK_ID_MAIN_NET: constants.EXTRA_INFO_HEIGHT_MAINNET,
	NETWORK_ID_TEST_NET: constants.EXTRA_INFO_HEIGHT_TESTNET,
}

var ETH1559_HEIGHT = map[uint32]uint64{
	NETWORK_ID_MAIN_NET: constants.ETH1559_HEIGHT_MAINNET,
	NETWORK_ID_TEST_NET: constants.ETH1559_HEIGHT_TESTNET,
}

var POLYGON_SNAP_CHAINID = map[uint32]uint32{
	NETWORK_ID_MAIN_NET: constants.POLYGON_SNAP_CHAINID_MAINNET,
}

var (
	EXTRA_INFO_HEIGHT_FORK_CHECK bool
)

func GetNetworkMagic(id uint32) uint32 {
	nid, ok := NETWORK_MAGIC[id]
	if ok {
		return nid
	}
	return id
}

func GetPolygonSnapChainID(id uint32) uint32 {
	height := POLYGON_SNAP_CHAINID[id]
	return height
}

func GetEth1559Height(id uint32) uint64 {
	height := ETH1559_HEIGHT[id]
	if height == 0 {
		height = constants.ETH1559_HEIGHT_TESTNET
	}
	return height
}

func GetExtraInfoHeight(id uint32) uint32 {
	return EXTRA_INFO_HEIGHT[id]
}

func GetNetworkName(id uint32) string {
	name, ok := NETWORK_NAME[id]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", id)
}

func GetChainIdByNetId(id uint32) uint64 {
	chainId, ok := CHAIN_ID[id]
	if ok {
		return chainId
	}
	return uint64(id)
}

var PolarisConfig = &GenesisConfig{
	SeedList: []string{
		"beta1.poly.network:20338",
		"beta2.poly.network:20338",
		"beta3.poly.network:20338",
		"beta4.poly.network:20338",
		"beta5.poly.network:20338",
		"beta6.poly.network:20338",
		"beta7.poly.network:20338"},
	ConsensusType: CONSENSUS_TYPE_VBFT,
	VBFT: &VBFTConfig{
		BlockMsgDelay:        10000,
		HashMsgDelay:         10000,
		PeerHandshakeTimeout: 10,
		MaxBlockChangeView:   60000,
		VrfValue:             "1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
		VrfProof:             "c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
		Peers: []*VBFTPeerInfo{
			{
				Index:      1,
				PeerPubkey: "120503ef44beba84422bd76a599531c9fe50969a929a0fee35df66690f370ce19fa8c0",
				Address:    "ATkhypXmvPNSeX64ECyL6kT8zSruDBLtBJ",
			},
			{
				Index:      2,
				PeerPubkey: "1205038247efcfeae0fdf760685d1ac1c083be3ff5e9a4a548bc3a2e98f0434f092483",
				Address:    "AVrKPxydgNjtLEQxT2krpBkksBWZMq5Sv3",
			},
			{
				Index:      3,
				PeerPubkey: "1205022092e34e0176dccf8abb496b833d591d25533469b3caf0e279b9742955dd8fc3",
				Address:    "ARMkJPzh7dR4p6Wi44T55dJ4sjr6C5hBUg",
			},
			{
				Index:      4,
				PeerPubkey: "1205027bd771e68adb88398282e21a8b03c12f64c2351ea49a2ba06a0327c83b239ca9",
				Address:    "AScEX1ibeBZtGxkpujLKEB3T1yW6ufUj5C",
			},
			{
				Index:      5,
				PeerPubkey: "120502d0d0e883c73d8256cf4314822ddd973c0179b73d8ed3df85aad38d36a8b2b0c7",
				Address:    "AKPQsGZG5zyRouHT6MAWbvYpoy8nzBFwjV",
			},
			{
				Index:      6,
				PeerPubkey: "120503a4f44dd65cbcc52b1d1ac51747378a7f84753b5f7bf2760ca21390ced6b172bb",
				Address:    "ASR4S4eBWpftG4tVv1Ku9YWJubzFuw66iL",
			},
			{
				Index:      7,
				PeerPubkey: "120502696c0cbe74f01ee85e3c0ebe4ebdc5bea404f199d0262f1941fd39ff0d100257",
				Address:    "AdUy5qgWTkSf6x4kiWf1fZ9pw8cuvCqoaf",
			},
		},
	},
	DBFT: &DBFTConfig{},
	SOLO: &SOLOConfig{},
}

var MainNetConfig = &GenesisConfig{
	SeedList: []string{
		"seed.poly.network:20338",
		"poly.ont.io:20338",
		"poly.ngd.network:20338",
		"poly-1.switcheo.network:20338"},
	ConsensusType: CONSENSUS_TYPE_VBFT,
	VBFT: &VBFTConfig{
		BlockMsgDelay:        10000,
		HashMsgDelay:         10000,
		PeerHandshakeTimeout: 10,
		MaxBlockChangeView:   60000,
		VrfValue:             "1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
		VrfProof:             "c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
		Peers: []*VBFTPeerInfo{
			{
				Index:      1,
				PeerPubkey: "12050309c6475ce07577ab72a1f96c263e5030cb53a843b00ca1238a093d9dcb183e2f",
				Address:    "ARJrYgcF36uhE2RwDrxucXMrCWuTWWd77x",
			},
			{
				Index:      2,
				PeerPubkey: "1205032bed55e8c4d9cbc50657ff5909ee51dc394a92aad911c36bace83c4d63540794",
				Address:    "AeMwacPLLSSuUyDcwiqFMLRhbmmWmYQxn1",
			},
			{
				Index:      3,
				PeerPubkey: "120502e68a6e54bdfa0af47bd18465f4352f5151dc729c61a7399909f1cd1c6d816c02",
				Address:    "APDrR6TBQ6bN6AL9QVdXadNo9inqhoy3sp",
			},
			{
				Index:      4,
				PeerPubkey: "12050229e0d1c5b2ae838930ae1ad861ddd3d0745d1c7f142492cabd02b291d2c95c1d",
				Address:    "AZc4tbS4Q2p74jxMYqFGXB6Ky4Ap7gxE7Z",
			},
		},
	},
	DBFT: &DBFTConfig{},
	SOLO: &SOLOConfig{},
}

var DefConfig = NewOntologyConfig()

type GenesisConfig struct {
	SeedList      []string
	ConsensusType string
	VBFT          *VBFTConfig
	DBFT          *DBFTConfig
	SOLO          *SOLOConfig
}

func NewGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		SeedList:      make([]string, 0),
		ConsensusType: CONSENSUS_TYPE_DBFT,
		VBFT:          &VBFTConfig{},
		DBFT:          &DBFTConfig{},
		SOLO:          &SOLOConfig{},
	}
}

//
// VBFT genesis config, from local config file
//
type VBFTConfig struct {
	BlockMsgDelay        uint32          `json:"block_msg_delay"`
	HashMsgDelay         uint32          `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32          `json:"peer_handshake_timeout"`
	MaxBlockChangeView   uint32          `json:"max_block_change_view"`
	VrfValue             string          `json:"vrf_value"`
	VrfProof             string          `json:"vrf_proof"`
	Peers                []*VBFTPeerInfo `json:"peers"`
}

func (self *VBFTConfig) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(self.BlockMsgDelay)
	sink.WriteUint32(self.HashMsgDelay)
	sink.WriteUint32(self.PeerHandshakeTimeout)
	sink.WriteUint32(self.MaxBlockChangeView)
	sink.WriteString(self.VrfValue)
	sink.WriteString(self.VrfProof)
	sink.WriteVarUint(uint64(len(self.Peers)))
	for _, peer := range self.Peers {
		if err := peer.Serialization(sink); err != nil {
			return err
		}
	}
	return nil
}

func (this *VBFTConfig) Deserialization(source *common.ZeroCopySource) error {
	blockMsgDelay, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize blockMsgDelay error!")
	}
	hashMsgDelay, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize hashMsgDelay error!")
	}
	peerHandshakeTimeout, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize peerHandshakeTimeout error!")
	}
	maxBlockChangeView, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize maxBlockChangeView error!")
	}
	vrfValue, eof := source.NextString()
	if eof {
		return fmt.Errorf("serialization.ReadString, deserialize vrfValue error!")
	}
	vrfProof, eof := source.NextString()
	if eof {
		return fmt.Errorf("serialization.ReadString, deserialize vrfProof error!")
	}
	length, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peer length error!")
	}
	peers := make([]*VBFTPeerInfo, 0)
	for i := 0; uint64(i) < length; i++ {
		peer := new(VBFTPeerInfo)
		err := peer.Deserialization(source)
		if err != nil {
			return fmt.Errorf("deserialize peer error, error:%s", err)
		}
		peers = append(peers, peer)
	}
	this.BlockMsgDelay = blockMsgDelay
	this.HashMsgDelay = hashMsgDelay
	this.PeerHandshakeTimeout = peerHandshakeTimeout
	this.MaxBlockChangeView = maxBlockChangeView
	this.VrfValue = vrfValue
	this.VrfProof = vrfProof
	this.Peers = peers
	return nil
}

type VBFTPeerInfo struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

func (this *VBFTPeerInfo) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(this.Index)
	sink.WriteString(this.PeerPubkey)

	address, err := common.AddressFromBase58(this.Address)
	if err != nil {
		return fmt.Errorf("serialize VBFTPeerStackInfo error: %v", err)
	}
	address.Serialization(sink)
	return nil
}

func (this *VBFTPeerInfo) Deserialization(source *common.ZeroCopySource) error {
	index, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize index error!")
	}
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err := address.Deserialization(source)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error!")
	}
	this.Index = index
	this.PeerPubkey = peerPubkey
	this.Address = address.ToBase58()
	return nil
}

type DBFTConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

type SOLOConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

type CommonConfig struct {
	LogLevel       uint
	NodeType       string
	EnableEventLog bool
	SystemFee      map[string]int64
	GasLimit       uint64
	GasPrice       uint64
	DataDir        string
}

type ConsensusConfig struct {
	EnableConsensus bool
	MaxTxInBlock    uint
}

type P2PRsvConfig struct {
	ReservedPeers []string `json:"reserved"`
	MaskPeers     []string `json:"mask"`
}

type P2PNodeConfig struct {
	ReservedPeersOnly         bool
	ReservedCfg               *P2PRsvConfig
	NetworkMagic              uint32
	NetworkId                 uint32
	NetworkName               string
	NodePort                  uint
	NodeConsensusPort         uint
	DualPortSupport           bool
	IsTLS                     bool
	CertPath                  string
	KeyPath                   string
	CAPath                    string
	HttpInfoPort              uint
	MaxHdrSyncReqs            uint
	MaxConnInBound            uint
	MaxConnOutBound           uint
	MaxConnInBoundForSingleIP uint
}

type RpcConfig struct {
	EnableHttpJsonRpc bool
	HttpJsonPort      uint
	HttpLocalPort     uint
}

type RestfulConfig struct {
	EnableHttpRestful  bool
	HttpRestPort       uint
	HttpMaxConnections uint
	HttpCertPath       string
	HttpKeyPath        string
}

type WebSocketConfig struct {
	EnableHttpWs bool
	HttpWsPort   uint
	HttpCertPath string
	HttpKeyPath  string
}

type OntologyConfig struct {
	Genesis   *GenesisConfig
	Common    *CommonConfig
	Consensus *ConsensusConfig
	P2PNode   *P2PNodeConfig
	Rpc       *RpcConfig
	Restful   *RestfulConfig
	Ws        *WebSocketConfig
}

func NewOntologyConfig() *OntologyConfig {
	return &OntologyConfig{
		Genesis: MainNetConfig,
		Common: &CommonConfig{
			LogLevel:       DEFAULT_LOG_LEVEL,
			EnableEventLog: DEFAULT_ENABLE_EVENT_LOG,
			SystemFee:      make(map[string]int64),
			GasLimit:       DEFAULT_GAS_LIMIT,
			DataDir:        DEFAULT_DATA_DIR,
		},
		Consensus: &ConsensusConfig{
			EnableConsensus: true,
			MaxTxInBlock:    DEFAULT_MAX_TX_IN_BLOCK,
		},
		P2PNode: &P2PNodeConfig{
			ReservedCfg:               &P2PRsvConfig{},
			ReservedPeersOnly:         false,
			NetworkId:                 NETWORK_ID_MAIN_NET,
			NetworkName:               GetNetworkName(NETWORK_ID_MAIN_NET),
			NetworkMagic:              GetNetworkMagic(NETWORK_ID_MAIN_NET),
			NodePort:                  DEFAULT_NODE_PORT,
			NodeConsensusPort:         DEFAULT_CONSENSUS_PORT,
			DualPortSupport:           true,
			IsTLS:                     false,
			CertPath:                  "",
			KeyPath:                   "",
			CAPath:                    "",
			HttpInfoPort:              DEFAULT_HTTP_INFO_PORT,
			MaxHdrSyncReqs:            DEFAULT_MAX_SYNC_HEADER,
			MaxConnInBound:            DEFAULT_MAX_CONN_IN_BOUND,
			MaxConnOutBound:           DEFAULT_MAX_CONN_OUT_BOUND,
			MaxConnInBoundForSingleIP: DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP,
		},
		Rpc: &RpcConfig{
			EnableHttpJsonRpc: true,
			HttpJsonPort:      DEFAULT_RPC_PORT,
			HttpLocalPort:     DEFAULT_RPC_LOCAL_PORT,
		},
		Restful: &RestfulConfig{
			EnableHttpRestful: true,
			HttpRestPort:      DEFAULT_REST_PORT,
		},
		Ws: &WebSocketConfig{
			EnableHttpWs: true,
			HttpWsPort:   DEFAULT_WS_PORT,
		},
	}
}

func (this *OntologyConfig) GetBookkeepers() ([]keypair.PublicKey, error) {
	var bookKeepers []string
	switch this.Genesis.ConsensusType {
	case CONSENSUS_TYPE_VBFT:
		for _, peer := range this.Genesis.VBFT.Peers {
			bookKeepers = append(bookKeepers, peer.PeerPubkey)
		}
	case CONSENSUS_TYPE_DBFT:
		bookKeepers = this.Genesis.DBFT.Bookkeepers
	case CONSENSUS_TYPE_SOLO:
		bookKeepers = this.Genesis.SOLO.Bookkeepers
	default:
		return nil, fmt.Errorf("Does not support %s consensus", this.Genesis.ConsensusType)
	}

	pubKeys := make([]keypair.PublicKey, 0, len(bookKeepers))
	for _, key := range bookKeepers {
		pubKey, err := hex.DecodeString(key)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return nil, fmt.Errorf("Incorrectly book keepers key:%s", err)
		}
		pubKeys = append(pubKeys, k)
	}
	keypair.SortPublicKeys(pubKeys)
	return pubKeys, nil
}

func (this *OntologyConfig) GetDefaultNetworkId() (uint32, error) {
	defaultNetworkId, err := this.getDefNetworkIDFromGenesisConfig(this.Genesis)
	if err != nil {
		return 0, err
	}
	mainNetId, err := this.getDefNetworkIDFromGenesisConfig(MainNetConfig)
	if err != nil {
		return 0, err
	}
	testnetId, err := this.getDefNetworkIDFromGenesisConfig(PolarisConfig)
	if err != nil {
		return 0, err
	}
	switch defaultNetworkId {
	case mainNetId:
		return NETWORK_ID_MAIN_NET, nil
	case testnetId:
		return NETWORK_ID_TEST_NET, nil
	}
	return defaultNetworkId, nil
}

func (this *OntologyConfig) getDefNetworkIDFromGenesisConfig(genCfg *GenesisConfig) (uint32, error) {
	var configData []byte
	var err error
	switch this.Genesis.ConsensusType {
	case CONSENSUS_TYPE_VBFT:
		configData, err = json.Marshal(genCfg.VBFT)
	case CONSENSUS_TYPE_DBFT:
		configData, err = json.Marshal(genCfg.DBFT)
	case CONSENSUS_TYPE_SOLO:
		return NETWORK_ID_SOLO_NET, nil
	default:
		return 0, fmt.Errorf("unknown consensus type:%s", this.Genesis.ConsensusType)
	}
	if err != nil {
		return 0, fmt.Errorf("json.Marshal error:%s", err)
	}
	data := sha256.Sum256(configData)
	return binary.LittleEndian.Uint32(data[0:4]), nil
}
