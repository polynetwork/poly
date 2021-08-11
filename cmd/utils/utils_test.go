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

package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenExportBlocksFileName(t *testing.T) {
	name := "blocks.dat"
	start := uint32(0)
	end := uint32(100)
	fileName := GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.dat", fileName)
	name = "blocks"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100", fileName)
	name = "blocks."
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.", fileName)
	name = "blocks.export.dat"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks.export_0_100.dat", fileName)
}
