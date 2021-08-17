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

package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	ccmcom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/btc"
	hscom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
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
		"3045022100f77b28268bfed3c0ddc8d35e556164d7ce2f571715b5afbf79373e699f4023e102200f86442f305528088caca96453d7f4fd782c831198c51e7637a7024ca844e10d01",
		"30450221009a320292de5b0881f2988b603d7dac98c666c9653c177e1add6e26b2ae7902860220211d58d5e261e9ce80fa1b12892f02eb1a10c04712fe1609464629a794f5eaa001",
		"30440220269ba3ab6e8d2513bb908e13c73b89fad28743136e56434bce7f9467c92a70cc02207149b2ae0b8648d5048d2505464582972f12522b36f1517af16af5408c0ed01a01",
		"3044022046e9fa4e00ace70f5f1136a97257144cb83947d2685cfb6167fb3d61e563896802202f085fd5a202b1ccd471b4341effb788e36a5beed9b1bed4865e2c9a8a5f86e501",
		"3045022100b457c88a42276924eafbb07e48f642410a07d72516739ab077386088e31ed6dc0220679a39f6aef46d980d031a0df301d74ee750d82ac2c8e0951d2ab1d8444346ec01",
	}

	signers = []string{
		"mxtJn3aRsKLrWRLLAhu2nCBuK2brfazedj",
		"msu1qgtn4FsQh7xDP15ggStwW4yHquUTYE",
		"mzr5T4PmqzNtusmM2S889LUySC9BBbwRzd",
		"mmDSSyis1sjysaCKZ9eK1ww92Vr9S1CNEX",
		"n1WmUbJ4dQfvzvtRaNcjmxH6nADhhv7W8c",
	}

	getNativeFunc = func(args []byte, db *storage.CacheDB) *native.NativeService {
		if db == nil {
			store, _ := leveldbstore.NewMemLevelDBStore()
			db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		}

		service, _ := native.NewNativeService(db, &types.Transaction{ChainID: 0}, 0, 0, common.Uint256{0}, 0, args, false)
		return service
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
		cb := &side_chain_manager.ContractBinded{
			Contract: ca,
			Ver:      0,
		}
		sink := common.NewZeroCopySink(nil)
		cb.Serialization(sink)
		rk, _ := hex.DecodeString(utxoKey)
		redeem, _ := hex.DecodeString(rdm)
		db.Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.REDEEM_BIND),
			utils.GetUint64Bytes(1), utils.GetUint64Bytes(2), rk), states.GenRawStorageItem(sink.Bytes()))
		db.Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.REDEEM_SCRIPT),
			utils.GetUint64Bytes(1), []byte(utxoKey)), states.GenRawStorageItem(redeem))
		return db
	}

	setBtcTxParam = func(db *storage.CacheDB, redeemK string) *storage.CacheDB {
		detail := &side_chain_manager.BtcTxParamDetial{
			FeeRate:   2,
			MinChange: 2000,
		}
		sink := common.NewZeroCopySink(nil)
		detail.Serialization(sink)
		rk, _ := hex.DecodeString(redeemK)
		db.Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.BTC_TX_PARAM), rk,
			utils.GetUint64Bytes(1)), states.GenRawStorageItem(sink.Bytes()))
		return db
	}
)

func TestBTCHandler_MakeDepositProposal(t *testing.T) {
	gh := netParam.GenesisBlock.Header
	mr, _ := chainhash.NewHashFromStr("502e1d655973488e2394b56865f46cf204e5e2fdd0ea5873c51c65a3125ab3dd")
	gh.MerkleRoot = *mr
	db, err := syncGenesisHeader(&gh)
	if err != nil {
		t.Fatal(err)
	}
	db = registerRC(db)

	txid, _ := chainhash.NewHashFromStr("67cb330dc68d90a376444a6c8b3e37445050453e72ca43305874daff4b6c51d0")
	rawTx, _ := hex.DecodeString("01000000015dbdab5a45905efd23e0753d1aaf2a417d77dd8c079499a1643bc168817bf8ab4f0000006a47304402206553c4a3cb1c37cd68b4bb25412cc35d73b731dcef3874635761172f53d70bbf0220264e5afd78936a920d6bcc0720ef5f5d25e7a153f25e18264bd3952038780224012102141d092eca49eac51de2760d28cbced212b60efc23fdcbb57304823bb17aa64effffffff031027000000000000220020216a09cb8ee51da1a91ea8942552d7936c886a10b507299003661816c0e9f18b0000000000000000286a26cc02000000000000000000000000000000145cd3143f91a13fe971043e1e4605c1c23b46bf44a85b0100000000001976a9145f35a2cc0318fbc17c4c479964734e7a9f8819d788ac00000000")
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
	proof, _ = hex.DecodeString("0000002037083b799b61659dedf733d4945e4ce65e31018ca7e1c2a247f0120000000000ddb35a12a3651cc57358ead0fde2e504f26cf46568b594238e487359651d2e5060d5715effff001d74ec61d6370100000a4702e34d13d88ca00bcea9e15428040de063fd3772fb0492b46bc9ac734612f7d1f8a7ffd7d1f965cad52b3ec06efa3e49e20344de6463d7688453050a37b52b09a2a2efe3057dca55982d5f7ff3b1f36fda89d2b2a1f015acd3ce7afda0abfe96662da89072ef81d5795add6f50dee212a41dbdd2720a1d8c53520bed8e7fa8732bbc20668e26657be4de157fe22cbb508e6e92030bf97b75298db89026f027d0516c4bffda74583043ca723e45505044373e8b6c4a4476a3908dc60d33cb6721b2bc97b3e2074d2ab6617ad3204fec91130fe06e5736ac9d07f66caee0c05309d5d8e752dadbe4c365f815e1902f6ce80be7269f296cb49bfd832c243dd4580dcba943ed5b67f8d233d19b6402fcc39e61bfe01938dc98e4dd2043efed8dabbd65df34229b60bd0a0afd0823ef8c8055cd52d1737d3a991575a6a41cbaeb1e03b75a00")
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
	assert.Equal(t, "67cb330dc68d90a376444a6c8b3e37445050453e72ca43305874daff4b6c51d0:0", utxos.Utxos[0].Op.String())

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
	setSideChain(ns)
	registerRC(ns.GetCacheDB())
	setBtcTxParam(ns.GetCacheDB(), utxoKey)

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
	_ = addUtxos(ns, 1, 0, mtx)
	setBtcTxParam(ns.GetCacheDB(), utxoKey)
	registerRC(ns.GetCacheDB())

	rb, _ := hex.DecodeString(rdm)
	err := makeBtcTx(ns, 1, map[string]int64{"mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57": 6000}, []byte{123},
		2, rb, btcutil.Hash160(rb))
	assert.NoError(t, err)
	stateArr := ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, "makeBtcTx", stateArr[0].(string))
	assert.Equal(t, utxoKey, stateArr[1].(string))

	stxos, err := getStxos(ns, 1, utxoKey)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(stxos.Utxos))
	assert.Equal(t, uint64(10000), stxos.Utxos[0].Value)
	assert.Equal(t, fromBtcTxid+":0", stxos.Utxos[0].Op.String())

	rawTx, _ = hex.DecodeString(stateArr[2].(string))
	_ = mtx.BtcDecode(bytes.NewBuffer(rawTx), wire.ProtocolVersion, wire.LatestEncoding)
	assert.Equal(t, int64(4000), mtx.TxOut[1].Value)
	handler := NewBTCHandler()
	sigArr := getSigs()
	txid := mtx.TxHash()
	// commit no.1 to 4 sig
	for i, sig := range sigArr[:4] {
		msp := ccmcom.MultiSignParam{
			ChainID:   1,
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
		ChainID:   1,
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
		ChainID:   1,
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
		ChainID:   1,
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
	utxos, err = getUtxos(ns, 1, utxoKey)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(utxos.Utxos))
	assert.Equal(t, uint64(4000), utxos.Utxos[0].Value)
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
