/*
 * Copyright (C) 2021 The poly network Authors
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

func TestSyncGenesisHeaderParam(t *testing.T) {
	param := SyncGenesisHeaderParam{
		ChainID:       123,
		GenesisHeader: []byte{1, 2, 3},
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p SyncGenesisHeaderParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, p, param)
}

func TestSyncBlockHeaderParam(t *testing.T) {
	p := SyncBlockHeaderParam{
		ChainID: 123,
		Address: common.ADDRESS_EMPTY,
		Headers: [][]byte{{1, 2, 3}},
	}

	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	var param SyncBlockHeaderParam
	err := param.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	assert.NoError(t, err)

	assert.Equal(t, p, param)
}
