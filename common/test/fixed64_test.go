/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */
package test

import (
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFixed64_Serialize(t *testing.T) {
	val := common.Fixed64(10)
	buf := common.NewZeroCopySink(nil)
	val.Serialization(buf)
	val2 := common.Fixed64(0)
	val2.Deserialization(common.NewZeroCopySource(buf.Bytes()))

	assert.Equal(t, val, val2)
}

func TestFixed64_Deserialize(t *testing.T) {
	val := common.Fixed64(0)
	err := val.Deserialization(common.NewZeroCopySource([]byte{1, 2, 3}))

	assert.NotNil(t, err)

}
