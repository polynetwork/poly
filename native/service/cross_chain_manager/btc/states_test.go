package btc

import (
	"encoding/hex"
	"fmt"
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

func TestCoinSelector_SimpleBnbSearch2(t *testing.T) {
	p2ws, _ := hex.DecodeString("002044978a77e4e983136bf1cca277c45e5bd4eff6a7848e900416daf86fd32c2743")
	p2sh, _ := hex.DecodeString("a91487a9652e9b396545598c0fc72cb5a98848bf93d387")
	p2pkh, _ := hex.DecodeString("76a91428d2e8cee08857f569e5a1b147c5d5e87339e08188ac")

	us := &Utxos{
		Utxos: []*Utxo {
			{
				Op: &OutPoint{
					Hash: []byte{1},
					Index: 0,
				},
				Value: 1e8,
				ScriptPubkey: p2ws,
				AtHeight: 1,
			},
			{
				Op: &OutPoint{
					Hash: []byte{2},
					Index: 0,
				},
				Value: 27589,
				ScriptPubkey: p2sh,
				AtHeight: 1,
			},
		},
	}

	outs := []*wire.TxOut {
		{
			Value: 9904,
			PkScript: p2pkh,
		},
		{
			Value: 0,
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
	}

	res, sum, fee := cs.Select()
	fmt.Println("res", res[0].Value)
	fmt.Println(sum)
	fmt.Println(fee)
}