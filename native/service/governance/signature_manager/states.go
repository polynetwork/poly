/*
 * Copyright (C) 2022 The poly network Authors
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
package signature_manager

import (
	"fmt"
	"sort"

	"github.com/polynetwork/poly/common"
)

type SigInfo struct {
	Status  bool
	SigInfo map[string][]byte
}

func (this *SigInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBool(this.Status)
	sink.WriteUint64(uint64(len(this.SigInfo)))
	sigInfoList := make([]string, 0, len(this.SigInfo))
	for k := range this.SigInfo {
		sigInfoList = append(sigInfoList, k)
	}
	sort.SliceStable(sigInfoList, func(i, j int) bool {
		return sigInfoList[i] > sigInfoList[j]
	})
	for _, k := range sigInfoList {
		sink.WriteString(k)
		v := this.SigInfo[k]
		sink.WriteVarBytes(v)
	}
}

func (this *SigInfo) Deserialization(source *common.ZeroCopySource) error {
	status, eof := source.NextBool()
	if eof {
		return fmt.Errorf("SigInfo deserialize status length error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SigInfo deserialize SigInfo length error")
	}
	sigInfo := make(map[string][]byte)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("SigInfo deserialize key error")
		}
		v, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("SigInfo deserialize value error")
		}
		sigInfo[k] = v
	}
	this.Status = status
	this.SigInfo = sigInfo
	return nil
}
