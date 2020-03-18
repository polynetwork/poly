package btc

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"sort"
	"testing"
)

var (
	redeem     = "5521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae"
	sig1       = "30440220328fcf07c207b20309c2f42427079592771a1fe63e7196e476c258b32950cc0e022016207f8b39b6af70dd789524cb6bb30927f6e493f798ec29e742b82c119ab2da01"
	sig2       = "3045022100ee671cd934d687ab5f2e23dfe45fe26e40f9618635512c42a32c68836fb29dcd02204ba67c5a27fdc39eb9b2000d3bc68728d6c31433f9d4d7696a7c22194cc4a70301"
	sig3       = "3044022001d3a419341dbd7635a06f8c6ac1baae8cf7095d0b0911ae5b94b925b0b5e63d022037bb90c14f982423c7b2210fe29a78c458ecd8b905915204f3ae1c39bb23da8d01"
	sig4       = "3045022100ee90ef84e7b4dfdd5a2dd68c1fdd19fca53d848f90589da6954978f30398cc21022036d7c1c7659e30755f75c4685eb7f9ffae7c7102aa7400d43de1aa4a821fef2901"
	sig5       = "3044022016bc8e1c55b8b7f2e9ef348fb30e1cc0425c56ca1ba05a7ae0d16cb0484a809802200df54d13632d8435cedcd074935b1a5ddcf8fe87ed1c9f8c88557411b7d9650801"
	sig6       = "3044022042313cb73d49d1b9971e7d2d17bede1352fab46fcdc52770815edc13c46e5a0a022022fa0e208ac51fc1d8fd046a6b0c33e12a8c7bd32d0a4a5cb8378b1d62d26c0c01"
	sig7       = "3044022042c483b95db01dd232a94e21be0449d4ace68c531b89932ea63f3d7be4e34b90022058cb030461755b88531eb7efece2fd9f64a8981187b8ab29ac3cb4d3f39cf77401"
	unsignedTx = "01000000015ef067df7af576fa5b43bb7e99846c970af7e998cf060c9942920883a515cc6c0000000000ffffffff01401f00000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000"
	sigScript  = "004730440220328fcf07c207b20309c2f42427079592771a1fe63e7196e476c258b32950cc0e022016207f8b39b6af70dd789524cb6bb30927f6e493f798ec29e742b82c119ab2da01483045022100ee671cd934d687ab5f2e23dfe45fe26e40f9618635512c42a32c68836fb29dcd02204ba67c5a27fdc39eb9b2000d3bc68728d6c31433f9d4d7696a7c22194cc4a70301473044022001d3a419341dbd7635a06f8c6ac1baae8cf7095d0b0911ae5b94b925b0b5e63d022037bb90c14f982423c7b2210fe29a78c458ecd8b905915204f3ae1c39bb23da8d01483045022100ee90ef84e7b4dfdd5a2dd68c1fdd19fca53d848f90589da6954978f30398cc21022036d7c1c7659e30755f75c4685eb7f9ffae7c7102aa7400d43de1aa4a821fef2901473044022016bc8e1c55b8b7f2e9ef348fb30e1cc0425c56ca1ba05a7ae0d16cb0484a809802200df54d13632d8435cedcd074935b1a5ddcf8fe87ed1c9f8c88557411b7d96508014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae"

	wsigs = []string{
		"3045022100a5505918f8398492d6e1e3d7b9ec187884a4f43a7443c9d1659c82e781b13a0a02205cba15b80e45feec3d07bdaf5a5042520240b888d533fc2ba03b4e6778ddb74201",
		"3045022100c9c160b0076c43a0fa9c14c67bb33df9a15e971ddcf8a0fe73b487df3c4c856b02202ccf25953602535e2e50987cefb108af91ae18aaa2e2ae7ee5baaba6dc4fb26801",
		"30450221009b235f86ad221171eb56b6d067d7ee831c84ce72cef48d87949d81b40052da2702207b049a7a3b27e93a71a2da41362b67a6a57f3785b9c759d259dc1c894885807201",
		"3045022100df664581c6fa42c24061ae426a67cc068282c67046c4e628e9db9860faa8bac702207c7f8dc11c0785b62f2add90e8bed4fda7066043499211c2b92951bb7d5eb0fa01",
		"304402202aeb76a730767520b06ae0aa6177ae9196a4dd0e244d8f879922e74d9dcec7d502201bfcfa608de19cc0c0e3865279c6fa1aa19a70cd1218631beb6c34140b82a6bc01",
	}
	wTx = "010000000168d852fcfee59bb68304feda29e78e9e5c508ff7fa7abbce3cc448c41da7b9250000000000ffffffff0130d9f505000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"

	witPubScript, _ = hex.DecodeString("002044978a77e4e983136bf1cca277c45e5bd4eff6a7848e900416daf86fd32c2743")
	p2sh, _         = hex.DecodeString("a91487a9652e9b396545598c0fc72cb5a98848bf93d387")

	utxos = &Utxos{
		Utxos: []*Utxo{
			{ // 10000000
				Value:    1e5,
				AtHeight: 1,
				Op: &OutPoint{
					Hash:  []byte("123"),
					Index: 0,
				},
				ScriptPubkey: witPubScript,
			},
			{ // 7500000
				Value:    5e4,
				AtHeight: 2,
				Op: &OutPoint{
					Hash:  []byte("123"),
					Index: 0,
				},
				ScriptPubkey: p2sh,
			},
			{ // 15000000
				Value:    3e5,
				AtHeight: 3,
				Op: &OutPoint{
					Hash:  []byte("123"),
					Index: 0,
				},
				ScriptPubkey: witPubScript,
			},
		},
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

	err := verifySigs(sigs, addrs[0].EncodeAddress(), addrs, rs, mtx, getPkSs("p2sh"), []uint64{})
	if err != nil {
		t.Fatal(err)
	}

	sig2b, _ := hex.DecodeString(sig2)
	sigs = [][]byte{sig2b}
	err = verifySigs(sigs, addrs[0].EncodeAddress(), addrs, rs, mtx, getPkSs("p2sh"), []uint64{})
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

func TestUtxos_Sort(t *testing.T) {
	utxos := &Utxos{
		Utxos: []*Utxo{
			{ // 1000
				Value: 10,
			},
			{ // 750
				Value: 5,
			},
			{ // 1500
				Value: 30,
			},
		},
	}

	sort.Sort(utxos)
	vals := []uint64{5, 10, 30}
	for i, u := range utxos.Utxos {
		//fmt.Println(u.Value, vals[i])
		if u.Value != vals[i] {
			t.Fatalf("not equal %d", i)
		}
	}
}

func TestUtxos_Choose(t *testing.T) {
	ns := getNativeFunc(nil, nil)
	r, _ := hex.DecodeString(redeem)
	rk := btcutil.Hash160(r)
	redeemKey := hex.EncodeToString(rk)
	setBtcTxParam(ns.GetCacheDB(), redeemKey)
	putUtxos(ns, 1, redeemKey, utxos)

	txb, _ := hex.DecodeString(wTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)

	set, sum, _, err := chooseUtxos(ns, 1, 35e4, mtx.TxOut, rk, 5, 7)
	if err != nil {
		t.Fatal(err)
	}
	if !(sum == 35e4 && len(set) == 2 && bytes.Equal(set[0].ScriptPubkey, witPubScript) &&
		bytes.Equal(set[1].ScriptPubkey, p2sh)) {
		t.Fatal("wrong choose")
	}
	_, _, _, err = chooseUtxos(ns, 1, 100e4, mtx.TxOut, rk, 5, 7)
	if err == nil {
		t.Fatal("err should not be nil")
	}
}
