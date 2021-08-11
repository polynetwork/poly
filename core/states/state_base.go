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

package states

import (
	"fmt"
	"io"

	"github.com/polynetwork/poly/common/serialization"
)

type StateBase struct {
	StateVersion byte
}

func (this *StateBase) Serialize(w io.Writer) error {
	serialization.WriteByte(w, this.StateVersion)
	return nil
}

func (this *StateBase) Deserialize(r io.Reader) error {
	stateVersion, err := serialization.ReadByte(r)
	if err != nil {
		return fmt.Errorf("[StateBase], StateBase Deserialize failed,%s", err)
	}
	this.StateVersion = stateVersion
	return nil
}
