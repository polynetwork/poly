package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"testing"
)

var Height = uint32(1571076)
var Proof = "000000204792ea66b92f6674fb6fee519057f6393e9b16dc23a5b012d302000000000000b8656b21e501521420a99cf5649805462d13fbc07dcaa74dd80cc9eced95326e00cc3a5dffff001d6f3673ae190100000a552427bdd22ca1dcb153c4a14834c2f8e51fb1dfac048bc8861074c9dfa50afcb7a24cac7560701775dbcae736dda808d2be586c9b8296ea222ad0d95e6945dc4aa4af9dfa6725a01c58a074d3c62b37c4df7e106ede73730084c60aec83c28c318723e3a1f76a73e4e629d57c016817bf5be473bb8050859056c181a416116c3981ddca291cc22bac64ddfb5000e1f0cc3a44744d46c75bcb7ebc16dd85e552230a05bef81f4d27376f312c405bc02369aed75fa6211c9be7f4e2833242df065bf579da2b19e530143ae0914a6b28d362e7911097a0c1ddc7b9d897f895b62dd3316c0a3752749d54a92f31b8269a63d61bbe2751daec138794901c8071c846530e55a188f59b57f3884076bdb4272ff54471852d340f9858ed9836cde3e7df4f4cba97dfc7051650a3f5cb771d20b442040c3712f6ecdd8caee897412cae3c035b2f00"
var Header = "000000204792ea66b92f6674fb6fee519057f6393e9b16dc23a5b012d302000000000000b8656b21e501521420a99cf5649805462d13fbc07dcaa74dd80cc9eced95326e00cc3a5dffff001d6f3673ae"
var RawTxStr = "01000000019c91fd64e5424a6c7dc62805b18342a81c95f241f3f559475f5529d32e8fc143000000006b4830450221008102bc2ac7d6bd82347da24cf080e20a6e4d85b854b05ba671cea47b09a12e34022066d0d82a92b98eee43e0488b9a5229c370edeed577f0a9a9c0e8559284c7700d012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff03d007000000000000f15421023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae00000000000000002c6a2a66000000000000000100000000000003e86f28d2e8cee08857f569e5a1b147c5d5e87339e081b5ae5dfa00230f00000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"

var addr = "mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57"
var privkey = "cRRMYvoHPNQu1tCz4ajPxytBVc2SN6GWLAVuyjzm4MVwyqZVrAcX"

//var addr1 = "mj3LUsSvk9ZQH1pSHvC8LBtsYXsZvbky8H"
//var priv1 = "cTqbqa1YqCf4BaQTwYDGsPAB4VmWKUU67G5S1EtrHSWNRwY6QSag"
//var addr2 = "mtNiC48WWbGRk2zLqiTMwKLhrCk6rBqBen"
//var priv2 = "cT2HP4QvL8c6otn4LrzUWzgMBfTo1gzV2aobN1cTiuHPXH9Jk2ua"
//var addr3 = "mi1bYK8SR3Qsf2cdrxgak3spzFx4EVH1pf"
//var priv3 = "cSQmGg6spbhd23jHQ9HAtz3XU7GYJjYaBmFLWHbyKa9mWzTxEY5A"
//var addr4 = "mz3bTZaQ2tNzsn4szNE8R6gp5zyHuqN29V"
//var priv4 = "cPYAx61EjwshK5SQ6fqH7QGjc8L48xiJV7VRGpYzPSbkkZqrzQ5b"
//var addr5 = "mfzbFf6njbEuyvZGDiAdfKamxWfAMv47NG"
//var priv5 = "cVV9UmtnnhebmSQgHhbDZWCb7zBHbiAGDB9a5M2ffe1WpqvwD5zg"
//var addr6 = "n4ESieuFJq5HCvE5GU8B35YTfShZmFrCKM"
//var priv6 = "cNK7BwHmi8rZiqD2QfwJB1R6bF6qc7iVTMBNjTr2ACbsoq1vWau8"
//var addr7 = "msK9xpuXn5xqr4UK7KyWi9VCaFhiwCqqq6"
//var priv7 = "cUZdDF9sL11ya5civzMRYVYojoojjHbmWWm1yC5uRzfBRePVbQTZ"

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
			name:   "positive case",
			proof:  hexToBytes(Proof),
			height: Height,
			tx:     hexToBytes(RawTxStr),
			pks: func() [][]byte {
				_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
				_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
				_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
				_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
				_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
				_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
				_, pubk7 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv7))
				pubks := make([][]byte, 7)
				pubks[0] = pubk1.SerializeCompressed()
				pubks[1] = pubk2.SerializeCompressed()
				pubks[2] = pubk3.SerializeCompressed()
				pubks[3] = pubk4.SerializeCompressed()
				pubks[4] = pubk5.SerializeCompressed()
				pubks[5] = pubk6.SerializeCompressed()
				pubks[6] = pubk7.SerializeCompressed()

				return pubks
			}(),
			req:        4,
			isPositive: true,
			service: func() *native.NativeService {
				return nil
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
		isPassed, err := VerifyBtcTx(test.service, test.proof, test.tx, test.height)
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

func TestRestClient_GetHeaderFromSpv(t *testing.T) {
	header := wire.BlockHeader{}
	hb, err := hex.DecodeString(Header)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}
	err = header.BtcDecode(bytes.NewBuffer(hb), wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		t.Fatalf("Failed to decode header: %v", err)
	}

	h, err := NewRestClient().GetHeaderFromSpv(Height)
	if err != nil {
		t.Fatalf("Failed to get header: %v\n", err)
	}
	if bytes.Equal(h.MerkleRoot[:], header.MerkleRoot[:]) {
		t.Fatalf("Merkle root %s from spv not equal to %s\n", h.MerkleRoot.String(), header.MerkleRoot[:])
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

	s, err := buildScript(getPubKeys(), 4)
	if err != nil {
		t.Fatalf("FAiled to build: %v", err)
	}

	fmt.Println(hex.EncodeToString(s))

	addr, err := btcutil.NewAddressScriptHash(s, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatalf("Failed to get addr: %v", err)
	}

	fmt.Println(addr.EncodeAddress())

	s, err := txscript.PayToAddrScript(addr)

}

func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}
