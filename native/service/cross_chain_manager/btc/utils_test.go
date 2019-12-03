package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/ontio/multi-chain/native/storage"
	"gotest.tools/assert"

	"sort"
	"testing"
)

var (
	redeem        = "5521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae"
	sig1          = "30440220328fcf07c207b20309c2f42427079592771a1fe63e7196e476c258b32950cc0e022016207f8b39b6af70dd789524cb6bb30927f6e493f798ec29e742b82c119ab2da01"
	sig2          = "3045022100ee671cd934d687ab5f2e23dfe45fe26e40f9618635512c42a32c68836fb29dcd02204ba67c5a27fdc39eb9b2000d3bc68728d6c31433f9d4d7696a7c22194cc4a70301"
	sig3          = "3044022001d3a419341dbd7635a06f8c6ac1baae8cf7095d0b0911ae5b94b925b0b5e63d022037bb90c14f982423c7b2210fe29a78c458ecd8b905915204f3ae1c39bb23da8d01"
	sig4          = "3045022100ee90ef84e7b4dfdd5a2dd68c1fdd19fca53d848f90589da6954978f30398cc21022036d7c1c7659e30755f75c4685eb7f9ffae7c7102aa7400d43de1aa4a821fef2901"
	sig5          = "3044022016bc8e1c55b8b7f2e9ef348fb30e1cc0425c56ca1ba05a7ae0d16cb0484a809802200df54d13632d8435cedcd074935b1a5ddcf8fe87ed1c9f8c88557411b7d9650801"
	sig6          = "3044022042313cb73d49d1b9971e7d2d17bede1352fab46fcdc52770815edc13c46e5a0a022022fa0e208ac51fc1d8fd046a6b0c33e12a8c7bd32d0a4a5cb8378b1d62d26c0c01"
	sig7          = "3044022042c483b95db01dd232a94e21be0449d4ace68c531b89932ea63f3d7be4e34b90022058cb030461755b88531eb7efece2fd9f64a8981187b8ab29ac3cb4d3f39cf77401"
	unsignedTx    = "01000000015ef067df7af576fa5b43bb7e99846c970af7e998cf060c9942920883a515cc6c0000000000ffffffff01401f00000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000"
	sigScript     = "004730440220328fcf07c207b20309c2f42427079592771a1fe63e7196e476c258b32950cc0e022016207f8b39b6af70dd789524cb6bb30927f6e493f798ec29e742b82c119ab2da01483045022100ee671cd934d687ab5f2e23dfe45fe26e40f9618635512c42a32c68836fb29dcd02204ba67c5a27fdc39eb9b2000d3bc68728d6c31433f9d4d7696a7c22194cc4a70301473044022001d3a419341dbd7635a06f8c6ac1baae8cf7095d0b0911ae5b94b925b0b5e63d022037bb90c14f982423c7b2210fe29a78c458ecd8b905915204f3ae1c39bb23da8d01483045022100ee90ef84e7b4dfdd5a2dd68c1fdd19fca53d848f90589da6954978f30398cc21022036d7c1c7659e30755f75c4685eb7f9ffae7c7102aa7400d43de1aa4a821fef2901473044022016bc8e1c55b8b7f2e9ef348fb30e1cc0425c56ca1ba05a7ae0d16cb0484a809802200df54d13632d8435cedcd074935b1a5ddcf8fe87ed1c9f8c88557411b7d96508014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae"

	wsigs = []string{
		"3045022100a5505918f8398492d6e1e3d7b9ec187884a4f43a7443c9d1659c82e781b13a0a02205cba15b80e45feec3d07bdaf5a5042520240b888d533fc2ba03b4e6778ddb74201",
		"3045022100c9c160b0076c43a0fa9c14c67bb33df9a15e971ddcf8a0fe73b487df3c4c856b02202ccf25953602535e2e50987cefb108af91ae18aaa2e2ae7ee5baaba6dc4fb26801",
		"30450221009b235f86ad221171eb56b6d067d7ee831c84ce72cef48d87949d81b40052da2702207b049a7a3b27e93a71a2da41362b67a6a57f3785b9c759d259dc1c894885807201",
		"3045022100df664581c6fa42c24061ae426a67cc068282c67046c4e628e9db9860faa8bac702207c7f8dc11c0785b62f2add90e8bed4fda7066043499211c2b92951bb7d5eb0fa01",
		"304402202aeb76a730767520b06ae0aa6177ae9196a4dd0e244d8f879922e74d9dcec7d502201bfcfa608de19cc0c0e3865279c6fa1aa19a70cd1218631beb6c34140b82a6bc01",
	}
	wTx = "010000000168d852fcfee59bb68304feda29e78e9e5c508ff7fa7abbce3cc448c41da7b9250000000000ffffffff0130d9f505000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"

	getNativeFunc = func() *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, nil, false, nil)
		return service
	}

	getPkSs = func(ty string) [][]byte {
		if ty == "p2sh" {
			p2sh, _ := hex.DecodeString("a91487a9652e9b396545598c0fc72cb5a98848bf93d387") //p2sh
			ss := make([][]byte, 1)
			ss[0] = p2sh
			return ss
		} else if ty == "wit" {
			witlock, _ := hex.DecodeString("002044978a77e4e983136bf1cca277c45e5bd4eff6a7848e900416daf86fd32c2743")
			ss := make([][]byte, 1)
			ss[0] = witlock
			return ss
		}
		return nil
	}
)

func TestVerifySigs(t *testing.T) {
	rs, _ := hex.DecodeString(redeem)
	_, addrs, _, _ := txscript.ExtractPkScriptAddrs(rs, &chaincfg.TestNet3Params)

	sig1b, _ := hex.DecodeString(sig1)
	sigs := [][]byte{sig1b}

	txb, _ := hex.DecodeString(unsignedTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)

	err := verifySigs(sigs, addrs[0].EncodeAddress(), addrs, rs, mtx, getPkSs("p2sh"),[]uint64{})
	if err != nil {
		t.Fatal(err)
	}

	sig2b, _ := hex.DecodeString(sig2)
	sigs = [][]byte{sig2b}
	err = verifySigs(sigs, addrs[0].EncodeAddress(), addrs, rs, mtx, getPkSs("p2sh"),[]uint64{})
	if err == nil {
		t.Fatal("err should not be nil")
	}

	wsig1b, _ := hex.DecodeString(wsigs[0])
	sigs = [][]byte{wsig1b}

	txb, _ = hex.DecodeString(wTx)
	mtx = wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)

	err = verifySigs(sigs, addrs[0].EncodeAddress(), addrs, rs, mtx, getPkSs("wit"), []uint64{btcutil.SatoshiPerBitcoin})
	if err != nil {
		t.Fatal(err)
	}

	wsig2b, _ := hex.DecodeString(wsigs[1])
	sigs = [][]byte{wsig2b}
	err = verifySigs(sigs, addrs[0].EncodeAddress(), addrs, rs, mtx, getPkSs("wit"), []uint64{btcutil.SatoshiPerBitcoin})
	if err == nil {
		t.Fatalf("err should not be nil")
	}

	err = verifySigs(sigs, addrs[1].EncodeAddress(), addrs, rs, mtx, getPkSs("wit"), []uint64{1000})
	if err == nil {
		t.Fatalf("err should not be nil")
	}
}

func TestAddSigToTx(t *testing.T) {
	sigArr := make([][]byte, 0)
	sig1b, _ := hex.DecodeString(sig1)
	sig2b, _ := hex.DecodeString(sig2)
	sig3b, _ := hex.DecodeString(sig3)
	sig4b, _ := hex.DecodeString(sig4)
	sig5b, _ := hex.DecodeString(sig5)
	sig6b, _ := hex.DecodeString(sig6)
	sig7b, _ := hex.DecodeString(sig7)
	sigArr = append(sigArr, sig1b, sig2b, sig3b, sig4b, sig5b, sig6b, sig7b)

	rs, _ := hex.DecodeString(redeem)
	_, addrs, _, _ := txscript.ExtractPkScriptAddrs(rs, &chaincfg.TestNet3Params)
	sigMap := new(MultiSignInfo)

	sigMap.MultiSignInfo = make(map[string][][]byte)
	for i, addr := range addrs {
		if i == 5 {
			break
		}
		sigMap.MultiSignInfo[addr.EncodeAddress()] = [][]byte{sigArr[i]}
	}

	txb, _ := hex.DecodeString(unsignedTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)

	err := addSigToTx(sigMap, addrs, rs, mtx, getPkSs("p2sh"))
	if err != nil {
		t.Fatal(err)
	}
	s, _ := hex.DecodeString(sigScript)

	if !bytes.Equal(mtx.TxIn[0].SignatureScript, s) {
		t.Fatal("not equal")
	}

	wsigArr := make([][]byte, 0)
	for _, s := range wsigs {
		sb, _ := hex.DecodeString(s)
		wsigArr = append(wsigArr, sb)
	}

	sigMap = new(MultiSignInfo)
	sigMap.MultiSignInfo = make(map[string][][]byte)
	for i, addr := range addrs {
		if i == 5 {
			break
		}
		sigMap.MultiSignInfo[addr.EncodeAddress()] = [][]byte{wsigArr[i]}
	}

	txb, _ = hex.DecodeString(wTx)
	mtx = wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)
	err = addSigToTx(sigMap, addrs, rs, mtx, getPkSs("wit"))
	if err != nil {
		t.Fatal(err)
	}

	vm, err := txscript.NewEngine(getPkSs("wit")[0], mtx, 0, txscript.StandardVerifyFlags, nil,
		nil, btcutil.SatoshiPerBitcoin)
	if err != nil {
		t.Fatal(err)
	}
	err = vm.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestEstimateSerializedTxSize(t *testing.T) {
	str := "010000000311fea08426d1ab156c894fced9e7ddc795d625891d587d81bf011fe7e7f7437200000000fd5e0200483045022100823e0f0d14a297a5c7cb590d0ea216469273bf12fb5e4c437ac3eec84f7e1e3c02202a356cac0a54fe336b75c74702e7fb1f759336f372b3e743f2603581ee5902ec01483045022100f86d265859bb3f4619d9f32d91705b19e8e94df3734d81cca05b04ffb3927829022015ff7f88e635b59f08ee11c5c967a4de9578f4136b3996307427f3dea05993ae01473044022066301d1e335554e828d7d70cd6df89352acc475a7bbf17e16b87e1ac18ca2141022019b219423f415c848291f07eb68f6bf2e12d79ecaf95e7b90656eb09969fec250147304402202fc13748831c2c998a9f33bb7482afb8f6e3b89191e0e24e18b82d353b66b5cf02202ac77c4b048a375b122b1c0601975eaf8dbf6fe1b1913605f95a96e07c0a333701473044022058327e2e6e3d69362ff27f034949390ed2669baf64705c2986057484d23db0a502207d5f4a34b5bfb17edbf76d999d013d92f5382ee7964f06f1711d291016a1f8c7014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffff89d3036606437e789a81a9518b4d6384b2e43987a93fd6030cd289ad1508a5e200000000fd600200483045022100ad76ae04baee7d4d6ce30968f5140e76a72b2683c524b47efc3a20ff10ae14450220066e1f62d0df58fcf2274e2d37d5f3d20bff7cf2d96f84b5dccbd18d8a1ab31001483045022100e02f15541560e606bb00b3487e8421f440d1a709f2f0896ef1bd61736f4fc5f402204a68b4c850e2efdd39c08f1db2cdb9aff87da78503e39dc2e2cf076715da908e01483045022100e348cbeea6f20cab4e02fa90639809ddb259b45090bdd3a46796347ee7a8396802206f4a826af2f7af369c986c4539d6fdd9bcad38113e7449fd4816ee6605f2943c014730440220751574412e392ac8cb88a7fa47931bfe4cf3542771ce22f95d1b5d38764e7ae20220523b9273b294f619eedd920f70cba8724bd9e7bf3192abf731e7c4706f2fc50001483045022100d57eaa1a500620543a3f3d553d5613d29df9771ec9bc1a042c20a925bf1bfdb102206c103fb167cb52e23e1fa23d8e2aec6b67ed2d13d6c74bd77014b10d93e22f9e014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffffb585e37d38bb275278e1622b66cf53259ca33e08929e4f16f6439e0cdc7fdb7500000000fd5e0200483045022100f2a3422523ff6c33c9e10188d2b62503ab7dfdde03118f52140be4393ba0c86502203afafde7c3c0d620db1732cbc89ee6088294ad10a0c33fc9e5d283ba1bed8e9c0147304402203734f5d065c703541a240ef3a350ba3e1c0d566f81c59f358887d9877cc7cdd70220746adbb82be715b331e197e5bb33d929298b99784f4739edb82b7ff0858a79930147304402206d524931764e0fc89962c5828fd35c4389a7ffff5b335b1d66eee08ba9fff7e102203f9182b03a1d111506602f481c86c2a6a337cb33f313af8b31e4434025de36b1014730440220134063af81d4fba15f721289ae57e370d65d6a7a7146076a828e6ccf53249404022036485246954d7b04992c65e0f58307117391347199eee287537eea4a18ca7b6101483045022100ba8a32796185819649911e5e530522af06c1350e22965104278980f5625ffed8022000fb054e21004140e7be4a7caf8f7c756b4a98136c2be79440a03b438681b0f4014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffff0246010000000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188acae0000000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000"
	mtx := wire.NewMsgTx(wire.TxVersion)
	txb, _ := hex.DecodeString(str)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.ProtocolVersion, wire.LatestEncoding)
	es := estimateSerializedTxSize(len(mtx.TxIn), mtx.TxOut[:len(mtx.TxOut)-1], mtx.TxOut[len(mtx.TxOut)-1])
	fmt.Println(len(txb), es, mtx.TxOut[1].Value)
	assert.Assert(t, es >= len(txb), "size not right")
}

func TestBTCHandler_MakeTransaction(t *testing.T) {
	str := "010000000311fea08426d1ab156c894fced9e7ddc795d625891d587d81bf011fe7e7f7437200000000fd5e0200483045022100823e0f0d14a297a5c7cb590d0ea216469273bf12fb5e4c437ac3eec84f7e1e3c02202a356cac0a54fe336b75c74702e7fb1f759336f372b3e743f2603581ee5902ec01483045022100f86d265859bb3f4619d9f32d91705b19e8e94df3734d81cca05b04ffb3927829022015ff7f88e635b59f08ee11c5c967a4de9578f4136b3996307427f3dea05993ae01473044022066301d1e335554e828d7d70cd6df89352acc475a7bbf17e16b87e1ac18ca2141022019b219423f415c848291f07eb68f6bf2e12d79ecaf95e7b90656eb09969fec250147304402202fc13748831c2c998a9f33bb7482afb8f6e3b89191e0e24e18b82d353b66b5cf02202ac77c4b048a375b122b1c0601975eaf8dbf6fe1b1913605f95a96e07c0a333701473044022058327e2e6e3d69362ff27f034949390ed2669baf64705c2986057484d23db0a502207d5f4a34b5bfb17edbf76d999d013d92f5382ee7964f06f1711d291016a1f8c7014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffff89d3036606437e789a81a9518b4d6384b2e43987a93fd6030cd289ad1508a5e200000000fd600200483045022100ad76ae04baee7d4d6ce30968f5140e76a72b2683c524b47efc3a20ff10ae14450220066e1f62d0df58fcf2274e2d37d5f3d20bff7cf2d96f84b5dccbd18d8a1ab31001483045022100e02f15541560e606bb00b3487e8421f440d1a709f2f0896ef1bd61736f4fc5f402204a68b4c850e2efdd39c08f1db2cdb9aff87da78503e39dc2e2cf076715da908e01483045022100e348cbeea6f20cab4e02fa90639809ddb259b45090bdd3a46796347ee7a8396802206f4a826af2f7af369c986c4539d6fdd9bcad38113e7449fd4816ee6605f2943c014730440220751574412e392ac8cb88a7fa47931bfe4cf3542771ce22f95d1b5d38764e7ae20220523b9273b294f619eedd920f70cba8724bd9e7bf3192abf731e7c4706f2fc50001483045022100d57eaa1a500620543a3f3d553d5613d29df9771ec9bc1a042c20a925bf1bfdb102206c103fb167cb52e23e1fa23d8e2aec6b67ed2d13d6c74bd77014b10d93e22f9e014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffffb585e37d38bb275278e1622b66cf53259ca33e08929e4f16f6439e0cdc7fdb7500000000fd5e0200483045022100f2a3422523ff6c33c9e10188d2b62503ab7dfdde03118f52140be4393ba0c86502203afafde7c3c0d620db1732cbc89ee6088294ad10a0c33fc9e5d283ba1bed8e9c0147304402203734f5d065c703541a240ef3a350ba3e1c0d566f81c59f358887d9877cc7cdd70220746adbb82be715b331e197e5bb33d929298b99784f4739edb82b7ff0858a79930147304402206d524931764e0fc89962c5828fd35c4389a7ffff5b335b1d66eee08ba9fff7e102203f9182b03a1d111506602f481c86c2a6a337cb33f313af8b31e4434025de36b1014730440220134063af81d4fba15f721289ae57e370d65d6a7a7146076a828e6ccf53249404022036485246954d7b04992c65e0f58307117391347199eee287537eea4a18ca7b6101483045022100ba8a32796185819649911e5e530522af06c1350e22965104278980f5625ffed8022000fb054e21004140e7be4a7caf8f7c756b4a98136c2be79440a03b438681b0f4014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffff0246010000000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188acae0000000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000"
	mtx := wire.NewMsgTx(wire.TxVersion)
	txb, _ := hex.DecodeString(str)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.ProtocolVersion, wire.LatestEncoding)
	txid := mtx.TxHash()

	ns := getNativeFunc()
	utxos := Utxos{
		Utxos: []*Utxo{&Utxo{
			Op:           &OutPoint{Hash: txid[:], Index: 1},
			Value:        3000,
			ScriptPubkey: mtx.TxOut[1].PkScript,
			AtHeight:     0,
		}},
	}
	putUtxos(ns, 0, &utxos)
	putBtcRedeemScript(ns, redeem)
	amounts := make(map[string]int64)
	amounts["mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57"] = 2500
	err := makeBtcTx(ns, 0, amounts, []byte("123"), 2)
	if err != nil {
		t.Fatalf("failed to make btc tx: %v", err)
	}

	notify := ns.GetNotify()
	if notify[0].ContractAddress.ToHexString() != utils.CrossChainManagerContractAddress.ToHexString() {
		t.Fatal("wrong contract address")
	}
	states := notify[0].States.([]interface{})
	key := states[0].(string)
	if key != "makeBtcTx" {
		t.Fatal("wrong key")
	}

	rawTx := states[1].(string)
	rawTxb, _ := hex.DecodeString(rawTx)
	mtx.BtcDecode(bytes.NewBuffer(rawTxb), wire.ProtocolVersion, wire.LatestEncoding)
	if mtx.TxIn[0].PreviousOutPoint.Hash.String() != txid.String() {
		t.Fatal("wrong outpoint")
	}

	fmt.Printf("txsize %d, out0 %d, out1 %d, fee %d", len(rawTxb), mtx.TxOut[0].Value, mtx.TxOut[1].Value, 2500-mtx.TxOut[0].Value)
}

func TestUtxos_Sort(t *testing.T) {
	utxos := &Utxos{
		Utxos: []*Utxo {
			{ // 1000
				Value: 10,
				Confs: 100,
			},
			{ // 750
				Value: 5,
				Confs: 150,
			},
			{ // 1500
				Value: 30,
				Confs: 50,
			},
		},
	}

	sort.Sort(utxos)
	vals := []uint64{5, 10, 30}
	for i, u := range utxos.Utxos {
		//fmt.Println(u.Value, vals[i])
		assert.Assert(t, u.Value == vals[i], "not equal %d", i)
	}
}

func TestUtxos_Choose(t *testing.T) {
	ns := getNativeFunc()
	utxos := &Utxos{
		Utxos: []*Utxo {
			{ // 10000000
				Value: 1e5,
				AtHeight: 100,
				Op: &OutPoint{
					Hash: []byte("123"),
					Index: 0,
				},
				ScriptPubkey: []byte("1"),
			},
			{ // 7500000
				Value: 5e4,
				AtHeight: 50,
				Op: &OutPoint{
					Hash: []byte("123"),
					Index: 0,
				},
				ScriptPubkey: []byte("2"),
			},
			{ // 15000000
				Value: 3e5,
				AtHeight: 150,
				Op: &OutPoint{
					Hash: []byte("123"),
					Index: 0,
				},
				ScriptPubkey: []byte("3"),
			},
		},
	}
	putUtxos(ns, 0, utxos)

	set, sum, err := chooseUtxos(ns, 0, 35e4)
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, sum == 4e5 && len(set) == 2 && bytes.Equal(set[0].ScriptPubkey, []byte("3")) &&
		bytes.Equal(set[1].ScriptPubkey, []byte("1")), "wrong sum")

	_, _, err = chooseUtxos(ns, 0, 100e4)
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

