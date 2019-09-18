package common

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncBlockHeaderParam(t *testing.T) {
	p := SyncBlockHeaderParam{
		ChainID: 123,
		Address: common.ADDRESS_EMPTY,
		Headers: [][]byte{{1, 2, 3}},
	}

	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	var param SyncBlockHeaderParam
	err := param.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	assert.NoError(t, err)

	assert.Equal(t, p, param)
}
