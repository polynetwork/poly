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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint64Slice(t *testing.T) {
	data1 := []uint64{3, 2, 4, 1}
	data2 := []uint64{1, 2, 3, 4}

	common.SortUint64s(data1)

	assert.Equal(t, data1, data2)

}
