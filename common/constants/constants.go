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

package constants

import (
	"time"
)

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(time.Date(2020, time.August, 10, 0, 0, 0, 0, time.UTC).Unix())
)

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 16

// transaction constants
const TX_MAX_SIG_SIZE = 16

// network magic number
const (
	NETWORK_MAGIC_MAINNET = 0x8c6077ab
	NETWORK_MAGIC_TESTNET = 0x2ddf8829
)

// extra info change height
const EXTRA_INFO_HEIGHT_MAINNET = 2917744
const EXTRA_INFO_HEIGHT_TESTNET = 1664798

// eth 1559 height
const ETH1559_HEIGHT_MAINNET = 12965000
const ETH1559_HEIGHT_TESTNET = 10499401

// heco 120 height
const HECO120_HEIGHT_MAINNET = 8606000
const HECO120_HEIGHT_TESTNET = 8290000

const POLYGON_SNAP_CHAINID_MAINNET = 16
