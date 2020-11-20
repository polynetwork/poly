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

import (
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/common/attrmgr"
	"github.com/polynetwork/poly/common"
	"github.com/tjfoc/gmsm/sm2"
)

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

func getAttributesFromCert(cert *sm2.Certificate) (*attrmgr.Attributes, error) {
	for _, ext := range cert.Extensions {
		if isAttrOID(ext.Id) {
			attrs := &attrmgr.Attributes{}
			err := json.Unmarshal(ext.Value, attrs)
			if err != nil {
				return nil, fmt.Errorf("Failed to unmarshal attributes from certificate: %v", err)
			}
			return attrs, nil
		}
	}
	return nil, nil
}

func isAttrOID(oid asn1.ObjectIdentifier) bool {
	if len(oid) != len(attrmgr.AttrOID) {
		return false
	}
	for idx, val := range oid {
		if val != attrmgr.AttrOID[idx] {
			return false
		}
	}
	return true
}
