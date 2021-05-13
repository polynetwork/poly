package neo3_state_manager

import (
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStateValidatorListParam_Serialization(t *testing.T) {
	params := new(StateValidatorListParam)
	params.StateValidators = []string{
		"023e9b32ea89b94d066e649b124fd50e396ee91369e8e2a6ae1b11c170d022256d",
		"03009b7540e10f2562e5fd8fac9eaec25166a58b26e412348ff5a86927bfac22a2",
		"02ba2c70f5996f357a43198705859fae2cfea13e1172962800772b3d588a9d4abd",
		"03408dcd416396f64783ac587ea1e1593c57d9fea880c8a6a1920e92a259477806",
	}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	source := common.NewZeroCopySource(sink.Bytes())
	var p StateValidatorListParam
	err := p.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(p.StateValidators))
}
