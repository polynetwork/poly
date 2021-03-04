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
		ChainId:   ZILChainID,
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
	txBlock1Raw := "{\"BlockHash\":[56,40,135,50,178,230,126,194,104,230,177,166,241,195,181,119,72,230,177,102,171,121,58,163,41,139,18,92,138,231,108,39],\"Cosigs\":{\"CS1\":{\"R\":79461090997780129048034156976579207017607593312295382854180954812611062499786,\"S\":91742597770497613815760351313815911030151335677216474693626678776993533665077},\"B1\":[true,true,true,true,true,true,true,false,false,false],\"CS2\":{\"R\":22513373955460225598459727159582327633978685137898384968505490930737451385826,\"S\":69259600331273345477024713698850194580163873601064291944707625774006410868491},\"B2\":[true,true,false,true,true,true,true,true,false,false]},\"Timestamp\":1614851084113383,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[144,78,35,242,84,150,244,171,215,191,207,200,228,18,4,75,188,156,242,96,234,28,171,227,90,127,173,150,197,48,76,231],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[171,57,165,166,188,170,165,153,119,109,69,231,171,86,24,230,64,155,13,154,233,104,156,53,214,30,42,57,70,180,219,46],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[59,61,191,206,53,105,23,193,9,0,195,32,41,52,29,157,182,192,2,221,165,75,6,239,121,24,166,25,78,63,43,201]},\"NumTxs\":0,\"MinerPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[110,156,68,14,64,80,203,185,47,33,51,251,33,253,134,144,165,106,177,248,40,162,95,175,149,198,226,104,42,133,141,147],\"Cosigs\":{\"CS1\":{\"R\":82378159645007731822019453933480162750953116703416061439006821187696821549829,\"S\":76062990255435928402343339335249009291551400166495738641213601543709385329901},\"B1\":[true,true,true,false,true,true,true,true,false,false],\"CS2\":{\"R\":93854453672214038567945187798065701969482439716186198113435402044731863802032,\"S\":99194416943091049405554646407147384870804679504363691825129226286226313434999},\"B2\":[true,true,true,true,true,true,false,false,true,false]},\"Timestamp\":1614851053611705,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[168,148,171,148,251,117,232,182,248,255,95,207,117,54,148,68,162,158,59,213,12,56,135,233,222,70,197,78,138,205,243,168],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2\":{\"IpAddress\":3672036406,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"0VOZ8Rwe5L9/H5LD5FtSOWf9XK5dSilsYmSbzoF7Bjo=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}},\"PrevDSHash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("3672036406", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48",
		},
		{
			PubKey: "021D439D1CCCAE17C3D6E855BC78E96438C808D16D1CBF8D7ABD391E41CEE9B1BF",
		},
		{
			PubKey: "021EDDE95598F5F59708D2E728E00EDB2ECF278C16BD389384320B1AF998DCC2FD",
		},
		{
			PubKey: "02445FE498E7FBB240BDF9185EB5E7642AF1AF36852D1E132E198A222FBAC617A0",
		},
		{
			PubKey: "0256EC4BC62FB56C83A3F6160E67499A9E381CF7A613EBF34B9ECDB9E64171DDF4",
		},
		{
			PubKey: "0264D991762D81DD6557BCB33EC8AA3F621B4CB790852F2231C864921387B76862",
		},
		{
			PubKey: "027A00916BDD3CF954ED13A0494BFB73FF95BF28C54004F2749F1A8E8CC1AB5B3D",
		},
		{
			PubKey: "0297C693FBEBAF397CBDE616F605920EF70D7F6E5EC8DD82E71AE1E812E5E0B303",
		},
		{
			PubKey: "02AE5ADF63E9161000713987B5EBB490B5E6B57CF5B7F9799B4AB907BA19D468F6",
		},
		{
			PubKey: "0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2",
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
	// [56,40,135,50,178,230,126,194,104,230,177,166,241,195,181,119,72,230,177,102,171,121,58,163,41,139,18,92,138,231,108,39]
	assert.Equal(t, "38288732b2e67ec268e6b1a6f1c3b57748e6b166ab793aa3298b125c8ae76c27", util.EncodeHex(headerHash))
}

func TestSyncGenesisHeaderNoOperator(t *testing.T) {
	txBlock1Raw := "{\"BlockHash\":[56,40,135,50,178,230,126,194,104,230,177,166,241,195,181,119,72,230,177,102,171,121,58,163,41,139,18,92,138,231,108,39],\"Cosigs\":{\"CS1\":{\"R\":79461090997780129048034156976579207017607593312295382854180954812611062499786,\"S\":91742597770497613815760351313815911030151335677216474693626678776993533665077},\"B1\":[true,true,true,true,true,true,true,false,false,false],\"CS2\":{\"R\":22513373955460225598459727159582327633978685137898384968505490930737451385826,\"S\":69259600331273345477024713698850194580163873601064291944707625774006410868491},\"B2\":[true,true,false,true,true,true,true,true,false,false]},\"Timestamp\":1614851084113383,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[144,78,35,242,84,150,244,171,215,191,207,200,228,18,4,75,188,156,242,96,234,28,171,227,90,127,173,150,197,48,76,231],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[171,57,165,166,188,170,165,153,119,109,69,231,171,86,24,230,64,155,13,154,233,104,156,53,214,30,42,57,70,180,219,46],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[59,61,191,206,53,105,23,193,9,0,195,32,41,52,29,157,182,192,2,221,165,75,6,239,121,24,166,25,78,63,43,201]},\"NumTxs\":0,\"MinerPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[110,156,68,14,64,80,203,185,47,33,51,251,33,253,134,144,165,106,177,248,40,162,95,175,149,198,226,104,42,133,141,147],\"Cosigs\":{\"CS1\":{\"R\":82378159645007731822019453933480162750953116703416061439006821187696821549829,\"S\":76062990255435928402343339335249009291551400166495738641213601543709385329901},\"B1\":[true,true,true,false,true,true,true,true,false,false],\"CS2\":{\"R\":93854453672214038567945187798065701969482439716186198113435402044731863802032,\"S\":99194416943091049405554646407147384870804679504363691825129226286226313434999},\"B2\":[true,true,true,true,true,true,false,false,true,false]},\"Timestamp\":1614851053611705,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[168,148,171,148,251,117,232,182,248,255,95,207,117,54,148,68,162,158,59,213,12,56,135,233,222,70,197,78,138,205,243,168],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2\":{\"IpAddress\":3672036406,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"0VOZ8Rwe5L9/H5LD5FtSOWf9XK5dSilsYmSbzoF7Bjo=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}},\"PrevDSHash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("3672036406", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48",
		},
		{
			PubKey: "021D439D1CCCAE17C3D6E855BC78E96438C808D16D1CBF8D7ABD391E41CEE9B1BF",
		},
		{
			PubKey: "021EDDE95598F5F59708D2E728E00EDB2ECF278C16BD389384320B1AF998DCC2FD",
		},
		{
			PubKey: "02445FE498E7FBB240BDF9185EB5E7642AF1AF36852D1E132E198A222FBAC617A0",
		},
		{
			PubKey: "0256EC4BC62FB56C83A3F6160E67499A9E381CF7A613EBF34B9ECDB9E64171DDF4",
		},
		{
			PubKey: "0264D991762D81DD6557BCB33EC8AA3F621B4CB790852F2231C864921387B76862",
		},
		{
			PubKey: "027A00916BDD3CF954ED13A0494BFB73FF95BF28C54004F2749F1A8E8CC1AB5B3D",
		},
		{
			PubKey: "0297C693FBEBAF397CBDE616F605920EF70D7F6E5EC8DD82E71AE1E812E5E0B303",
		},
		{
			PubKey: "02AE5ADF63E9161000713987B5EBB490B5E6B57CF5B7F9799B4AB907BA19D468F6",
		},
		{
			PubKey: "0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2",
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
	txBlock1Raw := "{\"BlockHash\":[56,40,135,50,178,230,126,194,104,230,177,166,241,195,181,119,72,230,177,102,171,121,58,163,41,139,18,92,138,231,108,39],\"Cosigs\":{\"CS1\":{\"R\":79461090997780129048034156976579207017607593312295382854180954812611062499786,\"S\":91742597770497613815760351313815911030151335677216474693626678776993533665077},\"B1\":[true,true,true,true,true,true,true,false,false,false],\"CS2\":{\"R\":22513373955460225598459727159582327633978685137898384968505490930737451385826,\"S\":69259600331273345477024713698850194580163873601064291944707625774006410868491},\"B2\":[true,true,false,true,true,true,true,true,false,false]},\"Timestamp\":1614851084113383,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[144,78,35,242,84,150,244,171,215,191,207,200,228,18,4,75,188,156,242,96,234,28,171,227,90,127,173,150,197,48,76,231],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[171,57,165,166,188,170,165,153,119,109,69,231,171,86,24,230,64,155,13,154,233,104,156,53,214,30,42,57,70,180,219,46],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[59,61,191,206,53,105,23,193,9,0,195,32,41,52,29,157,182,192,2,221,165,75,6,239,121,24,166,25,78,63,43,201]},\"NumTxs\":0,\"MinerPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[110,156,68,14,64,80,203,185,47,33,51,251,33,253,134,144,165,106,177,248,40,162,95,175,149,198,226,104,42,133,141,147],\"Cosigs\":{\"CS1\":{\"R\":82378159645007731822019453933480162750953116703416061439006821187696821549829,\"S\":76062990255435928402343339335249009291551400166495738641213601543709385329901},\"B1\":[true,true,true,false,true,true,true,true,false,false],\"CS2\":{\"R\":93854453672214038567945187798065701969482439716186198113435402044731863802032,\"S\":99194416943091049405554646407147384870804679504363691825129226286226313434999},\"B2\":[true,true,true,true,true,true,false,false,true,false]},\"Timestamp\":1614851053611705,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[168,148,171,148,251,117,232,182,248,255,95,207,117,54,148,68,162,158,59,213,12,56,135,233,222,70,197,78,138,205,243,168],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2\":{\"IpAddress\":3672036406,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"0VOZ8Rwe5L9/H5LD5FtSOWf9XK5dSilsYmSbzoF7Bjo=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}},\"PrevDSHash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("3672036406", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48",
		},
		{
			PubKey: "021D439D1CCCAE17C3D6E855BC78E96438C808D16D1CBF8D7ABD391E41CEE9B1BF",
		},
		{
			PubKey: "021EDDE95598F5F59708D2E728E00EDB2ECF278C16BD389384320B1AF998DCC2FD",
		},
		{
			PubKey: "02445FE498E7FBB240BDF9185EB5E7642AF1AF36852D1E132E198A222FBAC617A0",
		},
		{
			PubKey: "0256EC4BC62FB56C83A3F6160E67499A9E381CF7A613EBF34B9ECDB9E64171DDF4",
		},
		{
			PubKey: "0264D991762D81DD6557BCB33EC8AA3F621B4CB790852F2231C864921387B76862",
		},
		{
			PubKey: "027A00916BDD3CF954ED13A0494BFB73FF95BF28C54004F2749F1A8E8CC1AB5B3D",
		},
		{
			PubKey: "0297C693FBEBAF397CBDE616F605920EF70D7F6E5EC8DD82E71AE1E812E5E0B303",
		},
		{
			PubKey: "02AE5ADF63E9161000713987B5EBB490B5E6B57CF5B7F9799B4AB907BA19D468F6",
		},
		{
			PubKey: "0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2",
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
	txBlock1Raw := "{\"BlockHash\":[56,40,135,50,178,230,126,194,104,230,177,166,241,195,181,119,72,230,177,102,171,121,58,163,41,139,18,92,138,231,108,39],\"Cosigs\":{\"CS1\":{\"R\":79461090997780129048034156976579207017607593312295382854180954812611062499786,\"S\":91742597770497613815760351313815911030151335677216474693626678776993533665077},\"B1\":[true,true,true,true,true,true,true,false,false,false],\"CS2\":{\"R\":22513373955460225598459727159582327633978685137898384968505490930737451385826,\"S\":69259600331273345477024713698850194580163873601064291944707625774006410868491},\"B2\":[true,true,false,true,true,true,true,true,false,false]},\"Timestamp\":1614851084113383,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[144,78,35,242,84,150,244,171,215,191,207,200,228,18,4,75,188,156,242,96,234,28,171,227,90,127,173,150,197,48,76,231],\"PrevHash\":[25,71,113,139,67,29,37,221,101,194,38,247,159,62,10,156,201,106,148,136,153,218,179,66,41,147,222,241,73,74,156,149]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":1,\"HashSet\":{\"StateRootHash\":[171,57,165,166,188,170,165,153,119,109,69,231,171,86,24,230,64,155,13,154,233,104,156,53,214,30,42,57,70,180,219,46],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[59,61,191,206,53,105,23,193,9,0,195,32,41,52,29,157,182,192,2,221,165,75,6,239,121,24,166,25,78,63,43,201]},\"NumTxs\":0,\"MinerPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"DSBlockNum\":1}}"
	var txblock core.TxBlock
	json.Unmarshal([]byte(txBlock1Raw), &txblock)
	dsBlock1Raw := "{\"BlockHash\":[110,156,68,14,64,80,203,185,47,33,51,251,33,253,134,144,165,106,177,248,40,162,95,175,149,198,226,104,42,133,141,147],\"Cosigs\":{\"CS1\":{\"R\":82378159645007731822019453933480162750953116703416061439006821187696821549829,\"S\":76062990255435928402343339335249009291551400166495738641213601543709385329901},\"B1\":[true,true,true,false,true,true,true,true,false,false],\"CS2\":{\"R\":93854453672214038567945187798065701969482439716186198113435402044731863802032,\"S\":99194416943091049405554646407147384870804679504363691825129226286226313434999},\"B2\":[true,true,true,true,true,true,false,false,true,false]},\"Timestamp\":1614851053611705,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[168,148,171,148,251,117,232,182,248,255,95,207,117,54,148,68,162,158,59,213,12,56,135,233,222,70,197,78,138,205,243,168],\"PrevHash\":[15,0,233,211,23,83,0,252,40,120,18,210,1,237,207,191,203,129,101,128,150,6,84,85,149,191,83,112,12,82,70,72]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"BlockNum\":1,\"EpochNum\":1,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2\":{\"IpAddress\":3672036406,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"0VOZ8Rwe5L9/H5LD5FtSOWf9XK5dSilsYmSbzoF7Bjo=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}},\"PrevDSHash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}"
	var dsblock core.DsBlock
	json.Unmarshal([]byte(dsBlock1Raw), &dsblock)
	ipaddr, _ := new(big.Int).SetString("3672036406", 10)
	initComm := []core.PairOfNode{
		{
			PubKey: "02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48",
		},
		{
			PubKey: "021D439D1CCCAE17C3D6E855BC78E96438C808D16D1CBF8D7ABD391E41CEE9B1BF",
		},
		{
			PubKey: "021EDDE95598F5F59708D2E728E00EDB2ECF278C16BD389384320B1AF998DCC2FD",
		},
		{
			PubKey: "02445FE498E7FBB240BDF9185EB5E7642AF1AF36852D1E132E198A222FBAC617A0",
		},
		{
			PubKey: "0256EC4BC62FB56C83A3F6160E67499A9E381CF7A613EBF34B9ECDB9E64171DDF4",
		},
		{
			PubKey: "0264D991762D81DD6557BCB33EC8AA3F621B4CB790852F2231C864921387B76862",
		},
		{
			PubKey: "027A00916BDD3CF954ED13A0494BFB73FF95BF28C54004F2749F1A8E8CC1AB5B3D",
		},
		{
			PubKey: "0297C693FBEBAF397CBDE616F605920EF70D7F6E5EC8DD82E71AE1E812E5E0B303",
		},
		{
			PubKey: "02AE5ADF63E9161000713987B5EBB490B5E6B57CF5B7F9799B4AB907BA19D468F6",
		},
		{
			PubKey: "0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2",
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
		tx2 := "{\"BlockHash\":[163,89,34,108,0,230,155,70,16,129,194,188,23,67,41,21,61,189,90,209,20,180,182,44,198,39,61,181,52,30,53,243],\"Cosigs\":{\"CS1\":{\"R\":20525929347658909011168352167799173309070658876050417230429696088964525592242,\"S\":2219392327625100730171301532015490664087961938867140359806968133574873232168},\"B1\":[true,false,false,true,true,true,false,true,true,true],\"CS2\":{\"R\":53749525636648123624118982425909778014109266050590580304633769798569847769793,\"S\":56608271032305944506610174732540992807907583290460864606501026242732057361399},\"B2\":[true,true,true,true,false,true,true,false,true,false]},\"Timestamp\":1614851111306746,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":1,\"CommitteeHash\":[144,78,35,242,84,150,244,171,215,191,207,200,228,18,4,75,188,156,242,96,234,28,171,227,90,127,173,150,197,48,76,231],\"PrevHash\":[56,40,135,50,178,230,126,194,104,230,177,166,241,195,181,119,72,230,177,102,171,121,58,163,41,139,18,92,138,231,108,39]},\"GasLimit\":90000,\"GasUsed\":0,\"Rewards\":0,\"BlockNum\":2,\"HashSet\":{\"StateRootHash\":[171,57,165,166,188,170,165,153,119,109,69,231,171,86,24,230,64,155,13,154,233,104,156,53,214,30,42,57,70,180,219,46],\"DeltaHash\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"MbInfoHash\":[39,190,11,72,215,247,139,136,191,194,94,170,44,12,80,251,161,196,104,50,158,124,76,87,2,22,235,206,70,5,118,35]},\"NumTxs\":0,\"MinerPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"DSBlockNum\":1}}"
		var txBlock2 core.TxBlock
		json.Unmarshal([]byte(tx2), &txBlock2)
		txHeader2 := core.TxBlockOrDsBlock{
			DsBlock: nil,
			TxBlock: &txBlock2,
		}
		txBlock2Raw, _ := json.Marshal(txHeader2)

		ds2 := "{\"BlockHash\":[205,19,106,177,186,166,122,7,228,4,225,112,33,137,186,176,74,45,197,236,192,25,11,3,136,169,12,13,125,190,113,101],\"Cosigs\":{\"CS1\":{\"R\":113845197918272214030248143119634993813849346187160330249100390280078987321075,\"S\":11960680464418859657288883008275781310992984177885610702751259800671104357467},\"B1\":[true,true,true,true,false,true,true,true,false,false],\"CS2\":{\"R\":7234070785416569578761825307955696574556888586709383967879349923626005495389,\"S\":97178275103152357684957378624466229863078242905485616872025908445142523648308},\"B2\":[true,true,true,true,true,false,false,true,true,false]},\"Timestamp\":1614851217734567,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[144,78,35,242,84,150,244,171,215,191,207,200,228,18,4,75,188,156,242,96,234,28,171,227,90,127,173,150,197,48,76,231],\"PrevHash\":[110,156,68,14,64,80,203,185,47,33,51,251,33,253,134,144,165,106,177,248,40,162,95,175,149,198,226,104,42,133,141,147]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x02105342331FCD7CA95648DF8C5373C596982544F35E90849B1E619DFC59F03D48\",\"BlockNum\":2,\"EpochNum\":5,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x02D3CB3FFC8DDE2A55AC29D013CEB5636806C6FC61C5AF077B6313DC636027A602\":{\"IpAddress\":652191523,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":null,\"DSBlockHashSet\":{\"ShadingHash\":\"NUtjgqK3Kc3eRarrOxplUdHqkABFrKJ72eg3Vsbafhs=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}},\"PrevDSHash\":\"6e9c440e4050cbb92f2133fb21fd8690a56ab1f828a25faf95c6e2682a858d93\"}"
		var dsBlock core.DsBlock
		json.Unmarshal([]byte(ds2), &dsBlock)
		dxBlock2 := core.TxBlockOrDsBlock{
			DsBlock: &dsBlock,
			TxBlock: nil,
		}
		dsBlock2Raw, _ := json.Marshal(dxBlock2)

		ds3 := "{\"BlockHash\":[199,85,247,209,245,2,19,4,42,231,183,36,5,39,198,78,197,113,181,41,225,110,233,202,179,18,63,205,101,236,205,125],\"Cosigs\":{\"CS1\":{\"R\":110340268485404115175746702398383209217033662733336753183073802770775446464756,\"S\":52760024325796949149879390515431387711178642228852831431611015780002471233922},\"B1\":[false,true,false,true,true,true,false,true,true,true],\"CS2\":{\"R\":36533094902270957813438053098747117203126258416136078436733430097157722028362,\"S\":62712538696541879916678179078352886679713735102688192266931612601681694134230},\"B2\":[true,true,true,true,false,false,true,true,true,false]},\"Timestamp\":1614851402461407,\"BlockHeader\":{\"BlockHeaderBase\":{\"Version\":2,\"CommitteeHash\":[168,148,171,148,251,117,232,182,248,255,95,207,117,54,148,68,162,158,59,213,12,56,135,233,222,70,197,78,138,205,243,168],\"PrevHash\":[205,19,106,177,186,166,122,7,228,4,225,112,33,137,186,176,74,45,197,236,192,25,11,3,136,169,12,13,125,190,113,101]},\"DsDifficulty\":5,\"Difficulty\":3,\"LeaderPubKey\":\"0x0297C693FBEBAF397CBDE616F605920EF70D7F6E5EC8DD82E71AE1E812E5E0B303\",\"BlockNum\":3,\"EpochNum\":10,\"GasPrice\":\"2000000000\",\"SwInfo\":{\"ZilliqaMajorVersion\":0,\"ZilliqaMinorVersion\":0,\"ZilliqaFixVersion\":0,\"ZilliqaUpgradeDS\":0,\"ZilliqaCommit\":0,\"ScillaMajorVersion\":0,\"ScillaMinorVersion\":0,\"ScillaFixVersion\":0,\"ScillaUpgradeDS\":0,\"ScillaCommit\":0},\"PoWDSWinners\":{\"0x0374A5CA5D76BEE5A1DE132AE72184AB084D23EC7A4867CCD562C58405BBB663E2\":{\"IpAddress\":3672036406,\"ListenPortHost\":33133,\"HostName\":\"\"}},\"RemoveDSNodePubKeys\":[\"0x02D3CB3FFC8DDE2A55AC29D013CEB5636806C6FC61C5AF077B6313DC636027A602\"],\"DSBlockHashSet\":{\"ShadingHash\":\"YxF2nnl4gt/8n8UkCtpOe5TUuwXrhzEwdLAs4PUe0Y0=\",\"ReservedField\":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},\"GovDSShardVotesMap\":{}},\"PrevDSHash\":\"cd136ab1baa67a07e404e1702189bab04a2dc5ecc0190b0388a90c0d7dbe7165\"}"
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
