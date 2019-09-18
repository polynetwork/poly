package common

import (
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntranceParam(t *testing.T) {
	param := EntranceParam {
		SourceChainID: 123,
		TxData: "123",
		Height: 123,
		Proof: "123",
		RelayerAddress: "123",
		TargetChainID: 123,
		Value: "123",
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p EntranceParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}

func TestMakeTxParam(t *testing.T) {
	param := MakeTxParam {
		TxHash: "123",
		FromChainID: 123,
		FromContractAddress: "123",
		ToChainID: 123,
		Method: "test",
		Args: []byte("test"),
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p MakeTxParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}

func TestVoteParam(t *testing.T) {
	param := VoteParam{
		FromChainID: 123,
		Address: "1234",
		TxHash: []byte{1, 2, 3,},
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

