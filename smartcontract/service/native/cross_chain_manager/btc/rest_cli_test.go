package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"testing"
)

func TestRestClient_ChangeSpvWatchedAddr(t *testing.T) {
	cli := NewRestClient(IP)
	err := cli.ChangeSpvWatchedAddr("2N5cY8y9RtbbvQRWkX5zAwTPCxSZF9xEj2C", "add")
	if err != nil {
		t.Fatalf("Failed to change addr: %v", err)
	}
	addrs, err := cli.GetWatchedAddrsFromSpv()
	if err != nil {
		t.Fatalf("Failed to get watched addrs: %v", err)
	}
	flag := 0
	for _, a := range addrs {
		fmt.Println(a)
		if a == "2N5cY8y9RtbbvQRWkX5zAwTPCxSZF9xEj2C" {
			flag = 1
			break
		}
	}
	if flag != 1 {
		t.Fatalf("addr not found")
	}
}

func TestRestClient_GetCurrentHeightFromSpv(t *testing.T) {
	cli := NewRestClient("172.168.3.75:50071")
	h, err := cli.GetCurrentHeightFromSpv()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}
	fmt.Printf("height is %d\n", h)
}

func TestRestClient_GetWatchedAddrsFromSpv(t *testing.T) {
	cli := NewRestClient("172.168.3.73:50071")
	addrs, err := cli.GetWatchedAddrsFromSpv()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}
	fmt.Printf("addrs is %v\n", addrs)
}

func TestRestClient_GetUtxosFromSpv(t *testing.T) {
	cli := NewRestClient("172.168.3.73:50071")
	ins, sum, err := cli.GetUtxosFromSpv("2N5cY8y9RtbbvQRWkX5zAwTPCxSZF9xEj2C", 1000, 100, true)
	if err != nil {
		t.Fatalf("Failed to get utxos: %v", err)
	}
	fmt.Printf("Get %d utxos, total %d satoshi\n", len(ins), sum)
}

func TestRestClient_GetHeaderFromSpv(t *testing.T) {
	cli := NewRestClient("172.168.3.73:50071")
	h, err := cli.GetHeaderFromSpv(1571626)
	if err != nil {
		t.Fatalf("Failed to get header: %v", err)
	}

	if h.MerkleRoot.String() != "82c5b1b0aa49033ef8dcd3ac674e1f62713394a1792feffad9e77ed7dcb1708b" ||
		h.PrevBlock.String() != "0000000000000114f8f59f527da59504816b4a8aa558186ce6d105d2be7d0ac6" {
		t.Fatalf("wrong header")
	}
}

func TestRestClient_UnlockUtxoInSpv(t *testing.T) {
	for _, i := range []string{"73", "74", "75", "76", "78", "79", "80"}{
		cli := NewRestClient("172.168.3." + i + ":50071")
		err := cli.UnlockUtxoInSpv("d5b57529cc831b1eafa78452f6c6cf0f1782572e3b29a3130010334605946cca", 0)
		err = cli.UnlockUtxoInSpv("974533f80f82943b26933c140ed56d4a3805167c711de61a8939160096161011", 0)
		err = cli.UnlockUtxoInSpv("1318594c73bf8ba1597b2dd3b643ec141607180310a9eaa170b53013bd4d0db4", 0)
		err = cli.UnlockUtxoInSpv("aa03857ff7b13d565b79d3724e516822cca223eb5dd83dd7cb35094bb7070032", 0)
		err = cli.UnlockUtxoInSpv("7a334fc633db4f8fb272f05c367899f363c350dd25c0f0c3a5e0962e6fb30b20", 1)

		if err != nil {
			t.Fatalf("Failed to unlock: %v", err)
		}
	}

}

func TestRestClient_GetFeeRateFromSpv(t *testing.T) {
	cli := NewRestClient(IP)
	rate, err := cli.GetFeeRateFromSpv(0)
	if err != nil {
		t.Fatalf("Failed to get fee rate: %v", err)
	}

	fmt.Printf("rate is %d\n", rate)
}

func TestRestClient_BroadcastTxBySpv(t *testing.T) {
	cli := NewRestClient("172.168.3.80:50071")
	raw := "010000000111101696001639891ae61d717c1605384a6dd50e143c93263b94820ff833459700000000fd5e02004830450221009e9b8aa58a58539f1cdd5af70aba26500ee144504f1a893e119ce359f69afab40220392ba14975fc52443b745813bec66830f198c8fa19213dded02ccf5bf0c9482301473044022055b76cbd2755c27652a566465a6646e6b3f17e688bc20556434bb3f69d261e2e02206f23363b9eb4fc0d925f8f9580e0e119ca5c92a05473cdb61f1ec905d0104f830147304402202723f53db17c5b3fe2683eb4b341fbffbc2226369effad492f877ec0f947e680022012b1cdb390f940e35ee8a3aac2a1ea4df5d738d1e7df00a4d2dfebc727693a2001483045022100b984d24c3e02aa9f364e9c6c198d75781d42e8ad44725f74385bf38622f5df4a0220457ef55566995b36dde3d980e462089ea6901cc5951d93b990eef770d06090fb01473044022071c008ef6ee46bd0b54535b686f21ba621cddb20b36489253849645a0c78ac19022070cdde2bf1d7c346d0f39f98d86121475d83d9c0175f263a6d2824fcea055009014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffff02e8030000000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac703a0f000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000"

	rawtx, _ := hex.DecodeString(raw)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(rawtx), wire.ProtocolVersion, wire.LatestEncoding)
	err := cli.BroadcastTxBySpv(mtx)
	if err != nil {
		t.Fatalf("Failed to broadcast tx: %v", err)
	}
}

func TestRestClient_RollbackSpv(t *testing.T) {
	cli := NewRestClient("0.0.0.0:50071")

	prevh, err := cli.GetCurrentHeightFromSpv()
	if err != nil {
		t.Fatalf("Failed to get height from spv: %v", err)
	}

	err = cli.RollbackSpv("2019-08-01 22:21:10")
	if err != nil {
		t.Fatalf("Failed to roll back: %v", err)
	}

	nowh, err := cli.GetCurrentHeightFromSpv()
	if err != nil {
		t.Fatalf("Failed to get height from spv: %v", err)
	}

	fmt.Printf("prev height is %d, now is %d\n", prevh, nowh)
}
