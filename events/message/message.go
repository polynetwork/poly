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

package message

import (
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/types"
)

const (
	TOPIC_SAVE_BLOCK_COMPLETE       = "svblkcmp"
	TOPIC_NEW_INVENTORY             = "newinv"
	TOPIC_NODE_DISCONNECT           = "noddis"
	TOPIC_NODE_CONSENSUS_DISCONNECT = "nodcnsdis"
	TOPIC_SMART_CODE_EVENT          = "scevt"
)

type SaveBlockCompleteMsg struct {
	Block *types.Block
}

type NewInventoryMsg struct {
	Inventory *common.Inventory
}

type SmartCodeEventMsg struct {
	Event *types.SmartCodeEvent
}

type BlockConsensusComplete struct {
	Block *types.Block
}
