package ont

import (
	"github.com/ontio/multi-chain/common"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFromMerkleValue(t *testing.T) {
	m := FromMerkleValue {
		TxHash: common.UINT256_EMPTY,
		CreateCrossChainTxMerkle: &CreateCrossChainTxMerkle{
			FromChainID: 123,
			FromContractAddress: "123",
			ToChainID: 123,
			Fee: 123,
			Method: "123",
			Args: []byte{1, 2, 3},
		},
	}

	sink := common.NewZeroCopySink(nil)
	m.Serialization(sink)

	var mer FromMerkleValue
	err := mer.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}


func TestToMerkleValue(t *testing.T) {
	m := ToMerkleValue {
		TxHash: common.UINT256_EMPTY,
		ToContractAddress: "123",
		MakeTxParam: &crosscommon.MakeTxParam{
			TxHash: "123",
			FromChainID: 123,
			FromContractAddress: "123",
			ToChainID: 123,
			Method: "123",
			Args: []byte{1, 2, 3},
		},
	}
	sink := common.NewZeroCopySink(nil)
	m.Serialization(sink)

	var mer ToMerkleValue
	err := mer.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}