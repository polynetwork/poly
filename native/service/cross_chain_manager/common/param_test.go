package common

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestMakeTxParamWithSender(t *testing.T) {
	txParam := MakeTxParam{TxHash: []byte("hash"),
		CrossChainID:        []byte("1"),
		FromContractAddress: []byte("from addr"),
		ToChainID:           1,
		ToContractAddress:   []byte("to addr"),
		Method:              "test",
		Args:                []byte("args")}

	value := MakeTxParamWithSender{Sender: ethcommon.HexToAddress("abc"), MakeTxParam: txParam}
	data, err := value.Serialization()

	assert.Nil(t, err)

	var decoded MakeTxParamWithSender
	err = decoded.Deserialization(data)
	assert.Nil(t, err)

	assert.Equal(t, value, decoded)
}
