/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package neo

import (
	"testing"

	"encoding/binary"
	"github.com/joeqian10/neo-gogogo/block"
	"github.com/joeqian10/neo-gogogo/helper"
	"github.com/joeqian10/neo-gogogo/mpt"
	tx2 "github.com/joeqian10/neo-gogogo/tx"
	"github.com/polynetwork/poly/common"
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
	err := paramSerialize.Serialization(sink)
	assert.Nil(t, err)

	paramDeserialize := new(NeoBlockHeader)
	err = paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
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
	err := paramSerialize.Serialization(sink)
	assert.Nil(t, err)

	paramDeserialize := new(NeoCrossChainMsg)
	err = paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}
