package common

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVoteParam(t *testing.T) {
	param := VoteParam{
		Address: "1234",
		TxHash:  []byte{1, 2, 3},
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p VoteParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}

func TestVote(t *testing.T) {
	m := make(map[string]string, 0)
	m["123"] = "123"
	vote := Vote{
		VoteMap: m,
	}
	sink := common.NewZeroCopySink(nil)
	vote.Serialization(sink)

	var v Vote
	err := v.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}
