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
	"bytes"
	"crypto/rand"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/serialization"
	"testing"
)

func BenchmarkZeroCopySource(b *testing.B) {
	const N = 12000
	buf := make([]byte, N)
	rand.Read(buf)

	for i := 0; i < b.N; i++ {
		source := common.NewZeroCopySource(buf)
		for j := 0; j < N/100; j++ {
			source.NextUint16()
			source.NextByte()
			source.NextUint64()
			source.NextVarUint()
			source.NextBytes(20)
		}
	}

}

func BenchmarkDerserialize(b *testing.B) {
	const N = 12000
	buf := make([]byte, N)
	rand.Read(buf)

	for i := 0; i < b.N; i++ {
		reader := bytes.NewBuffer(buf)
		for j := 0; j < N/100; j++ {
			serialization.ReadUint16(reader)
			serialization.ReadByte(reader)
			serialization.ReadUint64(reader)
			serialization.ReadVarUint(reader, 0)
			serialization.ReadBytes(reader, 20)
		}
	}

}
