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

package neo3

import (
	"github.com/joeqian10/neo3-gogogo/block"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/rpc/models"
	"log"
	"testing"

	"github.com/joeqian10/neo3-gogogo/helper"
	"github.com/joeqian10/neo3-gogogo/mpt"
	tx2 "github.com/joeqian10/neo3-gogogo/tx"
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
)

func Test_NeoConsensus_Serialization(t *testing.T) {
	nextConsensus, _ := crypto.AddressToScriptHash("NYmMriQPYAiNxHC7tziV4ABJku5Yqe79N4", helper.DefaultAddressVersion)
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
	prevHash, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	merkleRoot, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
	nextConsensus, _ := crypto.AddressToScriptHash("NVg7LjGcUSrgxgjX3zEgqaksfMaiS8Z6e1", helper.DefaultAddressVersion)
	//vs, _ := crypto.Base64Decode("EQ==")
	witness := tx2.Witness{
		InvocationScript:   []byte{},
		VerificationScript: []byte{},
	}
	genesisHeader := block.NewBlockHeader()
	genesisHeader.SetVersion(0)
	genesisHeader.SetPrevHash(prevHash)
	genesisHeader.SetMerkleRoot(merkleRoot)
	genesisHeader.SetTimeStamp(1468595301000)
	genesisHeader.SetIndex(0)
	genesisHeader.SetPrimaryIndex(0x00)
	genesisHeader.SetNextConsensus(nextConsensus)
	genesisHeader.SetWitnesses([]tx2.Witness{witness})
	paramSerialize := new(NeoBlockHeader)
	paramSerialize.Header = genesisHeader
	sink := common.NewZeroCopySink(nil)
	err := paramSerialize.Serialization(sink)
	assert.Nil(t, err)

	log.Println(helper.BytesToHex(sink.Bytes()))

	paramDeserialize := new(NeoBlockHeader)
	err = paramDeserialize.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, paramDeserialize, paramSerialize)
}

func Test_NeoCrossChainMsg_Serialization(t *testing.T) {
	paramSerialize := &NeoCrossChainMsg{
		StateRoot: &mpt.StateRoot{
			Version:  0,
			Index:    1000,
			RootHash: "0x53360e02b03f548c6fc2f74f760d82a6749df6a844fd117ad7b62504390c8f8c",
			Witnesses: []models.RpcWitness{
				{
					Invocation:   "DEBnIRHFS8tG/6pw4cqZQbOQZri6rboaQPUJTCVS2ZD/HwOZG2m9IG3NJ8E/gTV++o7G1r35l+p5aQcAbqwoP1wTDECeyQcx2M1DP/irLP7sQy/tNRyina2rdK6ATV/QY+Ib4tJ3sYpXaiPx4iGo+AgqUeTRDmD8anfUNtYzjYgos6x9DEDS+medyKx59813WgtCusxLIK0tx50H36tbMGmTUQxR5nHzrpG8nzQ8HKNKRNMgQNBoT4U3pcHMpwJY9bXUge4R",
					Verification: "EwwhAnIujtkuXxpCUIyfti3TyTtoOhUd/wjLU4lwdHzBhsu6DCECkeyvwoMh29AA30IiQMankxndS3LESLsUGkLoTzfm/doMIQOyS9DtdzAVHs/Ne1yheExdO8NYTw1NkKyi1i4gd5tG1wwhA/sZ1ZOuaNWI9PnKa/3WYbE9xjnVVYWhVYYXFD38xLJWFEF7zmyl",
				},
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
