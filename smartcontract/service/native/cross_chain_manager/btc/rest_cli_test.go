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
	cli := NewRestClient("172.168.3.73:50071")
	h, err := cli.GetCurrentHeightFromSpv()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}
	fmt.Printf("height is %d\n", h)
}

func TestRestClient_GetUtxosFromSpv(t *testing.T) {
	cli := NewRestClient(IP)
	ins, sum, err := cli.GetUtxosFromSpv("2N5cY8y9RtbbvQRWkX5zAwTPCxSZF9xEj2C", 1000, 100)
	if err != nil {
		t.Fatalf("Failed to get utxos: %v", err)
	}
	fmt.Printf("Get %d utxos, total %d satoshi\n", len(ins), sum)
}

func TestRestClient_GetHeaderFromSpv(t *testing.T) {
	cli := NewRestClient(IP)
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
	cli := NewRestClient("172.168.3.73:50071")
	err := cli.UnlockUtxoInSpv("d5b57529cc831b1eafa78452f6c6cf0f1782572e3b29a3130010334605946cca", 0)
	err = cli.UnlockUtxoInSpv("974533f80f82943b26933c140ed56d4a3805167c711de61a8939160096161011", 0)
	err = cli.UnlockUtxoInSpv("1318594c73bf8ba1597b2dd3b643ec141607180310a9eaa170b53013bd4d0db4", 0)
	err = cli.UnlockUtxoInSpv("aa03857ff7b13d565b79d3724e516822cca223eb5dd83dd7cb35094bb7070032", 0)
	err = cli.UnlockUtxoInSpv("7a334fc633db4f8fb272f05c367899f363c350dd25c0f0c3a5e0962e6fb30b20", 1)

	if err != nil {
		t.Fatalf("Failed to unlock: %v", err)
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
	raw := "01000000016c115b4a219c2826f372f71ee9cd1d98954a96d11b71ef2d2972c82b48bfb0cd01000000fd600200483045022100df3a0785c79c6bea62d51c3e622bdb8886e6c61760e0b5f5de29951e4e076e43022034cde848ff2330232e3254381d99c4c2f8da1919c57bee29478f03e405284c9d0148304502210093858c96f57e0748f99f448e6d17270686fa4f956d4495a1470192132d1a46fb02203375bf0822b89bf60ab77dea184944c9e8f755d2dd2a2674d874be0d178f1be401483045022100b4870965144c52c20baf7969e644f574a5eea1142a4f29bdcb44b48c3406510b0220659943363fc5e8aec029cd36ea88575619440ff5d56344006f42215e239cf94101483045022100a9a1edfb5e35076f144d1629fde1a519438fd6eb295369f03af47f8f0644939102206aac3c7a9ae164febb1d5755c5c57a6d2a9950a5e3e63ce32a6319fdced79d0a0147304402206b1e0f0cf5c4a890da6e64faf78ad320e82cff63aabf56da9a2e2dd4f3395b6f0220439b6d7518a12cec3bfef2ffcd9ae30f4c852029d3c234e8f709e0ae207fe21e014cf15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57aeffffffff02e8030000000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac24360f000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000"

	rawtx, _ := hex.DecodeString(raw)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(rawtx), wire.ProtocolVersion, wire.LatestEncoding)
	err := cli.BroadcastTxBySpv(mtx)
	if err != nil {
		t.Fatalf("Failed to broadcast tx: %v", err)
	}
}
