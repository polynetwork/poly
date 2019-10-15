package btc

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
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
		MultiSignInfo:make(map[string][][]byte),
	}

	err := u.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, multiSignInfo, u)
}