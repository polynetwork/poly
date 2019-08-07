package btc

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"testing"
)

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
