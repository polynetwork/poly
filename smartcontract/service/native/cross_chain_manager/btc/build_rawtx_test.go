package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil/base58"
	"strconv"
	"testing"
)

var inputs = []btcjson.TransactionInput{
	btcjson.TransactionInput{
		Txid: "50a882877b3336dd45f5558308f865745de4f0b836b2fdd1956e4c8e1e54cf92",
		Vout: 1,
	},

	btcjson.TransactionInput{
		Txid: "0b6f173eaa7872c5f40e52d1dca9b050eb72550563c558dbdabca5f4065c39d4",
		Vout: 2,
	},
}
var addrToPay = "1PoAci77dCxnGjx95JhyVksZyVykZoFLq6"
var value = 1.2
var amounts = map[string]float64{
	addrToPay: value,
}

// Test methed `getRawTx` and check the raw transaction's inputs and outputs, especially
// the scriptPub
// This test is for pay2PKH
func TestGetRawTx(t *testing.T) {
	mtx, err := getRawTx(inputs, amounts, nil)
	if err != nil {
		t.Fatalf("Failed to get raw transaction: %v\n", err)
	}
	fmt.Printf("raw tx hash: %x\n", mtx.TxHash())
	pubRipe := base58.Decode(addrToPay)
	pubRipe = pubRipe[1 : len(pubRipe)-4]

	pkScript, err := txscript.ParsePkScript(mtx.TxOut[0].PkScript)
	if err != nil {
		t.Fatalf("Failed to parse pkScript: %v\n", err)
	}

	if int64(value*1e8) != mtx.TxOut[0].Value {
		t.Fatalf("Value in tx's output are not equal: right value is %d, not %d",
			int64(value*1e8), mtx.TxOut[0].Value)
	}
	fmt.Printf("PkScript is : %s\n", pkScript.String())
	if hex.EncodeToString(pubRipe) != pkScript.String()[18:58] {
		t.Fatalf("Pubkey not equal: right key is %s, not %s",
			hex.EncodeToString(pubRipe), pkScript.String()[18:58])
	}
	if inputs[0].Txid+":"+strconv.Itoa(int(inputs[0].Vout)) != mtx.TxIn[0].PreviousOutPoint.String() {
		t.Fatalf("Check input fail: %s is correct, not %s",
			inputs[0].Txid+":"+strconv.Itoa(int(inputs[0].Vout)), mtx.TxIn[0].PreviousOutPoint.String())
	}
	fmt.Printf("Tx Input[0]: %s\n", mtx.TxIn[0].PreviousOutPoint.String())
	fmt.Printf("Tx Output[0]: %d\n", mtx.TxOut[0].Value)
	fmt.Printf("Result of ripe160 is %x, it should equal with the pubkey in pkScript\n", pubRipe)

	t.Log("Build raw transaction test pass")
}

func TestDeserializeRawTx(t *testing.T) {
	mtx, err := getRawTx(inputs, amounts, nil)
	if err != nil {
		t.Fatalf("Failed to get rawtransaction: %v\n", err)
	}

	var mtxInBytes bytes.Buffer
	err = mtx.BtcEncode(&mtxInBytes, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		t.Fatalf("Failed to encode rawtx: %v", err)
	}

	newMtx := wire.NewMsgTx(wire.TxVersion)
	err = newMtx.BtcDecode(&mtxInBytes, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		t.Fatalf("Failed to decode bytes: %v\n", err)
	}

	pkScript, err := txscript.ParsePkScript(newMtx.TxOut[0].PkScript)
	if err != nil {
		t.Fatalf("Failed to parse pkScript: %v\n", err)
	}

	fmt.Printf("PkScript is %s\n", pkScript.String())
}
