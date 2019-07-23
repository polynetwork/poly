package btc

import (
	"encoding/hex"
	"testing"
)

var Height = uint32(1385811)
var Proof = "000000208913fab25ce8cfd55d4b7dd16b021f066334889d4769f603a93e2f00000000002509ac37c2192b288cbb35556bed2ffb9b0b985fe14fea43316d1236ac4c851807967d5b67d8001d73033f6c0c00000005c12809363517c0f8add27808892783ddcb06799f55a9a34dc0c21dd3b3cdd886d87d5399defa486dcdc96f3f10ef9ce046d28fa85b7a696d510634e2a2edbb5363b592f9002b6e38eedac6fef5f879a4cd8c39a642a9e9b034feedb0c60e8a09248e24a1d29371a5bf7e57fcb50d462a7ab6d2811eb7432b24937d42f98c7335053d7f04752c9a650925882d2ecfec09774be53d21fe888e93b6feeac894b88e023700"
var Mr = "18854cac36126d3143ea4fe15f980b9bfb2fed6b5535bb8c282b19c237ac0925"
var TxStr = "010000000122c28e72ecb31b948a2e4fd8fd9ef46c7d8fc391d2c05521b6cc87fb06c4c45a010000006b483045022100b0c4102e4b52556926df870c0a060b1fdf3259a6c4b8bffc2ecf33af334da837022042649c2b1c0bcc3143c8c6913220f8602fc67f136dbc9b2f37b6682e79175ef5012102098bf881769260de6ff5d62a2ea83fc7749c94ceca3538fe12854891d9a7ba50ffffffff02305f2bf9000000001976a914ca5317b7e57d325a620338f247af334f0a85ab2d88ac0000000000000000536a4c500001b86900011c55553eb858abc1c54f7e624cc852f40bb6f11ac0ea9576f03899f6046f2c3a8cfdbe5e8fefcd5858c677ab7ff65b7d95e1060547cd32138db801a9b5dee1a5be8c3f0207f30eee6ce000000000"

// This test's data is from TestNet, So it's not a good case
func TestVerifyBtc(t *testing.T) {
	//Successful situation
	proof, err := hex.DecodeString(Proof)
	if err != nil {
		t.Fatalf("Failed to decode proof: %v", err)
	}

	txInBytes, err := hex.DecodeString(TxStr)
	if err != nil {
		t.Fatalf("Failed to decode string to hex: %v", err)
	}

	res, err := VerifyBtcTx(proof, txInBytes, Height)
	if err != nil || res != true {
		t.Fatalf("Failed to verify: %v", err)
	}
	t.Log("Successful situation pass")
	//Failure situation

}

func TestRestClient_GetHeaderFromSpv(t *testing.T) {
	h, err := NewRestClient().GetHeaderFromSpv(Height)
	if err != nil {
		t.Fatalf("Failed to get header: %v\n", err)
	}
	if h.MerkleRoot.String() != Mr {
		t.Fatalf("Merkle root %s from spv not equal to %s\n", h.MerkleRoot.String(), Mr)
	}
	t.Log("get header from spv passed")
}
