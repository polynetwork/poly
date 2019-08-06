package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/gcash/bchd/txscript"
	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/smartcontract"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
	"github.com/ontio/multi-chain/smartcontract/storage"
	"testing"
)

var Height = uint32(1571626)
var Proof = "00000020c60a7dbed205d1e66c1858a58a4a6b810495a57d529ff5f814010000000000008b70b1dcd77ee7d9faef2f79a1943371621f4e67acd3dcf83e0349aab0b1c5822732405dffff001dae137c5d900000000967e5d59f9b3aea1dacdb50321dbdbd38404fcde592c08edb5ec25d8910ee161b97ac69bea21159a411690ca2e89ea9a1c87fb9dfc1bd6a55cb7d238eb7c21d8dca6c94054633100013a3293b2e5782170fcfc6f65284a7af1e1b83cc2975b5d5a4d13efb1fc25a91a6dd8af7a5e1a624d7641118c3e16b9a9258aa49f7c3cde31adaecb1943d4f432e751f93bf58481fd67b753ca07aab04823e0e5abc515b36c298305f40a432b0559ca74aab15ae5ff77a24da5c177d8308cb5f38c981ed25de8c82c5fd94dfb81cb4cf9685e39abbae8b41327a8b0542ae6810b87d45237adb9f21077f3151a651ac47501ee0b5b8097c1bb8508c69e7eaae359586274d8c77b7d42a3c113f0d9e62d21953e30085b751173ace8abbb16905fed55b3f041503f70600"
var Header = "00000020c60a7dbed205d1e66c1858a58a4a6b810495a57d529ff5f814010000000000008b70b1dcd77ee7d9faef2f79a1943371621f4e67acd3dcf83e0349aab0b1c5822732405dffff001dae137c5d"
var RawTxStr = "01000000013981ddca291cc22bac64ddfb5000e1f0cc3a44744d46c75bcb7ebc16dd85e552020000006a47304402206d5a299ab1e7e69b5c051581a6a3cf0e1ea4cd93354b6676874c5ac59a08f07b02207ff679e76507162890ea51f7b215f999562a3b912c6eb89a3667dcaca27e2dc9012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff03d00700000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000000000002c6a2a66000000000000000100000000000003e86f28d2e8cee08857f569e5a1b147c5d5e87339e081b5ae5dfa60130f00000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"

var addr = "mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57"
var privkey = "cRRMYvoHPNQu1tCz4ajPxytBVc2SN6GWLAVuyjzm4MVwyqZVrAcX"
var getNativeFunc = func() *native.NativeService {
	store, _ := leveldbstore.NewMemLevelDBStore()
	cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	sc := smartcontract.SmartContract{
		CacheDB: cacheDB,
	}
	service := &native.NativeService{
		CacheDB:    sc.CacheDB,
		ServiceMap: make(map[string]native.Handler),
	}

	return service
}

//func TestBTCHandler_Verify(t *testing.T) {
//	handler := NewBTCHandler()
//	service := getNativeFunc()
//
//
//	p, err := handler.Verify(service)
//	if err != nil {
//
//	}
//}

// This test's data is from TestNet.
// Need to start a spv client daemon
func TestVerifyBtcTx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		proof      []byte
		height     uint32
		tx         []byte
		pks        [][]byte
		req        int
		isPositive bool
		service    *native.NativeService
	}{
		{
			name:       "positive case",
			proof:      hexToBytes(Proof),
			height:     Height,
			tx:         hexToBytes(RawTxStr),
			req:        5,
			isPositive: true,
			service: func() *native.NativeService {
				service := getNativeFunc()
				sideChain := &side_chain_manager.SideChain{
					Chainid:      1,
					Name:         "ONT",
					BlocksToWait: 6,
				}
				contract := utils.SideChainManagerContractAddress
				sink := common.NewZeroCopySink(nil)
				_ = sideChain.Serialization(sink)
				chainidByte, _ := utils.GetUint64Bytes(sideChain.Chainid)
				service.CacheDB.Put(utils.ConcatKey(contract, []byte(side_chain_manager.SIDE_CHAIN), chainidByte),
					cstates.GenRawStorageItem(sink.Bytes()))
				return service
			}(),
		},
		//{
		//	name: "negative case",
		//	proof: hexToBytes("00000020e6a0582678e1524992024f449d1f15dcd83290a7de5fe4b797f20100000000002dcb0a90a23fc8054197a01301baa03350429a62d381fd8bb35254d1560fdc533dbb3a5dffff001d42c197b34c0100000a3ba36968845e5b0a8f84b23046b6f382644d682faef46ae9d451d208a8381e210099d8d70842d0e0759f9e490c3b636dd38a0b40feecdb0338cb29e3dcececd36713b76ffcbb051f17d74d74618db6765853c9a53f7b7dbc2b457fa2a596b82cc49d0cf0b8dfd86cbb9791c3cda6070c8588190f024e7337e7196a0754cdc11d9c91fd64e5424a6c7dc62805b18342a81c95f241f3f559475f5529d32e8fc143afc0648a8b46cabc3f8d032de142fe50113e28b68a38bf5170a2bcc70fe5cbafce3cd0f949444ceff1a26014f95b8b68091ea8daa53a43a4b01cfde95206c162d867c23063c5bd6ec7bff2f794d1e3ffebccf5ba92187c9c5127c4fd31da2f80c7e925e9937c0b7ef7eea38a42bfa1bed2789ad0c3b050a64fba8ce57f623ba5d20189d9b88d2beb24a990456cd2e06594ce91c145b73d775f5d9464d450e407036b3d00"),
		//	height: 1571068,
		//	tx: hexToBytes("0100000001c86e0560ea1b56824c3988a606cd7499493c53320bc321c7c98a4e9be543a58c000000006a4730440220380858f558a3d9403b154a856453fb27d380422d27ad10973ab03f81628971c502201b29725043618db50892ef2f29535b21e69448d1c56e5642c74a33c2d04ea7c2012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff02a0320f00000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188acd007000000000000f15421023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae00000000"),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		_, pubk7 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv7))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//		pubks[6] = pubk7.SerializeCompressed()
		//
		//		return pubks
		//	}(),
		//	req: 4,
		//	isPositive: false,
		//},
		//{
		//	name: "wrong proof",
		//	proof: hexToBytes("00000020e6a0582678e1524992024f449d1f15dcd83290a7de5fe4b797f20100000000002dcb0a90a23fc8054197a01301baa03350429a62d381fd8bb35254d1560fdc533dbb3a5dffff001d42c197b34c0100000a3ba36968845e5b0a8f84b23046b6f382644d682faef46ae9d451d208a8381e210099d8d70842d0e0759f9e490c3b636dd38a0b40feecdb0338cb29e3dcececd36713b76ffcbb051f17d74d74618db6765853c9a53f7b7dbc2b457fa2a596b82cc49d0cf0b8dfd86cbb9791c3cda6070c8588190f024e7337e7196a0754cdc11d9c91fd64e5424a6c7dc62805b18342a81c95f241f3f559475f5529d32e8fc143afc0648a8b46cabc3f8d032de142fe50113e28b68a38bf5170a2bcc70fe5cbafce3cd0f949444ceff1a26014f95b8b68091ea8daa53a43a4b01cfde95206c162d867c23063c5bd6ec7bff2f794d1e3ffebccf5ba92187c9c5127c4fd31da2f80c7e925e9937c0b7ef7eea38a42bfa1bed2789ad0c3b050a64fba8ce57f623ba5d20189d9b88d2beb24a990456cd2e06594ce91c145b73d775f5d9464d450e407036b3d00"),
		//	height: Height,
		//	tx: hexToBytes(RawTxStr),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		_, pubk7 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv7))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//		pubks[6] = pubk7.SerializeCompressed()
		//
		//		return pubks
		//	}(),
		//	req: 4,
		//	isPositive: false,
		//},
		//{
		//	name: "wrong height",
		//	proof: hexToBytes(Proof),
		//	height: Height + 1,
		//	tx: hexToBytes(RawTxStr),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		_, pubk7 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv7))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//		pubks[6] = pubk7.SerializeCompressed()
		//
		//		return pubks
		//	}(),
		//	req: 4,
		//	isPositive: false,
		//},
		//{
		//	name: "wrong transaction",
		//	proof: hexToBytes(Proof),
		//	height: Height,
		//	tx: hexToBytes("0100000001c86e0560ea1b56824c3988a606cd7499493c53320bc321c7c98a4e9be543a58c000000006a4730440220380858f558a3d9403b154a856453fb27d380422d27ad10973ab03f81628971c502201b29725043618db50892ef2f29535b21e69448d1c56e5642c74a33c2d04ea7c2012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff02a0320f00000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188acd007000000000000f15421023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae00000000"),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		_, pubk7 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv7))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//		pubks[6] = pubk7.SerializeCompressed()
		//
		//		return pubks
		//	}(),
		//	req: 4,
		//	isPositive: false,
		//},
		//{
		//	name: "wrong pubkeys",
		//	proof: hexToBytes(Proof),
		//	height: Height,
		//	tx: hexToBytes(RawTxStr),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//		pubks[6] = pubk1.SerializeCompressed() //here
		//
		//		return pubks
		//	}(),
		//	req: 4,
		//	isPositive: false,
		//},
		//{
		//	name: "wrong length of pubkeys",
		//	proof: hexToBytes(Proof),
		//	height: Height,
		//	tx: hexToBytes(RawTxStr),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//
		//		return pubks
		//	}(),
		//	req: 4,
		//	isPositive: true,
		//},
		//{
		//	name: "wrong require",
		//	proof: hexToBytes(Proof),
		//	height: Height,
		//	tx: hexToBytes(RawTxStr),
		//	pks: func() ([][]byte) {
		//		_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
		//		_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
		//		_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
		//		_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
		//		_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
		//		_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
		//		pubks := make([][]byte, 0)
		//		pubks[0] = pubk1.SerializeCompressed()
		//		pubks[1] = pubk2.SerializeCompressed()
		//		pubks[2] = pubk3.SerializeCompressed()
		//		pubks[3] = pubk4.SerializeCompressed()
		//		pubks[4] = pubk5.SerializeCompressed()
		//		pubks[5] = pubk6.SerializeCompressed()
		//
		//		return pubks
		//	}(),
		//	req: 5,
		//	isPositive: true,
		//},
	}

	for _, test := range tests {
		isPassed, _, err := verifyBtcTx(test.service, test.proof, test.tx, test.height)
		if test.isPositive && (!isPassed || err != nil) {
			t.Fatalf("Failed to verify this positive case: %s-%v", test.name, err)
		} else if !test.isPositive && (isPassed || err == nil) {
			t.Fatalf("Failed to verify this negtive case: %s-%v", test.name, err)
		}

		if !test.isPositive {
			fmt.Printf("negative case: %s, error is %v", test.name, err)
		}
	}
}

func TestMakeBtcTx(t *testing.T) {
	service := getNativeFunc()
	err := makeBtcTx(service, map[string]int64{
		"mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57": 10,
	})
	if err != nil {
		t.Fatalf("Failed to make tx: %v", err)
	}

	tx, err := service.CacheDB.Get([]byte(BTC_TX_PREFIX))
	if err != nil {
		t.Fatalf("Failed to get tx: %v", err)
	}

	fmt.Printf("raw tx: %x\n", tx)

	mtx := wire.NewMsgTx(wire.TxVersion)
	err = mtx.BtcDecode(bytes.NewBuffer(tx), wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		t.Fatalf("Failed to decode tx: %v", err)
	}

	for i, in := range mtx.TxIn {
		fmt.Printf("No%d input: %s\n", i, in.PreviousOutPoint.String())
	}

	for i, out := range mtx.TxOut {
		s, err := txscript.DisasmString(out.PkScript)
		if err != nil {
			t.Fatalf("Failed to disasm: %v", err)
		}
		fmt.Printf("No%d output: value: %d; pkScript: %s; \n", i, out.Value, s)
	}
}

func TestTargetChainParam(t *testing.T) {
	//flag := []byte{OP_RETURN_SCRIPT_FLAG}
	//chainId := make([]byte, 8)
	//binary.BigEndian.PutUint64(chainId, uint64(1))
	//fee := make([]byte, 8)
	//binary.BigEndian.PutUint64(fee, 100000)
	//addrStr := "AXEJrpNMUhAvRo9ETbzWqAVUBpXXeAFY9u"
	//addr := base58.Decode(addrStr)
	//
	//data := append(append(append(flag, chainId...), fee...), addr...)
	//s, err := txscript.NullDataScript(data)
	//if err != nil {
	//	t.Fatalf("Failed to build script: %v", err)
	//}
	//
	//var param targetChainParam
	//out := &wire.TxOut{
	//	Value:    0,
	//	PkScript: s,
	//}
	//err = param.resolve(1e10, out)
	//if err != nil {
	//	t.Fatalf("Failed to resolve param: %v", err)
	//}
	//
	//if param.ChainId != 1 || param.Fee != 100000 || !bytes.Equal(param.Addr, addr) {
	//	t.Fatal("wrong param")
	//}

	//s, err := buildScript(getPubKeys(), 5)
	//if err != nil {
	//	t.Fatalf("FAiled to build: %v", err)
	//}
	//
	//addr, err := btcutil.NewAddressScriptHash(s, &chaincfg.TestNet3Params)
	//if err != nil {
	//	t.Fatalf("Failed to get addr: %v", err)
	//}
	//
	//fmt.Println(addr.EncodeAddress())
	//
	//sc, err := txscript.PayToAddrScript(addr)
	//if err != nil {
	//	t.Fatalf("Failed to get script: %v", err)
	//}
	//str, err := txscript.DisasmString(sc)
	//if err != nil {
	//	t.Fatalf("Failed to disasm: %v", err)
	//}
	//fmt.Println(str)

	//proof, err := hex.DecodeString(Proof)
	//mb := wire_bch.MsgMerkleBlock{}
	//err = mb.BchDecode(bytes.NewReader(proof), wire_bch.ProtocolVersion, wire_bch.LatestEncoding)
	//if err != nil {
	//	t.Fatalf("Failed to get mb: %v", err)
	//}
	//
	////for _, hash := range mb.Hashes {
	////	fmt.Println(hash.String())
	////}
	//fmt.Println(len(mb.Hashes))
	//for i, f := range mb.Flags {
	//	fmt.Printf("No%d=%v\n", i, f)
	//}

	s, _ := buildScript(getPubKeys(), 5)
	fmt.Println(len(s))
}

func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}
