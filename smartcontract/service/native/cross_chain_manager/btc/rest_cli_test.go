package btc

import "testing"

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
		if a == "2N5cY8y9RtbbvQRWkX5zAwTPCxSZF9xEj2C" {
			flag = 1
			break
		}
	}
	if flag != 1 {
		t.Fatalf("addr not found")
	}
}
