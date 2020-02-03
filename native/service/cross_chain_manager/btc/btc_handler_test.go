package btc

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/native"
	ccmcom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/header_sync/btc"
	hscom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/ontio/multi-chain/native/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	acct *account.Account = account.NewAccount("")

	fromBtcTxid       = "f922ed4669f5f11910649133b582d476a316f3e7392dad9be09d08a1c4a81357"
	fromBtcRawTx      = "0100000001c8b3b4ee22c2bd31ea518bc159a412a7e873b739cc018bd063b9426c7862632a000000006a47304402202b79a1a2c2439dc2fc7dc1eab186566164cf0d317a8e23b340745dd73408ecab0220216b34a2ebfa9e84922f29f17ad54c9510d4b427ef17c960eacc302de10e54f7012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff03a17200000000000022002044978a77e4e983136bf1cca277c45e5bd4eff6a7848e900416daf86fd32c274300000000000000003d6a3b6602000000000000000000000000000000140876c408a5b0f3ad9a65995658494dd088926d3e14f3b8a17f1f957f60c88f105e32ebff3f022e56a48379052a010000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"
	fromBtcProof      = "0000002008c4148bf546eef416ff870326532741c87fe540cca7dbc913bbf635f6900a61ebc2d2937be45e1cb45852b2c48bb533212d35b6cd08acd2094aab7418e97b58619d215effff7f2003000000020000000231f4144cace7184cecd1d6542ccacef88d8d504e507b0693c9fc14079a956dff5713a8c4a1089de09bad2d39e7f316a376d482b53391641019f1f56946ed22f90105"
	fromBtcMerkleRoot = "587be91874ab4a09d2ac08cdb6352d2133b58bc4b25258b41c5ee47b93d2c2eb"
	toOntAddr         = "AdzZ2VKufdJWeB8t9a8biXoHbbMe2kZeyH"
	obtcxAddr         = "3e6d9288d04d49585699659aadf3b0a508c47608"
	utxoKey           = "87a9652e9b396545598c0fc72cb5a98848bf93d3"

	sigs = []string{
		"3045022100c6f0620de7b8e71801408cc690b21ffa9ad344311b5e7373dcd4090316cc02d3022078bfb2bd6d3fcdbd75e1ce2dbefdecc28a4955d9d4dfed1b98aa59b6e83c0bf401",
		"304402202a525ae6d1c10ad428aa90559a8e1226915a441d1f5ed06ea577bafd36fce883022072235823b7a9e597d613c463e5d0eefc491092ca39f412db8e55f5a5aff5711301",
		"3044022026a05a7026dca55cdd6bb15d72c28ad0b08e4c3f810820a4069547a8f82069010220468d49b47f86d899255adf0fbf26e2997c426eb96d22f9634f7dac13d87642eb01",
		"3044022028bc153b6f149e020a99c12d64f5a7d9b213f358980dfa4fe556ad8c448ebe0502206db9f7b34c1658c75b45da0edf8c618b884ae0b9ca2346714f9bf28a99f7ec5a01",
		"30440220168ba28cb6e0daea0269ba358c69bb88a8b3209e71db2daccb63cb1d65651d4702201acab7292731940becb7169db0014e59bc9cf0910b121f8baf1976352f27060301",
	}

	signers = []string{
		"mj3LUsSvk9ZQH1pSHvC8LBtsYXsZvbky8H",
		"mtNiC48WWbGRk2zLqiTMwKLhrCk6rBqBen",
		"mi1bYK8SR3Qsf2cdrxgak3spzFx4EVH1pf",
		"mz3bTZaQ2tNzsn4szNE8R6gp5zyHuqN29V",
		"mfzbFf6njbEuyvZGDiAdfKamxWfAMv47NG",
	}

	getNativeFunc = func(args []byte, db *storage.CacheDB) *native.NativeService {
		if db == nil {
			store, _ := leveldbstore.NewMemLevelDBStore()
			db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		}

		return native.NewNativeService(db, nil, 0, 0, common.Uint256{0}, 0, args, false)
	}

	setSideChain = func(ns *native.NativeService) {
		side := &side_chain_manager.SideChain{
			Name:         "btc",
			ChainId:      0,
			BlocksToWait: 1,
			Router:       0,
		}
		sink := common.NewZeroCopySink(nil)
		_ = side.Serialization(sink)

		ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress,
			[]byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(0)), states.GenRawStorageItem(sink.Bytes()))
	}

	getSigs = func() [][]byte {
		res := make([][]byte, len(sigs))
		for i, sig := range sigs {
			raw, _ := hex.DecodeString(sig)
			res[i] = raw
		}
		return res
	}
)

func TestBTCHandler_MakeDepositProposal(t *testing.T) {
	gh := netParam.GenesisBlock.Header
	mr, _ := chainhash.NewHashFromStr(fromBtcMerkleRoot)
	gh.MerkleRoot = *mr

	db, _ := syncGenesisHeader(&gh)

	txid, _ := chainhash.NewHashFromStr(fromBtcTxid)
	rawTx, _ := hex.DecodeString(fromBtcRawTx)
	scAddr, _ := common.AddressFromHexString(obtcxAddr)
	handler := NewBTCHandler()

	// wrong proof
	wrongProof := "0100003037db655b09de3449fe60bc0838ef3541e28d3ae31a05093f1bb63e4845a6b102695fd2a687fc1fc368f13227c2bb1b6b0fcac9760936d869a9ba01f8a75f825c5105245effff7f20050000000200000002e0c8d9fb711dd377d0ba8d1c16c154b903432aa8923c87f3f0fd6045be7b8c8a51cf2962a492309e6bd7aa56848b2817c1a760f9fb0823762200d2b286f988b90105"
	proof, _ := hex.DecodeString(wrongProof)
	params := new(ccmcom.EntranceParam)
	params.Height = 0
	params.SourceChainID = 0
	params.Proof = proof
	params.Extra = rawTx
	params.RelayerAddress = acct.Address[:]

	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)
	ns := getNativeFunc(sink.Bytes(), db)
	setSideChain(ns)
	p, err := handler.MakeDepositProposal(ns)
	assert.Error(t, err)

	// normal case
	proof, _ = hex.DecodeString(fromBtcProof)
	params.Proof = proof

	sink.Reset()
	params.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), db)

	p, err = handler.MakeDepositProposal(ns)
	assert.NoError(t, err)

	ontAddr, _ := common.AddressFromBase58(toOntAddr)
	sink.Reset()
	sink.WriteVarBytes(ontAddr[:])
	sink.WriteUint64(uint64(29345))

	assert.Equal(t, txid[:], p.CrossChainID)
	assert.Equal(t, txid[:], p.TxHash)
	assert.Equal(t, "unlock", p.Method)
	assert.Equal(t, []byte("btc"), p.FromContractAddress)
	assert.Equal(t, uint64(2), p.ToChainID)
	assert.Equal(t, sink.Bytes(), p.Args)
	assert.Equal(t, scAddr[:], p.ToContractAddress)

	utxos, err := getUtxos(ns, 0, utxoKey)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(utxos.Utxos))
	assert.Equal(t, uint64(29345), utxos.Utxos[0].Value)
	assert.Equal(t, fromBtcTxid+":0", utxos.Utxos[0].Op.String())

	// repeated commit
	sink.Reset()
	params.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), db)
	p, err = handler.MakeDepositProposal(ns)
	assert.Error(t, err)
}

func TestBTCHandler_MultiSign(t *testing.T) {
	rawTx, _ := hex.DecodeString(fromBtcRawTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	_ = mtx.BtcDecode(bytes.NewBuffer(rawTx), wire.ProtocolVersion, wire.LatestEncoding)

	ns := getNativeFunc(nil, nil)
	_ = addUtxos(ns, 0, 0, mtx)

	rb, _ := hex.DecodeString(redeem)
	err := makeBtcTx(ns, 0, map[string]int64{"mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57": 10000}, []byte{123}, 2, rb)
	assert.NoError(t, err)
	stateArr := ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, "makeBtcTx", stateArr[0].(string))
	assert.Equal(t, utxoKey, stateArr[1].(string))

	stxos, err := getStxos(ns, 0, utxoKey)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(stxos.Utxos))
	assert.Equal(t, uint64(29345), stxos.Utxos[0].Value)
	assert.Equal(t, fromBtcTxid+":0", stxos.Utxos[0].Op.String())

	rawTx, _ = hex.DecodeString(stateArr[2].(string))
	_ = mtx.BtcDecode(bytes.NewBuffer(rawTx), wire.ProtocolVersion, wire.LatestEncoding)
	assert.Equal(t, int64(19345), mtx.TxOut[1].Value)

	handler := NewBTCHandler()
	sigArr := getSigs()
	txid := mtx.TxHash()
	// commit no.1 to 4 sig
	for i, sig := range sigArr[:4] {
		msp := ccmcom.MultiSignParam{
			ChainID:   0,
			TxHash:    txid.CloneBytes(),
			Address:   signers[i],
			RedeemKey: utxoKey,
			Signs:     [][]byte{sig},
		}
		sink := common.NewZeroCopySink(nil)
		msp.Serialization(sink)
		ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())

		err = handler.MultiSign(ns)
		assert.NoError(t, err)
	}

	// repeated submit sig4
	msp := ccmcom.MultiSignParam{
		ChainID:   0,
		TxHash:    txid.CloneBytes(),
		Address:   signers[3],
		RedeemKey: utxoKey,
		Signs:     [][]byte{sigArr[3]},
	}
	sink := common.NewZeroCopySink(nil)
	msp.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = handler.MultiSign(ns)
	assert.Error(t, err)

	// right sig but wrong address
	msp = ccmcom.MultiSignParam{
		ChainID:   0,
		TxHash:    txid.CloneBytes(),
		Address:   signers[3],
		RedeemKey: utxoKey,
		Signs:     [][]byte{sigArr[4]},
	}
	sink.Reset()
	msp.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = handler.MultiSign(ns)
	assert.Error(t, err)

	// commit the last right sig
	msp = ccmcom.MultiSignParam{
		ChainID:   0,
		TxHash:    txid.CloneBytes(),
		Address:   signers[4],
		RedeemKey: utxoKey,
		Signs:     [][]byte{sigArr[4]},
	}
	sink.Reset()
	msp.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = handler.MultiSign(ns)
	assert.NoError(t, err)
	stateArr = ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, "btcTxToRelay", stateArr[0].(string))
	assert.Equal(t, hex.EncodeToString([]byte{123}), stateArr[4].(string))

	rawTx, err = hex.DecodeString(stateArr[3].(string))
	assert.NoError(t, err)
	err = mtx.BtcDecode(bytes.NewBuffer(rawTx), wire.ProtocolVersion, wire.LatestEncoding)
	assert.NoError(t, err)

	txid = mtx.TxHash()
	utxos, err = getUtxos(ns, 0, utxoKey)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(utxos.Utxos))
	assert.Equal(t, uint64(19345), utxos.Utxos[0].Value)
	assert.Equal(t, txid.String()+":1", utxos.Utxos[0].Op.String())
}

func syncGenesisHeader(genesisHeader *wire.BlockHeader) (*storage.CacheDB, error) {
	var buf bytes.Buffer
	_ = genesisHeader.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	btcHander := btc.NewBTCHandler()

	sink := new(common.ZeroCopySink)
	params := &hscom.SyncGenesisHeaderParam{
		ChainID:       0,
		GenesisHeader: buf.Bytes(),
	}
	sink = new(common.ZeroCopySink)
	params.Serialization(sink)
	ns := getNativeFunc(sink.Bytes(), nil)
	_ = btcHander.SyncGenesisHeader(ns)

	return ns.GetCacheDB(), nil
}
