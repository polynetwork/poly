package neo

import (
	"testing"

	"encoding/binary"
	"encoding/hex"
	"github.com/joeqian10/neo-gogogo/block"
	"github.com/joeqian10/neo-gogogo/helper"
	"github.com/joeqian10/neo-gogogo/mpt"
	tx2 "github.com/joeqian10/neo-gogogo/tx"
	"github.com/ontio/multi-chain/common"
	"github.com/stretchr/testify/assert"
)

func Test_NeoConsensus_Serialization(t *testing.T) {
	nextConsensus, _ := helper.UInt160FromString("APyEx5f4Zm4oCHwFWiSTaph1fPBxZacYVR")
	paramSerialize := &NeoConsensus{
		ChainID:       4,
		Height:        100,
		NextConsensus: nextConsensus,
	}
	sink := common.NewZeroCopySink(nil)
	paramSerialize.Serialization(sink)

	paramDeserialize := new(NeoConsensus)
	err := paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}

func Test_NeoBlockHeader_Serialization(t *testing.T) {
	paramSerialize := new(NeoBlockHeader)
	prevHash, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	merKleRoot, _ := helper.UInt256FromString("0x803ff4abe3ea6533bcc0be574efa02f83ae8fdc651c879056b0d9be336c01bf4")
	nextConsensus, _ := helper.AddressToScriptHash("APyEx5f4Zm4oCHwFWiSTaph1fPBxZacYVR")
	consensusData := binary.BigEndian.Uint64(helper.HexToBytes("000000007c2bac1d"))
	genesisHeader := &block.BlockHeader{
		Version:       0,
		PrevHash:      prevHash,
		MerkleRoot:    merKleRoot,
		Timestamp:     1468595301,
		Index:         0,
		NextConsensus: nextConsensus,
		ConsensusData: consensusData,
		Witness: &tx2.Witness{
			InvocationScript:   []byte{0},
			VerificationScript: []byte{81},
		},
	}
	paramSerialize.BlockHeader = genesisHeader
	sink := common.NewZeroCopySink(nil)
	paramSerialize.Serialization(sink)

	paramDeserialize := new(NeoBlockHeader)
	err := paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}

func Test_NeoCrossChainMsg_Serialization(t *testing.T) {
	paramSerialize := &NeoCrossChainMsg{
		StateRoot: &mpt.StateRoot{
			Version:   1,
			Index:     100,
			PreHash:   "803ff4abe3ea6533bcc0be574efa02f83ae8fdc651c879056b0d9be336c01bf4",
			StateRoot: "803ff4abe3ea6533bcc0be574efa02f83ae8fdc651c879056b0d9be336c01bf4",
			Witness: struct {
				InvocationScript   string `json:"invocation"`
				VerificationScript string `json:"verification"`
			}{
				InvocationScript:   "40424db765bc1e92e530292ec04ff8ddffb79bec13f04fd9f85c00163328aa9d64f0b40b74ca8c4b56445c9048c50e6a67df57ab221593612c6165251d9770f7e140465f1d1d3b532fcaa8a98633316e24a07358c857a3565f7cc9a1b87dd3e6dcbb191a7c78c1b57889924e813a0daacea5281884ce814d10469560f43c9d567cf440fd7252d9607389e9b61c577a8705b1d74165979dd9440c4a71d47443fc1014e46957b0a537e1244fd9b4363aefb2df5971749daf9073cfd014aecb7dba2b13ab40c141f6c63267ad12ebadb154a83a3444eccff046de534cda6f29059e531de58bfce6287ca68a62b45766df5522dfed449b3d1bdc0a319ab07d21cf8839f5b59240fee381887b2dc82447fbe9e6db6c1aa9adff8f7a7d2998cea4f901c002098115d7ba7e6218275c8690f86b92e8b641d59152243f2253ff86fa9c2b6413a52256",
				VerificationScript: "552102486fd15702c4490a26703112a5cc1d0923fd697a33406bd5a1c00e0013b09a7021024c7b7fb6c310fccf1ba33b082519d82964ea93868d676662d4a59ad548df0e7d2102aaec38470f6aad0042c6e877cfd8087d2676b0f516fddd362801b9bd3936399e2103b209fd4f53a7170ea4444e0cb0a6bb6a53c2bd016926989cf85f9b0fba17a70c2103b8d9d5771d8f513aa0869b9cc8d50986403b78c6da36890638c3d46a5adce04a2102ca0e27697b9c248f6f16e085fd0061e26f44da85b58ee835c110caa5ec3ba5542102df48f60e8f3e01c48ff40b9b7f1310d7a8b2a193188befe1c2e3df740e89509357ae",
			},
		},
	}

	sink := common.NewZeroCopySink(nil)
	paramSerialize.Serialization(sink)

	paramDeserialize := new(NeoCrossChainMsg)
	err := paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}

func Test_DeSerialization_NeoBlockHeader(t *testing.T) {
	var err error
	header1 := new(NeoBlockHeader)
	header1bs, _ := hex.DecodeString("00000000000000000000000000000000000000000000000000000000000000000000000058106c4e7be51589f2993ea20392f7470ed27fd5a767b006cd1a0d68cda43bee65fc8857000000001dac2b7c00000000ed02ac39071130e1f8fdadd7f0b4bb91a3c92138010000")
	source1 := common.NewZeroCopySource(header1bs)
	err = header1.Deserialization(source1)
	if err != nil {
		t.Errorf("header1.Deserialization = %v", err)
	}

	header1 = new(NeoBlockHeader)
	header1bs, _ = hex.DecodeString("000000004eeb127e012a91ee8e0122bf94165abf707c7619667d69f79e75bc5312f1d4c098f84f36bf669bd79ee28816fa33cf87412c5d898deefd98fe302bdee9aa39b9955b7c5e901a000071e269aa81ba5506ed02ac39071130e1f8fdadd7f0b4bb91a3c92138010000")
	source1 = common.NewZeroCopySource(header1bs)
	err = header1.Deserialization(source1)
	if err != nil {
		t.Errorf("header1.Deserialization = %v", err)
	}

	crossChainMsgStr := "000000005e1b000043dd8d18ed199817a69ab5531fc080d381951e043fb6635dd23a1d9a990bcd7da705796b6b9025476deb1ff2ebd70d75f0e344847a1d98017d8bc90595a0573501c3404dc81e192850389a5f776c5ab0e3247b5b9bc1a857ae1da2f0cb907bca1ad0092c1fd1ce7b0cf58a528d3dda7fa6badaa974187209c2cc7305bc3e99bfe9fd1a401ecef31fb0df7bf13959b52905e6ebc562da9424b341f1481fec5242a567bfe7682debb9332166029260b48e3565ab38219af247a4fd98caa366149c35f6743c4040706b57a38df1cabf5a91057998a9dc973d1ea4ed3a3207493cdbff5bdba799b64bd598e8c4f98dbf52d075271f28303852c7cf13e8cca69f76416521c4bd8e8b5321030f59a5482a4e42a2e5a848608dac4e84a698e567e2860e0ca5f23fc9e818d37c21032e78261370d4d62cf4c13584ca90f46c5565117b5b97544312f2e7b7c36b9eba21026e271722c21c482f0ac74dd932e61cdc2a2dd889633a2c5d8ecef43f2769f51e2103d55bfbcd493d06ab49c09cde0cea5d9ba890d81331a2fcd6f68d329932d0398f54ae"
	crossChainMsgBs, _ := hex.DecodeString(crossChainMsgStr)
	paramDeserialize := new(NeoCrossChainMsg)
	err = paramDeserialize.Deserialization(common.NewZeroCopySource(crossChainMsgBs))
	assert.Nil(t, err)

}
