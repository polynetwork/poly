/*
 * Copyright (C) 2020 The poly network Authors
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
package fabric

import "github.com/polynetwork/poly/common"

func GetSigArr(raw []byte) [][]byte {
	src := common.NewZeroCopySource(raw)
	res := make([][]byte, 0)
	for {
		val, eof := src.NextVarBytes()
		if eof {
			break
		}
		res = append(res, val)
	}
	return res
}

func SigArrSerialize(arr [][]byte) []byte {
	sink := common.NewZeroCopySink(nil)
	for _, v := range arr {
		sink.WriteVarBytes(v)
	}
	return sink.Bytes()
}
