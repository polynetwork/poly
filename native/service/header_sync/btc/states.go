/*
 * Copyright (C) 2021 The poly network Authors
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
	"bytes"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"math/big"
)

type StoredHeader struct {
	Header    wire.BlockHeader
	Height    uint32
	totalWork *big.Int
}

/*----- header serialization ------- */
/* byteLength   desc          at offset
   80	       header	           0
    4	       height             80
   32	       total work         84
*/

func (this *StoredHeader) Serialization(sink *common.ZeroCopySink) {
	buf := bytes.NewBuffer(nil)
	this.Header.Serialize(buf)
	sink.WriteVarBytes(buf.Bytes())
	sink.WriteUint32(this.Height)
	biBytes := this.totalWork.Bytes()
	pad := make([]byte, 32-len(biBytes))
	//serializedBI := append(pad, biBytes...)
	sink.WriteVarBytes(append(pad, biBytes...))
}

func (this *StoredHeader) Deserialization(source *common.ZeroCopySource) error {
	buf, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("StoredHeader get header bytes error")
	}
	blockHeader := new(wire.BlockHeader)
	err := blockHeader.Deserialize(bytes.NewBuffer(buf))
	if err != nil {
		log.Error("deserialize wire.blockheader error: ", err)
		return fmt.Errorf("StoredHeader deserialize header error")
	}
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("StoredHeader get height error")
	}
	totalWorkBytes, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("StoredHeader get total work bytes error")
	}
	totalWork := new(big.Int)
	totalWork.SetBytes(totalWorkBytes)
	this.Header = *blockHeader
	this.Height = height
	this.totalWork = totalWork
	return nil
}
