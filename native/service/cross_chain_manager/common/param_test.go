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
package common

import (
	"github.com/polynetwork/poly/common"
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
