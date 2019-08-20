package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil/base58"
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

var Height = uint32(1572760)
var Proof = "000000204149a82a4db84c25eabdd220ae55e568f3332f9a9d6bcc21be8d010000000000f783cb176b1c29fcb191eeb7299a105fc5db9a42be7cec34d08b8d819bb64fe44d1c495d71a5021a2ad64e1e26000000077702820166697756300bb36b2268ff36d93bbe63d09d42b42c7eb52a06aa9153320007b74b0935cbd73dd85deb23a2cc2268514e72d3795b563db1f77f8503aac3690bf489db8b0f3630a0f50a6767790c6f178d1027385f14d7e70ce2622a4a125da8708c3ddfb554fd8a636152007ca6f7ad7251c2514a07ea19a3718fb6b464259f0e6b7b06e34ae8f6c2e54d4d10c603cda1d2c1ebaf093c074e5b51e3a131b237e55e259bf74174441256a61f9d62d250d06ddcec3f6f94a3f6f43e3e3e59a4fc0e7c7dc59b926c2de2f4e9176ffbf7545e17b763cdc962d829500c321002bf00"
var Header = "00000020c60a7dbed205d1e66c1858a58a4a6b810495a57d529ff5f814010000000000008b70b1dcd77ee7d9faef2f79a1943371621f4e67acd3dcf83e0349aab0b1c5822732405dffff001dae137c5d"
var RawTxStr = "0100000001ba32eb944a29e6c0d26189cc0cc67c5bd34d48ba876de114255bb6e3284ea7d1000000006a473044022040f94d2f640377d28f6aa0176477d0924c13a4772d1344c824ed69aac0d8c48b02200f9d475ff9f877a37b7d3e418f9cca6c0cb1909d3aa16361fd256c7aa05f80e9012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff0300350c000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000000000002c6a2a66000000000000000200000000000003e81727e090b158ee5c69c7e46076a996c4bd6159286ef9621225a0860100000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"

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
					Chainid:      2,
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
		isPassed, cp, err := verifyBtcTx(test.service, test.proof, test.tx, test.height)

		if test.isPositive && (!isPassed || err != nil) {
			t.Fatalf("Failed to verify this positive case: %s-%v", test.name, err)
		} else if !test.isPositive && (isPassed || err == nil) {
			t.Fatalf("Failed to verify this negtive case: %s-%v", test.name, err)
		}

		if !test.isPositive {
			fmt.Printf("negative case: %s, error is %v", test.name, err)
		}

		addr := base58.Encode(cp.Addr)
		val := cp.Value
		cid := cp.ChainId
		fee := cp.Fee
		fmt.Printf("addr: %s, value: %d, chainId: %d, fee: %d\n", addr, val, cid, fee)
	}
}

func TestTargetChainParam(t *testing.T) {
	flag := []byte{OP_RETURN_SCRIPT_FLAG}
	chainId := make([]byte, 8)
	binary.BigEndian.PutUint64(chainId, uint64(1))
	fee := make([]byte, 8)
	binary.BigEndian.PutUint64(fee, 100000)
	addrStr := "AXEJrpNMUhAvRo9ETbzWqAVUBpXXeAFY9u"
	addr, _ := common.AddressFromBase58(addrStr)

	data := append(append(append(flag, chainId...), fee...), addr[:]...)
	s, err := txscript.NullDataScript(data)
	if err != nil {
		t.Fatalf("Failed to build script: %v", err)
	}

	var param targetChainParam
	out := &wire.TxOut{
		Value:    0,
		PkScript: s,
	}
	err = param.resolve(1e10, out)
	if err != nil {
		t.Fatalf("Failed to resolve param: %v", err)
	}

	if param.ChainId != 1 || param.Fee != 100000 || !bytes.Equal(param.Addr, addr[:]) {
		t.Fatal("wrong param")
	}

	parsed, _ := common.AddressParseFromBytes(param.Addr)
	fmt.Printf("addr is %s\n", parsed.ToBase58())

	sss, _ := buildScript(getPubKeys(), 5)
	fmt.Printf("s is :%s\n", hex.EncodeToString(sss))
}

func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}
