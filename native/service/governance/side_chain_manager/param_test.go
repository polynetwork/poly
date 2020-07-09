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
package side_chain_manager

import (
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterSideChain(t *testing.T) {
	param := RegisterSideChainParam{
		Address:      common.Address{1, 2, 3},
		ChainId:      123,
		Name:         "123456",
		BlocksToWait: 1234,
	}
	sink := common.NewZeroCopySink(nil)
	err := param.Serialization(sink)
	assert.NoError(t, err)

	source := common.NewZeroCopySource(sink.Bytes())
	var p RegisterSideChainParam
	err = p.Deserialization(source)
	assert.NoError(t, err)

	assert.Equal(t, param, p)
}

func TestChainidParam(t *testing.T) {
	p := ChainidParam{
		Chainid: 123,
	}

	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	var param ChainidParam
	err := param.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, p, param)
}
