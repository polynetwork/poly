package side_chain_manager

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterSideChain(t *testing.T) {
	param := RegisterSideChainParam{
		Address:      "123456",
		ChainId:      123,
		Name:         "123456",
		BlocksToWait: 1234,
	}
	sink := common.NewZeroCopySink(nil)
	err := param.Serialization(sink)
	assert.NoError(t, err)

	source := common.NewZeroCopySource(sink.Bytes())
	var p RegisterSideChainParam
	err = p.Deserialization(source)
	assert.NoError(t, err)

	assert.Equal(t, param, p)
}
