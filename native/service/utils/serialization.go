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

package utils

import (
	"fmt"
	"io"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/serialization"
)

func WriteAddress(w io.Writer, address common.Address) error {
	if err := serialization.WriteVarBytes(w, address[:]); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadAddress(r io.Reader) (common.Address, error) {
	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return common.Address{}, fmt.Errorf("[State] deserialize from error:%v", err)
	}
	return common.AddressParseFromBytes(from)
}

func DecodeAddress(source *common.ZeroCopySource) (common.Address, error) {
	from, eof := source.NextVarBytes()
	if eof {
		return common.Address{}, io.ErrUnexpectedEOF
	}

	return common.AddressParseFromBytes(from)
}

func DecodeVarBytes(source *common.ZeroCopySource) ([]byte, error) {
	v, eof := source.NextVarBytes()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	return v, nil
}

func DecodeString(source *common.ZeroCopySource) (string, error) {
	str, eof := source.NextString()
	if eof {
		return "", io.ErrUnexpectedEOF
	}

	return str, nil
}

func EncodeUint256(sink *common.ZeroCopySink, hash common.Uint256) (size uint64) {
	return sink.WriteVarBytes(hash[:])
}

func DecodeUint256(source *common.ZeroCopySource) (common.Uint256, error) {
	from, eof := source.NextVarBytes()
	if eof {
		return common.Uint256{}, io.ErrUnexpectedEOF
	}

	return common.Uint256ParseFromBytes(from)
}
