package btc

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/wire"
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestBtcProof(t *testing.T) {
	proof := BtcProof{
		Tx:           []byte{1, 2, 3, 4, 5},
		Proof:        []byte{1, 2, 3, 4, 5},
		Height:       12,
		BlocksToWait: 333,
	}
	sink := common.NewZeroCopySink(nil)
	proof.Serialization(sink)

	var p BtcProof
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, proof, p)
}

func TestUTXO(t *testing.T) {
	utxo := Utxo{
		Op: &OutPoint{
			Hash:  []byte{1, 2, 3, 4, 5},
			Index: 123,
		},
		AtHeight:     12,
		Value:        1111,
		ScriptPubkey: []byte{1, 2, 3, 4},
	}
	sink := common.NewZeroCopySink(nil)
	utxo.Serialization(sink)

	var u Utxo

	err := u.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, utxo, u)
}

func TestMultiSignInfo(t *testing.T) {
	multiSignInfo := &MultiSignInfo{
		MultiSignInfo: map[string][][]byte{"zmh": {[]byte("zmh")}},
	}
	sink := common.NewZeroCopySink(nil)
	multiSignInfo.Serialization(sink)

	u := &MultiSignInfo{
		MultiSignInfo: make(map[string][][]byte),
	}

	err := u.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, multiSignInfo, u)
}

func TestCoinSelector_getLossRatio(t *testing.T) {
	p2ws, _ := hex.DecodeString("002044978a77e4e983136bf1cca277c45e5bd4eff6a7848e900416daf86fd32c2743")
	p2sh, _ := hex.DecodeString("a91487a9652e9b396545598c0fc72cb5a98848bf93d387")
	p2pkh, _ := hex.DecodeString("76a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac")

	us := &Utxos{
		Utxos: []*Utxo{
			{
				Op: &OutPoint{
					Hash:  []byte{1},
					Index: 0,
				},
				Value:        1e8,
				ScriptPubkey: p2ws,
				AtHeight:     1,
			},
			{
				Op: &OutPoint{
					Hash:  []byte{2},
					Index: 0,
				},
				Value:        27589,
				ScriptPubkey: p2sh,
				AtHeight:     1,
			},
		},
	}

	outs := []*wire.TxOut{
		{
			Value:    9904,
			PkScript: p2pkh,
		},
		{
			Value:    0,
			PkScript: p2ws,
		},
	}

	sort.Sort(sort.Reverse(us))
	cs := &CoinSelector{
		SortedUtxos: us,
		Target:      uint64(9904),
		MaxP:        MAX_FEE_COST_PERCENTS,
		Tries:       MAX_SELECTING_TRY_LIMIT,
		Mc:          MIN_CHANGE,
		K:           SELECTING_K,
		TxOuts:      outs,
		feeRate:     2,
	}

	fee, lr := cs.getLossRatio(us.Utxos)
	assert.Equal(t, uint64(0x772), fee)
	assert.Equal(t, float64(0.1924474959612278), lr)
}

func TestCoinSelector_SimpleBnbSearch(t *testing.T) {
	ns := getNativeFunc(nil, nil)
	redeemKey := ""
	putUtxos(ns, 0, redeemKey, utxos)
	txb, _ := hex.DecodeString(wTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)
	sort.Sort(sort.Reverse(utxos))

	s := &CoinSelector{
		TxOuts:      mtx.TxOut,
		K:           1.5,
		Mc:          2000,
		Tries:       10000,
		MaxP:        0.2,
		Target:      35e4,
		SortedUtxos: utxos,
		feeRate:     2,
	}
	// normal case
	res, sum, _ := s.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if !(len(res) == 2 && sum == 35e4 && res[0].AtHeight == 3 && res[1].AtHeight == 2 && s.Tries == 9998) {
		t.Fatal("wrong selection")
	}

	s.Target = 12e4
	res, sum, _ = s.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if !(len(res) == 2 && sum == 15e4 && res[0].AtHeight == 2 && res[1].AtHeight == 1 && s.Tries == 9995) {
		t.Fatal("wrong selection")
	}

	s.Target = 34e4
	s.Mc = 3e4
	res, sum, _ = s.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if !(len(res) == 3 && sum == 45e4 && res[0].AtHeight == 3 && res[1].AtHeight == 2 && s.Tries == 9992) {
		t.Fatal("wrong selection")
	}

	// over fee rate
	s.MaxP = 0.001
	res, sum, _ = s.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if res != nil {
		t.Fatal("wrong selection")
	}

	// not enough
	s.Target = 5e5
	s.Mc = 2000
	s.MaxP = 0.2
	res, sum, _ = s.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if res != nil {
		t.Fatal("wrong selection")
	}

	// out of max
	s.Target = 5000
	res, sum, _ = s.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if res != nil {
		t.Fatal("should be nil")
	}
}

func TestCoinSelector_SortedSearch(t *testing.T) {
	ns := getNativeFunc(nil, nil)
	rs, _ := hex.DecodeString(redeem)
	redeemKey := GetUtxoKey(rs)
	putUtxos(ns, 0, redeemKey, utxos)
	txb, _ := hex.DecodeString(wTx)
	mtx := wire.NewMsgTx(wire.TxVersion)
	mtx.BtcDecode(bytes.NewBuffer(txb), wire.TxVersion, wire.LatestEncoding)
	sort.Sort(sort.Reverse(utxos))

	s := &CoinSelector{
		TxOuts:      mtx.TxOut,
		K:           1.5,
		Mc:          2000,
		Tries:       10000,
		MaxP:        0.2,
		Target:      35e4,
		SortedUtxos: utxos,
		feeRate:     2,
	}

	res, sum, _ := s.SortedSearch()
	if !(len(res) == 2 && sum == 35e4 && res[0].AtHeight == 3 && res[1].AtHeight == 2) {
		t.Fatal("wrong selection")
	}

	s.Target = 5e4
	res, sum, _ = s.SortedSearch()
	if !(len(res) == 1 && sum == 5e4 && res[0].AtHeight == 2) {
		t.Fatal("wrong selection")
	}

	s.Target = 5e5
	res, sum, _ = s.SortedSearch()
	if res != nil {
		t.Fatal("wrong")
	}

	s.Target = 41e4
	s.Mc = 5e4
	res, sum, _ = s.SortedSearch()
	if res != nil {
		t.Fatal("wrong")
	}
}
