package zilliqa

import (
	"encoding/json"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/states"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/genesis"
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
	"math/big"
	"strings"
	"testing"
)

var (
	acct     = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
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

func init() {
	setBKers()
}

const ZILChainID = 9

func typeOfError(e error) int {
	if e == nil {
		return SUCCESS
	}
	errDesc := e.Error()
	if strings.Contains(errDesc, "ZILHandler GetHeaderByHeight, genesis header had been initialized") {
		return GENESIS_INITIALIZED
	} else if strings.Contains(errDesc, "ZILHandler SyncGenesisHeader, contract params deserialize error:") {
		return GENESIS_PARAM_ERROR
	} else if strings.Contains(errDesc, "ZILHandler SyncGenesisHeader: getGenesisHeader, deserialize header err:") {
		return SYNCBLOCK_PARAM_ERROR
	} else if strings.Contains(errDesc, "SyncBlockHeader, get the parent block failed. Error:") {
		return SYNCBLOCK_ORPHAN
	} else if strings.Contains(errDesc, "SyncBlockHeader, invalid difficulty:") {
		return DIFFICULTY_ERROR
	} else if strings.Contains(errDesc, "SyncBlockHeader, verify header error:") {
		return NONCE_ERROR
	} else if strings.Contains(errDesc, "SyncGenesisHeader, checkWitness error:") {
		return OPERATOR_ERROR
	}
	return UNKNOWN
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
	n, _ := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	// add sidechain info
	extra := ExtraInfo{
		NumOfGuardList: 9,
	}
	extraBytes, _ := json.Marshal(extra)
	side_chain_manager.PutSideChain(n, &side_chain_manager.SideChain{
		ExtraInfo: extraBytes,
		ChainId: ZILChainID,

	})

	return n, nil
}

func getLatestHeight(native *native.NativeService) uint64 {
	height, _ := GetCurrentTxHeaderHeight(native, ZILChainID)
	return height
}

func getHeaderHashByHeight(native *native.NativeService, height uint64) []byte {
	block, _ := GetTxHeaderByHeight(native, height, ZILChainID)
	return block.BlockHash[:]
}

func TestSyncGenesisHeader(t *testing.T) {
	txBlock1Raw := "{\"BlockHash\":[103,194,252,74,171,72,88,227,100,137,183,27,173,193,159,52,147,39,121,46,144,57,3,105,193,153,93,214,30,122,207,170],\"Cosigs\":{\"CS1\":{\"R\":19759839172862701417386858538844933317602802047121528082298854292139034496025,\"S\":83836253134355891252395628558534222866311033418039146262386912657823592661207},\"B1\":[true,false,true,true,true,true,true,true,false,false],\"CS2\":{\"R\":28578847155903696420642616141070583041033136144988503442597941154558833417823,\"S\":4792624902640882210219044128084442282742375830610744089567360957013725377897},\"B2\":[true,true,true,false,false,true,true,false,true,true]},\"Timestamp\":1613994961431313,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[170,87,100,204,22,70,8,92,109,210,4,43,183,132,237,106,78,21,77,34,240,209,209,218,205,217,232,102,98,96,30,7],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[72,83,6,125,117,117,81,199,240,17,151,134,204,131,7,112,165,246,196,222,176,217,10,11,109,164,156,173,73,150,102,19],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[58,90,228,245,209,61,41,241,185,218,246,128,235,27,165,171,166,53,143,51,255,105,39,205,15,237,149,248,239,53,35,211]},\"NumTxs\":0,\"MinerPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[223,20,253,250,182,22,129,200,91,138,135,112,116,82,36,240,95,245,145,241,232,32,6,112,229,97,78,46,112,105,247,150],\"Cosigs\":{\"CS1\":{\"R\":3916020268539532848467644216722860137824554821660678985761908595132973072714,\"S\":97045995147022036660902295039327684939913829468538947744163884834333849946847},\"B1\":[true,true,true,true,true,true,false,true,false,false],\"CS2\":{\"R\":92028961177392717940679859135217775710532087351598448092697854633151818843931,\"S\":111389800061060649664721171870250671692292738272837785727000701591328031299981},\"B2\":[true,false,true,true,true,true,true,true,false,false]},\"Timestamp\":1613994931556497,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[208,223,176,86,146,162,2,192,168,3,21,120,94,74,135,8,59,245,169,91,185,41,215,1,80,100,251,132,248,159,106,7],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C\":{\"IpAddress\":2052726836,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"/f1RFgKiztjTn2LiZ0T6t/494ctwMkaWBe0k+MGLuJE=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}}}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("2052726836", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57",
		},
		{
			PubKey: "0239D4CAE39A7AC2F285796BABF7D28DC8EB7767E78409C70926D0929EA2941E36",
		},
		{
			PubKey: "02D2D695D4A352412E0D32A8BDF6EA3A606D35FE2C2F850C54D68727D065894986",
		},
		{
			PubKey: "02E5E1BE6C924349F2C2B20CE05A2650B3E56C7722A2E5952EE27D12DEE7A4A6E6",
		},
		{
			PubKey: "0300AB86B413FAA64A52FB61B5A28A6C361F87A5B0871C4F01C394D261415B0989",
		},
		{
			PubKey: "03019AF5B10FFE09FB0EE02B59195EF5E6F5BE51D17EAF5604EA452078CD465C4B",
		},
		{
			PubKey: "0323086D473DF937B6297FB755FA8E57C0FB2760512AED7757748B597C48F797A0",
		},
		{
			PubKey: "032AEE20CFC59EAEB7838DAC2A9BAF96C8D69CF2C866FB4A3F1DFB02BCFCA356BB",
		},
		{
			PubKey: "033207325A3CC671034FEBA86EC8D0AA412DF60C7E8292044D510DF582787DCC05",
		},
		{
			PubKey: "03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C",
			Peer: core.Peer{
				IpAddress:      ipaddr,
				ListenPortHost: 33133,
			},
		},
	}
	txBlockAndDsComm := &TxBlockAndDsComm{
		TxBlock: &txblock,
		DsBlock: &dsblock,
		DsComm:  initComm,
	}
	txBlockAndDsCommRaw, _ := json.Marshal(txBlockAndDsComm)
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ZILChainID
	param.GenesisHeader = txBlockAndDsCommRaw
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	n, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	zilHeader := NewHandler()

	err1 := zilHeader.SyncGenesisHeader(n)
	assert.Equal(t, SUCCESS, typeOfError(err1))

	height := getLatestHeight(n)
	assert.Equal(t, uint64(1), height)

	headerHash := getHeaderHashByHeight(n, height)
	// [103,194,252,74,171,72,88,227,100,137,183,27,173,193,159,52,147,39,121,46,144,57,3,105,193,153,93,214,30,122,207,170]
	assert.Equal(t, "67c2fc4aab4858e36489b71badc19f349327792e90390369c1995dd61e7acfaa", util.EncodeHex(headerHash))
}

func TestSyncGenesisHeaderNoOperator(t *testing.T) {
	txBlock1Raw := "{\"BlockHash\":[103,194,252,74,171,72,88,227,100,137,183,27,173,193,159,52,147,39,121,46,144,57,3,105,193,153,93,214,30,122,207,170],\"Cosigs\":{\"CS1\":{\"R\":19759839172862701417386858538844933317602802047121528082298854292139034496025,\"S\":83836253134355891252395628558534222866311033418039146262386912657823592661207},\"B1\":[true,false,true,true,true,true,true,true,false,false],\"CS2\":{\"R\":28578847155903696420642616141070583041033136144988503442597941154558833417823,\"S\":4792624902640882210219044128084442282742375830610744089567360957013725377897},\"B2\":[true,true,true,false,false,true,true,false,true,true]},\"Timestamp\":1613994961431313,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[170,87,100,204,22,70,8,92,109,210,4,43,183,132,237,106,78,21,77,34,240,209,209,218,205,217,232,102,98,96,30,7],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[72,83,6,125,117,117,81,199,240,17,151,134,204,131,7,112,165,246,196,222,176,217,10,11,109,164,156,173,73,150,102,19],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[58,90,228,245,209,61,41,241,185,218,246,128,235,27,165,171,166,53,143,51,255,105,39,205,15,237,149,248,239,53,35,211]},\"NumTxs\":0,\"MinerPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[223,20,253,250,182,22,129,200,91,138,135,112,116,82,36,240,95,245,145,241,232,32,6,112,229,97,78,46,112,105,247,150],\"Cosigs\":{\"CS1\":{\"R\":3916020268539532848467644216722860137824554821660678985761908595132973072714,\"S\":97045995147022036660902295039327684939913829468538947744163884834333849946847},\"B1\":[true,true,true,true,true,true,false,true,false,false],\"CS2\":{\"R\":92028961177392717940679859135217775710532087351598448092697854633151818843931,\"S\":111389800061060649664721171870250671692292738272837785727000701591328031299981},\"B2\":[true,false,true,true,true,true,true,true,false,false]},\"Timestamp\":1613994931556497,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[208,223,176,86,146,162,2,192,168,3,21,120,94,74,135,8,59,245,169,91,185,41,215,1,80,100,251,132,248,159,106,7],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C\":{\"IpAddress\":2052726836,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"/f1RFgKiztjTn2LiZ0T6t/494ctwMkaWBe0k+MGLuJE=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}}}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("2052726836", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57",
		},
		{
			PubKey: "0239D4CAE39A7AC2F285796BABF7D28DC8EB7767E78409C70926D0929EA2941E36",
		},
		{
			PubKey: "02D2D695D4A352412E0D32A8BDF6EA3A606D35FE2C2F850C54D68727D065894986",
		},
		{
			PubKey: "02E5E1BE6C924349F2C2B20CE05A2650B3E56C7722A2E5952EE27D12DEE7A4A6E6",
		},
		{
			PubKey: "0300AB86B413FAA64A52FB61B5A28A6C361F87A5B0871C4F01C394D261415B0989",
		},
		{
			PubKey: "03019AF5B10FFE09FB0EE02B59195EF5E6F5BE51D17EAF5604EA452078CD465C4B",
		},
		{
			PubKey: "0323086D473DF937B6297FB755FA8E57C0FB2760512AED7757748B597C48F797A0",
		},
		{
			PubKey: "032AEE20CFC59EAEB7838DAC2A9BAF96C8D69CF2C866FB4A3F1DFB02BCFCA356BB",
		},
		{
			PubKey: "033207325A3CC671034FEBA86EC8D0AA412DF60C7E8292044D510DF582787DCC05",
		},
		{
			PubKey: "03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C",
			Peer: core.Peer{
				IpAddress:      ipaddr,
				ListenPortHost: 33133,
			},
		},
	}
	txBlockAndDsComm := &TxBlockAndDsComm{
		TxBlock: &txblock,
		DsBlock: &dsblock,
		DsComm:  initComm,
	}
	txBlockAndDsCommRaw, _ := json.Marshal(txBlockAndDsComm)
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ZILChainID
	param.GenesisHeader = txBlockAndDsCommRaw
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{}
	n, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	zilHeader := NewHandler()
	err2 := zilHeader.SyncGenesisHeader(n)
	assert.Equal(t, OPERATOR_ERROR, typeOfError(err2), err2)

}

func TestSyncGenesisHeaderTwice(t *testing.T) {
	txBlock1Raw := "{\"BlockHash\":[103,194,252,74,171,72,88,227,100,137,183,27,173,193,159,52,147,39,121,46,144,57,3,105,193,153,93,214,30,122,207,170],\"Cosigs\":{\"CS1\":{\"R\":19759839172862701417386858538844933317602802047121528082298854292139034496025,\"S\":83836253134355891252395628558534222866311033418039146262386912657823592661207},\"B1\":[true,false,true,true,true,true,true,true,false,false],\"CS2\":{\"R\":28578847155903696420642616141070583041033136144988503442597941154558833417823,\"S\":4792624902640882210219044128084442282742375830610744089567360957013725377897},\"B2\":[true,true,true,false,false,true,true,false,true,true]},\"Timestamp\":1613994961431313,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[170,87,100,204,22,70,8,92,109,210,4,43,183,132,237,106,78,21,77,34,240,209,209,218,205,217,232,102,98,96,30,7],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[72,83,6,125,117,117,81,199,240,17,151,134,204,131,7,112,165,246,196,222,176,217,10,11,109,164,156,173,73,150,102,19],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[58,90,228,245,209,61,41,241,185,218,246,128,235,27,165,171,166,53,143,51,255,105,39,205,15,237,149,248,239,53,35,211]},\"NumTxs\":0,\"MinerPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[223,20,253,250,182,22,129,200,91,138,135,112,116,82,36,240,95,245,145,241,232,32,6,112,229,97,78,46,112,105,247,150],\"Cosigs\":{\"CS1\":{\"R\":3916020268539532848467644216722860137824554821660678985761908595132973072714,\"S\":97045995147022036660902295039327684939913829468538947744163884834333849946847},\"B1\":[true,true,true,true,true,true,false,true,false,false],\"CS2\":{\"R\":92028961177392717940679859135217775710532087351598448092697854633151818843931,\"S\":111389800061060649664721171870250671692292738272837785727000701591328031299981},\"B2\":[true,false,true,true,true,true,true,true,false,false]},\"Timestamp\":1613994931556497,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[208,223,176,86,146,162,2,192,168,3,21,120,94,74,135,8,59,245,169,91,185,41,215,1,80,100,251,132,248,159,106,7],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C\":{\"IpAddress\":2052726836,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"/f1RFgKiztjTn2LiZ0T6t/494ctwMkaWBe0k+MGLuJE=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}}}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("2052726836", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57",
		},
		{
			PubKey: "0239D4CAE39A7AC2F285796BABF7D28DC8EB7767E78409C70926D0929EA2941E36",
		},
		{
			PubKey: "02D2D695D4A352412E0D32A8BDF6EA3A606D35FE2C2F850C54D68727D065894986",
		},
		{
			PubKey: "02E5E1BE6C924349F2C2B20CE05A2650B3E56C7722A2E5952EE27D12DEE7A4A6E6",
		},
		{
			PubKey: "0300AB86B413FAA64A52FB61B5A28A6C361F87A5B0871C4F01C394D261415B0989",
		},
		{
			PubKey: "03019AF5B10FFE09FB0EE02B59195EF5E6F5BE51D17EAF5604EA452078CD465C4B",
		},
		{
			PubKey: "0323086D473DF937B6297FB755FA8E57C0FB2760512AED7757748B597C48F797A0",
		},
		{
			PubKey: "032AEE20CFC59EAEB7838DAC2A9BAF96C8D69CF2C866FB4A3F1DFB02BCFCA356BB",
		},
		{
			PubKey: "033207325A3CC671034FEBA86EC8D0AA412DF60C7E8292044D510DF582787DCC05",
		},
		{
			PubKey: "03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C",
			Peer: core.Peer{
				IpAddress:      ipaddr,
				ListenPortHost: 33133,
			},
		},
	}
	txBlockAndDsComm := &TxBlockAndDsComm{
		TxBlock: &txblock,
		DsBlock: &dsblock,
		DsComm:  initComm,
	}
	txBlockAndDsCommRaw, _ := json.Marshal(txBlockAndDsComm)
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ZILChainID
	param.GenesisHeader = txBlockAndDsCommRaw
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	n, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	zilHeader := NewHandler()

	err1 := zilHeader.SyncGenesisHeader(n)
	assert.Equal(t, SUCCESS, typeOfError(err1))

	err2 := zilHeader.SyncGenesisHeader(n)
	assert.Equal(t, GENESIS_INITIALIZED, typeOfError(err2), err2)
}

func TestSyncGenesisHeader_ParamError(t *testing.T) {
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ZILChainID
	param.GenesisHeader = nil
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	n, _ := NewNative(sink.Bytes(), tx, nil)
	handler := NewHandler()
	err := handler.SyncGenesisHeader(n)
	assert.Equal(t, SYNCBLOCK_PARAM_ERROR, typeOfError(err), err)
}

func TestSyncBlockHeader(t *testing.T) {
	txBlock1Raw := "{\"BlockHash\":[103,194,252,74,171,72,88,227,100,137,183,27,173,193,159,52,147,39,121,46,144,57,3,105,193,153,93,214,30,122,207,170],\"Cosigs\":{\"CS1\":{\"R\":19759839172862701417386858538844933317602802047121528082298854292139034496025,\"S\":83836253134355891252395628558534222866311033418039146262386912657823592661207},\"B1\":[true,false,true,true,true,true,true,true,false,false],\"CS2\":{\"R\":28578847155903696420642616141070583041033136144988503442597941154558833417823,\"S\":4792624902640882210219044128084442282742375830610744089567360957013725377897},\"B2\":[true,true,true,false,false,true,true,false,true,true]},\"Timestamp\":1613994961431313,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[170,87,100,204,22,70,8,92,109,210,4,43,183,132,237,106,78,21,77,34,240,209,209,218,205,217,232,102,98,96,30,7],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[72,83,6,125,117,117,81,199,240,17,151,134,204,131,7,112,165,246,196,222,176,217,10,11,109,164,156,173,73,150,102,19],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[58,90,228,245,209,61,41,241,185,218,246,128,235,27,165,171,166,53,143,51,255,105,39,205,15,237,149,248,239,53,35,211]},\"NumTxs\":0,\"MinerPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[223,20,253,250,182,22,129,200,91,138,135,112,116,82,36,240,95,245,145,241,232,32,6,112,229,97,78,46,112,105,247,150],\"Cosigs\":{\"CS1\":{\"R\":3916020268539532848467644216722860137824554821660678985761908595132973072714,\"S\":97045995147022036660902295039327684939913829468538947744163884834333849946847},\"B1\":[true,true,true,true,true,true,false,true,false,false],\"CS2\":{\"R\":92028961177392717940679859135217775710532087351598448092697854633151818843931,\"S\":111389800061060649664721171870250671692292738272837785727000701591328031299981},\"B2\":[true,false,true,true,true,true,true,true,false,false]},\"Timestamp\":1613994931556497,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[208,223,176,86,146,162,2,192,168,3,21,120,94,74,135,8,59,245,169,91,185,41,215,1,80,100,251,132,248,159,106,7],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C\":{\"IpAddress\":2052726836,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"/f1RFgKiztjTn2LiZ0T6t/494ctwMkaWBe0k+MGLuJE=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}}}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("2052726836", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57",
		},
		{
			PubKey: "0239D4CAE39A7AC2F285796BABF7D28DC8EB7767E78409C70926D0929EA2941E36",
		},
		{
			PubKey: "02D2D695D4A352412E0D32A8BDF6EA3A606D35FE2C2F850C54D68727D065894986",
		},
		{
			PubKey: "02E5E1BE6C924349F2C2B20CE05A2650B3E56C7722A2E5952EE27D12DEE7A4A6E6",
		},
		{
			PubKey: "0300AB86B413FAA64A52FB61B5A28A6C361F87A5B0871C4F01C394D261415B0989",
		},
		{
			PubKey: "03019AF5B10FFE09FB0EE02B59195EF5E6F5BE51D17EAF5604EA452078CD465C4B",
		},
		{
			PubKey: "0323086D473DF937B6297FB755FA8E57C0FB2760512AED7757748B597C48F797A0",
		},
		{
			PubKey: "032AEE20CFC59EAEB7838DAC2A9BAF96C8D69CF2C866FB4A3F1DFB02BCFCA356BB",
		},
		{
			PubKey: "033207325A3CC671034FEBA86EC8D0AA412DF60C7E8292044D510DF582787DCC05",
		},
		{
			PubKey: "03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C",
			Peer: core.Peer{
				IpAddress:      ipaddr,
				ListenPortHost: 33133,
			},
		},
	}
	txBlockAndDsComm := &TxBlockAndDsComm{
		TxBlock: &txblock,
		DsBlock: &dsblock,
		DsComm:  initComm,
	}
	txBlockAndDsCommRaw, _ := json.Marshal(txBlockAndDsComm)
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ZILChainID
	param.GenesisHeader = txBlockAndDsCommRaw
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	n, err := NewNative(sink.Bytes(), tx, nil)
	assert.NilError(t, err)
	zilHeader := NewHandler()

	err1 := zilHeader.SyncGenesisHeader(n)
	assert.Equal(t, SUCCESS, typeOfError(err1))

	{
		tx2 := "{\"BlockHash\":[220,123,128,146,170,253,111,45,52,150,204,33,39,67,185,235,104,238,133,10,101,139,114,126,94,98,189,99,220,93,104,203],\"Cosigs\":{\"CS1\":{\"R\":73879738620022999057765326504685592497771384639737070927541313517793667413261,\"S\":64529815331723787857729050912215558502224760488546369745524502843967035778091},\"B1\":[true,true,true,false,true,true,true,false,true,false],\"CS2\":{\"R\":23895515928550040283944537764516951680546857175643779656200050518047854152239,\"S\":108449727091413929440089808187769645312285742303517641149552933758581667443466},\"B2\":[true,true,true,true,false,false,true,false,true,true]},\"Timestamp\":1613994984033984,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[170,87,100,204,22,70,8,92,109,210,4,43,183,132,237,106,78,21,77,34,240,209,209,218,205,217,232,102,98,96,30,7],\"PrevHash\":[103,194,252,74,171,72,88,227,100,137,183,27,173,193,159,52,147,39,121,46,144,57,3,105,193,153,93,214,30,122,207,170]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":2,\"HashSet\":{\"StateRootHash\":[72,83,6,125,117,117,81,199,240,17,151,134,204,131,7,112,165,246,196,222,176,217,10,11,109,164,156,173,73,150,102,19],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[137,177,199,243,63,134,57,82,75,149,249,146,12,68,112,183,174,249,224,234,23,210,154,72,249,241,210,236,190,7,70,150]},\"NumTxs\":0,\"MinerPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"DSBlockNum\":1}}"
		var txBlock2 core.TxBlock
		json.Unmarshal([]byte(tx2), &txBlock2)
		txHeader2 := core.TxBlockOrDsBlock{
			DsBlock: nil,
			TxBlock: &txBlock2,
		}
		txBlock2Raw, _ := json.Marshal(txHeader2)

		ds2 := "{\"BlockHash\":[41,255,112,79,121,52,44,108,27,78,149,97,128,169,142,99,161,243,156,238,133,104,71,96,145,183,135,39,197,49,70,6],\"Cosigs\":{\"CS1\":{\"R\":103090883013418352191376392635877978567742961591351067462257838644847542617753,\"S\":73192789979531658691833346038151699109358057858171155520758710512729424155305},\"B1\":[true,true,true,false,false,true,false,true,true,true],\"CS2\":{\"R\":21094834133468218980862701513066743117080953298923136490727566534783122989520,\"S\":77080847615133551472581432869554298318934976820868819083893806928539647213534},\"B2\":[true,true,true,false,true,true,false,true,true,false]},\"Timestamp\":1613995092296300,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[170,87,100,204,22,70,8,92,109,210,4,43,183,132,237,106,78,21,77,34,240,209,209,218,205,217,232,102,98,96,30,7],\"PrevHash\":[223,20,253,250,182,22,129,200,91,138,135,112,116,82,36,240,95,245,145,241,232,32,6,112,229,97,78,46,112,105,247,150]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x0213D5A7F74B28F3F588FF6520748DBB541986E98F75FA78D6334B2D0AAB4C1E57\",\"BlockNum\":2,\"EpochNum\":5,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x0334AA0F7CA2EAA56B6B752533F9C60777E96C6D1ABE84B463F60ADD89843794AE\":{\"IpAddress\":1479260470,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"LWcNwYT+9n9M/gFquxF2nTd7tw03elRwimzHfzivGHI=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}}}\n"
		var dsBlock core.DsBlock
		json.Unmarshal([]byte(ds2), &dsBlock)
		dxBlock2 := core.TxBlockOrDsBlock{
			DsBlock: &dsBlock,
			TxBlock: nil,
		}
		dsBlock2Raw, _ := json.Marshal(dxBlock2)

		ds3 := "{\"BlockHash\":[33,188,238,210,49,208,154,245,110,16,59,110,164,235,9,36,48,10,40,41,8,135,66,2,145,13,137,11,64,206,177,217],\"Cosigs\":{\"CS1\":{\"R\":70005379043402879103191264842231553636814441841740525194020377184388453204537,\"S\":105784047080749507945716487656214776525296006169790096780085348989744530236603},\"B1\":[true,true,true,true,true,true,false,false,true,false],\"CS2\":{\"R\":19557218421616002213178764541557273522495340440998089712397058397979776655307,\"S\":29919240197457228433914090511792723453076169015930722604328227096249633759414},\"B2\":[true,true,true,true,true,true,false,false,true,false]},\"Timestamp\":1613995275718835,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[208,223,176,86,146,162,2,192,168,3,21,120,94,74,135,8,59,245,169,91,185,41,215,1,80,100,251,132,248,159,106,7],\"PrevHash\":[41,255,112,79,121,52,44,108,27,78,149,97,128,169,142,99,161,243,156,238,133,104,71,96,145,183,135,39,197,49,70,6]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x033207325A3CC671034FEBA86EC8D0AA412DF60C7E8292044D510DF582787DCC05\",\"BlockNum\":3,\"EpochNum\":10,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x03AFEEA358BBFD6A350B59943E11D65386142033C63B921032608DE65334C1DF8C\":{\"IpAddress\":2052726836,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"meelz7e1WvDJj0mHOlDHtRwPduNBrc9viwP/eXeAzKo=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}}}"
		var dsBlock3 core.DsBlock
		json.Unmarshal([]byte(ds3), &dsBlock3)

		dxBlock3 := core.TxBlockOrDsBlock{
			DsBlock: &dsBlock3,
			TxBlock: nil,
		}
		dsBlock3Raw, _ := json.Marshal(dxBlock3)

		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = ZILChainID
		param.Headers = append(param.Headers, txBlock2Raw)
		param.Headers = append(param.Headers, dsBlock2Raw)
		param.Headers = append(param.Headers, dsBlock3Raw)

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, _ := NewNative(sink.Bytes(), tx, n.GetCacheDB())

		err := zilHeader.SyncBlockHeader(native)
		assert.Equal(t, SUCCESS, typeOfError(err), err)

	}

}
