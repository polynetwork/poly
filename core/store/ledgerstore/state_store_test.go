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

package ledgerstore

import (
	"math/rand"
	"testing"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/merkle"
	"github.com/stretchr/testify/assert"
)

func TestStateMerkleRoot(t *testing.T) {
	teststatemerkleroot := func(H, effectiveStateHashHeight uint32) {
		diffHashes := make([]common.Uint256, 0, H)
		for i := uint32(0); i < H; i++ {
			var hash common.Uint256
			rand.Read(hash[:])
			diffHashes = append(diffHashes, hash)
		}
		db := NewMemStateStore(effectiveStateHashHeight)
		for h, hash := range diffHashes[:effectiveStateHashHeight] {
			height := uint32(h)
			db.NewBatch()
			err := db.AddStateMerkleTreeRoot(height, hash)
			assert.Nil(t, err)
			db.CommitTo()
			root, _ := db.GetStateMerkleRoot(height)
			assert.Equal(t, root, common.UINT256_EMPTY)
		}

		merkleTree := merkle.NewTree(0, nil, nil)
		for h, hash := range diffHashes[effectiveStateHashHeight:] {
			height := uint32(h) + effectiveStateHashHeight
			merkleTree.Append(hash.ToArray())
			root1 := db.GetStateMerkleRootWithNewHash(hash)
			db.NewBatch()
			err := db.AddStateMerkleTreeRoot(height, hash)
			assert.Nil(t, err)
			db.CommitTo()
			root2, _ := db.GetStateMerkleRoot(height)
			root3 := merkleTree.Root()

			assert.Equal(t, root1, root2)
			assert.Equal(t, root1, root3)
		}
	}

	for i := 0; i < 200; i++ {
		teststatemerkleroot(1024, uint32(i))
		h := rand.Uint32()%1000 + 1
		eff := rand.Uint32() % h
		teststatemerkleroot(h, eff)
	}

}
