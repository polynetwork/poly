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

	"github.com/polynetwork/poly/common"
	comm "github.com/polynetwork/poly/p2pserver/common"
)

type HeadersReq struct {
	Len       uint8
	HashStart common.Uint256
	HashEnd   common.Uint256
}

//Serialize message payload
func (this *HeadersReq) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint8(this.Len)
	sink.WriteHash(this.HashStart)
	sink.WriteHash(this.HashEnd)
	return nil
}

func (this *HeadersReq) CmdType() string {
	return comm.GET_HEADERS_TYPE
}

//Deserialize message payload
func (this *HeadersReq) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Len, eof = source.NextUint8()
	this.HashStart, eof = source.NextHash()
	this.HashEnd, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
