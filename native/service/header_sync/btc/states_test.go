package btc

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestStoredHeader(t *testing.T) {

	totalWork, _ := (new(big.Int)).SetString("123", 10)

	header := StoredHeader{
		Header: wire.BlockHeader{
			Version:    1,
			PrevBlock:  chainhash.Hash{1, 2, 3},
			MerkleRoot: chainhash.Hash{2, 2, 3},
			Timestamp:  time.Unix(time.Now().Unix(), 0),
			Nonce:      100,
			Bits:       200,
		},
		Height:    uint32(100),
		totalWork: totalWork,
	}
	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)

	h := new(StoredHeader)
	err := h.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, header, *h)
}
