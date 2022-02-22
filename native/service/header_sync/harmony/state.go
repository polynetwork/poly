/*
 * Copyright (C) 2022 The poly network Authors
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

package harmony

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/harmony-one/harmony/block"
	"github.com/harmony-one/harmony/shard"
)

var (
	// Harmony staking epoch
	stakingEpoch uint64 = 186
)

// Harmony Epoch
type Epoch struct {
	ID uint64
	Committee *shard.Committee
	StartHeight uint64
}

// Harmony Header with Signature
type HeaderWithSig struct {
	Header *block.Header
	Sig hexutil.Bytes
	Bitmap hexutil.Bytes
}