package ont

import (
	"testing"

	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestPeer_Serialization(t *testing.T) {
	paramSerialize := new(Peer)
	paramSerialize.Index = 1
	paramSerialize.PeerPubkey = "abcdefg"
	sink := common.NewZeroCopySink(nil)
	paramSerialize.Serialization(sink)

	paramDeserialize := new(Peer)
	err := paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}

func TestKeyHeights_Serialization(t *testing.T) {
	paramSerialize := new(KeyHeights)
	paramSerialize.HeightList = []uint32{1, 3, 5}
	sink := common.NewZeroCopySink(nil)
	paramSerialize.Serialization(sink)

	paramDeserialize := new(KeyHeights)
	err := paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}

//func TestConsensusPeers_Serialization(t *testing.T) {
//	paramSerialize := new(ConsensusPeers)
//	paramSerialize.Height = 1
//	paramSerialize.ChainID = 0
//	paramSerialize.PeerMap =
//	sink := common.NewZeroCopySink(nil)
//	paramSerialize.Serialization(sink)
//
//	paramDeserialize := new(ConsensusPeers)
//	err := paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
//	assert.Nil(t, err)
//	assert.Equal(t, paramDeserialize, paramSerialize)
//}
