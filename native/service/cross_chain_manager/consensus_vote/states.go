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

package consensus_vote

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	"sort"
)

type VoteInfo struct {
	Status   bool
	VoteInfo map[string]bool
}

func (this *VoteInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBool(this.Status)
	sink.WriteUint64(uint64(len(this.VoteInfo)))
	VoteInfoList := make([]string, 0, len(this.VoteInfo))
	for k := range this.VoteInfo {
		VoteInfoList = append(VoteInfoList, k)
	}
	sort.SliceStable(VoteInfoList, func(i, j int) bool {
		return VoteInfoList[i] > VoteInfoList[j]
	})
	for _, k := range VoteInfoList {
		sink.WriteString(k)
		v := this.VoteInfo[k]
		sink.WriteBool(v)
	}
}

func (this *VoteInfo) Deserialization(source *common.ZeroCopySource) error {
	status, eof := source.NextBool()
	if eof {
		return fmt.Errorf("VoteInfo deserialize status length error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("VoteInfo deserialize VoteInfo length error")
	}
	voteInfo := make(map[string]bool)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("VoteInfo deserialize key error")
		}
		v, eof := source.NextBool()
		if eof {
			return fmt.Errorf("VoteInfo deserialize value error")
		}
		voteInfo[k] = v
	}
	this.Status = status
	this.VoteInfo = voteInfo
	return nil
}
