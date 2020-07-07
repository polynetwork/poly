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

package btc

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestStoredHeader(t *testing.T) {

	totalWork, _ := (new(big.Int)).SetString("123", 10)

	header := StoredHeader{
		Header: wire.BlockHeader{
			Version:    1,
			PrevBlock:  chainhash.Hash{1, 2, 3},
			MerkleRoot: chainhash.Hash{2, 2, 3},
			Timestamp:  time.Unix(time.Now().Unix(), 0),
			Nonce:      100,
			Bits:       200,
		},
		Height:    uint32(100),
		totalWork: totalWork,
	}
	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)

	h := new(StoredHeader)
	err := h.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, header, *h)
}
