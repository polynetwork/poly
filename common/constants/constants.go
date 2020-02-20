/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package constants

import (
	"time"
)

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(time.Date(2018, time.June, 30, 0, 0, 0, 0, time.UTC).Unix())
)

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 16

// transaction constants
const TX_MAX_SIG_SIZE = 16

// network magic number
const (
	NETWORK_MAGIC_MAINNET = 0x8c77ab60
	NETWORK_MAGIC_TESTNET = 0x2d8829df
)

// ont constants
const (
	ONT_NAME         = "ONT Token"
	ONT_SYMBOL       = "ONT"
	ONT_DECIMALS     = 0
	ONT_TOTAL_SUPPLY = uint64(1000000000)
)

// ong constants
const (
	ONG_NAME         = "ONG Token"
	ONG_SYMBOL       = "ONG"
	ONG_DECIMALS     = 9
	ONG_TOTAL_SUPPLY = uint64(1000000000000000000)
)
