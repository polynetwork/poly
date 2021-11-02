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

package common

type InventoryType byte

const (
	TRANSACTION InventoryType = 0x01
	BLOCK       InventoryType = 0x02
	CONSENSUS   InventoryType = 0xe0
)

//TODO: temp inventory
type Inventory interface {
	//sig.SignableData
	Hash() Uint256
	Verify() error
	Type() InventoryType
}
