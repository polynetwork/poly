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
	cli := NewRestClient(IP)
	err := cli.UnlockUtxoInSpv("cdb0bf482bc872292def711bd1964a95981dcde91ef772f326289c214a5b116c", 1)
	err = cli.UnlockUtxoInSpv("974533f80f82943b26933c140ed56d4a3805167c711de61a8939160096161011", 0)
	err = cli.UnlockUtxoInSpv("1318594c73bf8ba1597b2dd3b643ec141607180310a9eaa170b53013bd4d0db4", 0)
	err = cli.UnlockUtxoInSpv("d5b57529cc831b1eafa78452f6c6cf0f1782572e3b29a3130010334605946cca", 0)
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
	cli := NewRestClient(IP)
	raw := "0100000001ba32eb944a29e6c0d26189cc0cc67c5bd34d48ba876de114255bb6e3284ea7d1000000006a473044022040f94d2f640377d28f6aa0176477d0924c13a4772d1344c824ed69aac0d8c48b02200f9d475ff9f877a37b7d3e418f9cca6c0cb1909d3aa16361fd256c7aa05f80e9012103128a2c4525179e47f38cf3fefca37a61548ca4610255b3fb4ee86de2d3e80c0fffffffff0300350c000000000017a91487a9652e9b396545598c0fc72cb5a98848bf93d38700000000000000002c6a2a66000000000000000200000000000003e81727e090b158ee5c69c7e46076a996c4bd6159286ef9621225a0860100000000001976a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac00000000"

	rawtx, _ := hex.DecodeString(raw)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(rawtx), wire.ProtocolVersion, wire.LatestEncoding)
	err := cli.BroadcastTxBySpv(mtx)
	if err != nil {
		t.Fatalf("Failed to broadcast tx: %v", err)
	}
}
