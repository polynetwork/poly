package relayer_manager

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelayerListParam_Serialization(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = [][]byte{{1, 2, 4, 6}, {1, 4, 5, 7}, {1, 3, 5, 7, 9}}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	source := common.NewZeroCopySource(sink.Bytes())
	var p RelayerListParam
	err := p.Deserialization(source)
	assert.Nil(t, err)
}
