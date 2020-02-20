/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package lock_proxy

import (
	"testing"

	"encoding/hex"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/stretchr/testify/assert"
)

func TestLockParam_Serialize(t *testing.T) {
	fromAddr, _ := common.AddressFromBase58("709c937270e1d5a490718a2b4a230186bdd06a01")
	toAddrBs, _ := hex.DecodeString("709c937270e1d5a490718a2b4a230186bdd06a02")
	param := LockParam{
		SourceAssetHash: utils.OntContractAddress,
		FromAddress:     fromAddr,
		ToChainID:       0,
		ToAddress:       toAddrBs,
		Value:           1,
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	param2 := LockParam{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := param2.Deserialization(source); err != nil {
		t.Fatal("state deserialize fail!")
	}

	assert.Equal(t, param, param2)
}

func TestArgs_Serialize(t *testing.T) {

	toAddrBs, _ := hex.DecodeString("709c937270e1d5a490718a2b4a230186bdd06a02")
	args := Args{
		TargetAssetHash: utils.OntContractAddress[:],
		ToAddress:       toAddrBs,
		Value:           1,
	}
	sink := common.NewZeroCopySink(nil)
	args.Serialization(sink)

	args1 := Args{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := args1.Deserialization(source); err != nil {
		t.Fatal("state deserialize fail!")
	}
	assert.Equal(t, args, args1)
}
