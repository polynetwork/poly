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
package relayer_manager

import (
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelayerListParam_Serialization(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = []common.Address{{1, 2, 4, 6}, {1, 4, 5, 7}, {1, 3, 5, 7, 9}}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	source := common.NewZeroCopySource(sink.Bytes())
	var p RelayerListParam
	err := p.Deserialization(source)
	assert.Nil(t, err)
}
