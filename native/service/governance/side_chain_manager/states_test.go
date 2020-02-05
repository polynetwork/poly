package side_chain_manager

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSideChain_Serialization(t *testing.T) {
	paramSerialize := new(SideChain)
	paramSerialize.Name = "own"
	paramSerialize.Router = 7
	paramSerialize.ChainId = 8
	paramSerialize.BlocksToWait = 10
	sink := common.NewZeroCopySink(nil)
	err := paramSerialize.Serialization(sink)
	assert.Nil(t, err)

	paramDeserialize := new(SideChain)
	err = paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}
