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
	"fmt"
	"io"

	"github.com/polynetwork/poly/common"
	ct "github.com/polynetwork/poly/core/types"
	comm "github.com/polynetwork/poly/p2pserver/common"
)

type BlkHeader struct {
	BlkHdr []*ct.Header
}

//Serialize message payload
func (this BlkHeader) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(uint32(len(this.BlkHdr)))

	for _, header := range this.BlkHdr {
		err := header.Serialization(sink)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *BlkHeader) CmdType() string {
	return comm.HEADERS_TYPE
}

//Deserialize message payload
func (this *BlkHeader) Deserialization(source *common.ZeroCopySource) error {
	var count uint32
	count, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(count); i++ {
		var headers ct.Header
		err := headers.Deserialization(source)
		if err != nil {
			return fmt.Errorf("deserialze BlkHeader error: %v", err)
		}
		this.BlkHdr = append(this.BlkHdr, &headers)
	}
	return nil
}
