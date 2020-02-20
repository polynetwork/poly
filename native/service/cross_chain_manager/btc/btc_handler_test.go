package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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
	"strings"
	"testing"
)

var (
	acct *account.Account = account.NewAccount("")

	rdm               = "552102dec9a415b6384ec0a9331d0cdf02020f0f1e5731c327b86e2b5a92455a289748210365b1066bcfa21987c3e207b92e309b95ca6bee5f1133cf04d6ed4ed265eafdbc21031104e387cd1a103c27fdc8a52d5c68dec25ddfb2f574fbdca405edfd8c5187de21031fdb4b44a9f20883aff505009ebc18702774c105cb04b1eecebcb294d404b1cb210387cda955196cc2b2fc0adbbbac1776f8de77b563c6d2a06a77d96457dc3d0d1f2102dd7767b6a7cc83693343ba721e0f5f4c7b4b8d85eeb7aec20d227625ec0f59d321034ad129efdab75061e8d4def08f5911495af2dae6d3e9a4b6e7aeb5186fa432fc57ae"
	fromBtcTxid       = "2587a59e8069c563d32de9d4a2b946760d740b6963566dd7b32d8ec549f2d238"
	fromBtcRawTx      = "010000000147d9b1bc6a52099f746863722282e3febc9ad3ad6b2eac0f2df6d2badf1df28a020000006b483045022100a1e573ba3589217e1b20d6ed53e2dda705deb3d284122c61987266e66aff074802200165734cf4519b560d806d392f10cec2aeb3071cf72c759a5abc9c33cd2f983f012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff031027000000000000220020216a09cb8ee51da1a91ea8942552d7936c886a10b507299003661816c0e9f18b00000000000000003d6a3b6602000000000000000000000000000000149702640a6b971ca18efc20ad73ca4e8ba390c910145cd3143f91a13fe971043e1e4605c1c23b46bf44620e0700000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"
	fromBtcProof      = "00000020775635e1ada1581f0fa6eff86bfc4720253c9c4fcd7165843e902600000000003faec6ef7165e988b344b553b15dff0d66eb62e71b1d93462c64b0eab1086852fef54c5effff001d6c2edd4f4d0200000b3a4a5328d2e6b72f26fb5f3aa6db80e8301c2746c5ce6e21813e884c3a08e96a8ba1ccfe764700b7d956acff0697680b0e9412517972d8e8a10c9ac37c96fd0c81a705037d9f8caaa679075d525cd12bbb698e6f6917e61aecbe3d529f65c7bd38d2f249c58e2db3d76d5663690b740d7646b9a2d4e92dd363c569809ea58725a47e736292f4a96de7c46462c53b823c1732cb2d863c402bc3dd96527e69305e0af15f3487c9093c59d7dc0e7fcde6db50354e73f640987e3305e917aad7531abaa9513d16228fb2b17c3cd04f9ec97e3c38de9dac7ff2af93184c338e86e6c2d8731bc8430a7f31bc050d11776d6e3b665951af070fe889cba7aec895e40b3e67107e62b1ec0ebef9a226abc458b55920060f0a5c06edd26432a987d3f6cfafefee3301b3281270ceb45e5831e435fa70056cd28927251eebe875f2fd810aa501cca43940fe1ba0e8be004e8f05f740b66e2e9a1a24b76bdd4c0c6d53230c1903ef2d00"
	fromBtcMerkleRoot = "526808b1eab0642c46931d1be762eb660dff5db153b544b388e96571efc6ae3f"
	toOntAddr         = "AdzZ2VKufdJWeB8t9a8biXoHbbMe2kZeyH"
	obtcxAddr         = "3e6d9288d04d49585699659aadf3b0a508c47608"
	utxoKey           = "c330431496364497d7257839737b5e4596f5ac06" //"87a9652e9b396545598c0fc72cb5a98848bf93d3"
	toEthAddr         = "0x5cD3143f91a13Fe971043E1e4605C1c23b46bF44"
	ebtcxAddr         = "0x9702640a6b971CA18EFC20AD73CA4e8bA390C910"

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
			ChainId:      1,
			BlocksToWait: 1,
			Router:       0,
		}
		sink := common.NewZeroCopySink(nil)
		_ = side.Serialization(sink)

		ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress,
			[]byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(1)), states.GenRawStorageItem(sink.Bytes()))
	}

	getSigs = func() [][]byte {
		res := make([][]byte, len(sigs))
		for i, sig := range sigs {
			raw, _ := hex.DecodeString(sig)
			res[i] = raw
		}
		return res
	}

	registerRC = func(db *storage.CacheDB) *storage.CacheDB {
		ca, _ := hex.DecodeString(strings.Replace(ebtcxAddr, "0x", "", 1))
		rk, _ := hex.DecodeString(utxoKey)
		db.Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.REDEEM_BIND),
			utils.GetUint64Bytes(1), utils.GetUint64Bytes(2), rk), states.GenRawStorageItem(ca))

		return db
	}
)

func TestBTCHandler_MakeDepositProposal(t *testing.T) {
	gh := netParam.GenesisBlock.Header
	mr, _ := chainhash.NewHashFromStr(fromBtcMerkleRoot)
	gh.MerkleRoot = *mr

	db, err := syncGenesisHeader(&gh)
	if err != nil {
		t.Fatal(err)
	}
	db = registerRC(db)

	txid, _ := chainhash.NewHashFromStr(fromBtcTxid)
	rawTx, _ := hex.DecodeString(fromBtcRawTx)
	scAddr, _ := hex.DecodeString(strings.Replace(ebtcxAddr, "0x", "", 1))
	handler := NewBTCHandler()

	// wrong proof
	wrongProof := "0100003037db655b09de3449fe60bc0838ef3541e28d3ae31a05093f1bb63e4845a6b102695fd2a687fc1fc368f13227c2bb1b6b0fcac9760936d869a9ba01f8a75f825c5105245effff7f20050000000200000002e0c8d9fb711dd377d0ba8d1c16c154b903432aa8923c87f3f0fd6045be7b8c8a51cf2962a492309e6bd7aa56848b2817c1a760f9fb0823762200d2b286f988b90105"
	proof, _ := hex.DecodeString(wrongProof)
	params := new(ccmcom.EntranceParam)
	params.Height = 0
	params.SourceChainID = 1
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

	ethAddr, _ := hex.DecodeString(strings.Replace(toEthAddr, "0x", "", 1))
	sink.Reset()
	sink.WriteVarBytes(ethAddr[:])
	sink.WriteUint64(uint64(10000))

	assert.Equal(t, txid[:], p.CrossChainID)
	assert.Equal(t, txid[:], p.TxHash)
	assert.Equal(t, "unlock", p.Method)
	assert.Equal(t, utxoKey, hex.EncodeToString(p.FromContractAddress))
	assert.Equal(t, uint64(2), p.ToChainID)
	assert.Equal(t, sink.Bytes(), p.Args)
	assert.Equal(t, scAddr[:], p.ToContractAddress)

	utxos, err := getUtxos(ns, 1, utxoKey)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(utxos.Utxos))
	assert.Equal(t, uint64(10000), utxos.Utxos[0].Value)
	assert.Equal(t, fromBtcTxid+":0", utxos.Utxos[0].Op.String())

	// repeated commit
	sink.Reset()
	params.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), db)
	p, err = handler.MakeDepositProposal(ns)
	assert.Error(t, err)
}

func TestBTCHandler_MakeTransaction(t *testing.T) {
	rawTx, _ := hex.DecodeString(fromBtcRawTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	_ = mtx.BtcDecode(bytes.NewBuffer(rawTx), wire.ProtocolVersion, wire.LatestEncoding)

	ns := getNativeFunc(nil, nil)
	_ = addUtxos(ns, 1, 0, mtx)
	registerRC(ns.GetCacheDB())
	setSideChain(ns)

	rk, _ := hex.DecodeString(utxoKey)
	scAddr, _ := hex.DecodeString(strings.Replace(ebtcxAddr, "0x", "", 1))
	r, _ := hex.DecodeString(rdm)

	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte("mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57"))
	sink.WriteUint64(6000)
	sink.WriteVarBytes(r)
	p := &ccmcom.MakeTxParam{
		ToChainID:           1,
		TxHash:              []byte{1},
		Method:              "unlock",
		ToContractAddress:   rk,
		CrossChainID:        []byte{1},
		FromContractAddress: scAddr,
		Args:                sink.Bytes(),
	}

	handler := NewBTCHandler()
	err := handler.MakeTransaction(ns, p, 2)
	assert.NoError(t, err)
	s := ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, utxoKey, s[1].(string))
}

func TestBTCHandler_MultiSign(t *testing.T) {
	rawTx, _ := hex.DecodeString(fromBtcRawTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	_ = mtx.BtcDecode(bytes.NewBuffer(rawTx), wire.ProtocolVersion, wire.LatestEncoding)

	ns := getNativeFunc(nil, nil)
	_ = addUtxos(ns, 0, 0, mtx)

	rb, _ := hex.DecodeString(rdm)
	err := makeBtcTx(ns, 0, map[string]int64{"mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57": 6000}, []byte{123},
		2, rb, hex.EncodeToString(btcutil.Hash160(rb)))
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

	h := make([]byte, 4)
	binary.BigEndian.PutUint32(h, 0)
	params := &hscom.SyncGenesisHeaderParam{
		ChainID:       1,
		GenesisHeader: append(buf.Bytes(), h...),
	}
	sink = new(common.ZeroCopySink)
	params.Serialization(sink)

	ns := getNativeFunc(sink.Bytes(), nil)
	err := btcHander.SyncGenesisHeader(ns)
	if err != nil {
		return nil, err
	}

	return ns.GetCacheDB(), nil
}
