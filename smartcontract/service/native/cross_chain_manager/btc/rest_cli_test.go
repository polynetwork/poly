package btc

import (
	"fmt"
	"testing"
)

func TestRestClient_ChangeSpvWatchedAddr(t *testing.T) {
	cli := NewRestClient()
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
	cli := NewRestClient()
	h, err := cli.GetCurrentHeightFromSpv()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}
	fmt.Printf("height is %d\n", h)
}

func TestRestClient_GetUtxosFromSpv(t *testing.T) {
	cli := NewRestClient()
	ins, sum, err := cli.GetUtxosFromSpv("2N5cY8y9RtbbvQRWkX5zAwTPCxSZF9xEj2C", 100, 100)
	if err != nil {
		t.Fatalf("Failed to get utxos: %v", err)
	}
	fmt.Printf("Get %d utxos, total %d satoshi\n", len(ins), sum)
}

func TestRestClient_GetHeaderFromSpv(t *testing.T) {
	cli := NewRestClient()
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
	cli := NewRestClient()
	err := cli.UnlockUtxoInSpv("d5b57529cc831b1eafa78452f6c6cf0f1782572e3b29a3130010334605946cca", 0)
	if err != nil {
		t.Fatalf("Failed to unlock: %v", err)
	}
}
