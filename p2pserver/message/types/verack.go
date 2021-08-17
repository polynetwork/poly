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

package types

import (
	"io"

	comm "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/p2pserver/common"
)

type VerACK struct {
	IsConsensus bool
}

//Serialize message payload
func (this *VerACK) Serialization(sink *comm.ZeroCopySink) error {
	sink.WriteBool(this.IsConsensus)
	return nil
}

func (this *VerACK) CmdType() string {
	return common.VERACK_TYPE
}

//Deserialize message payload
func (this *VerACK) Deserialization(source *comm.ZeroCopySource) error {
	var eof bool
	this.IsConsensus, eof = source.NextBool()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
