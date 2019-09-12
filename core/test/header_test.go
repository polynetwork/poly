package test

import (
	"bytes"
	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransaction(t *testing.T) {
	acc := account.NewAccount("123")
	header := types.Header{
		Version:    0,
		ChainID:   0,
		PrevBlockHash: common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStatesRoot: common.UINT256_EMPTY,
		BlockRoot: common.UINT256_EMPTY,
		Timestamp: 12,
		Height: 12,
		ConsensusData: 12,
		ConsensusPayload: []byte{1, 2},
		NextBookkeeper: common.ADDRESS_EMPTY,
		Bookkeepers: []keypair.PublicKey{acc.PublicKey},
		SigData: [][]byte{{1, 2, 3}},
	}

	buf := bytes.NewBuffer(nil)
	err := header.Serialize(buf)

	assert.NoError(t, err)

	var h types.Header
	err = h.Deserialize(buf)

	assert.NoError(t, err)

	assert.Equal(t, header, h)
}

