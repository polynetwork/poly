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

package side_chain_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"sort"
)

type SideChain struct {
	ChainId      uint64
	Router       uint64
	Name         string
	BlocksToWait uint64
}

func (this *SideChain) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.ChainId)
	sink.WriteUint64(this.Router)
	sink.WriteString(this.Name)
	sink.WriteUint64(this.BlocksToWait)
	return nil
}

func (this *SideChain) Deserialization(source *common.ZeroCopySource) error {
	chainId, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	router, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize router error")
	}
	name, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize name error")
	}
	blocksToWait, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize blocksToWait error")
	}

	this.ChainId = chainId
	this.Router = router
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type BindSignInfo struct {
	BindSignInfo map[string][]byte
}

func (this *BindSignInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.BindSignInfo)))
	var BindSignInfoList []string
	for k := range this.BindSignInfo {
		BindSignInfoList = append(BindSignInfoList, k)
	}
	sort.SliceStable(BindSignInfoList, func(i, j int) bool {
		return BindSignInfoList[i] > BindSignInfoList[j]
	})
	for _, k := range BindSignInfoList {
		sink.WriteString(k)
		sink.WriteVarBytes(this.BindSignInfo[k])
	}
}

func (this *BindSignInfo) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("BindSignInfo deserialize MultiSignInfo length error")
	}
	bindSignInfo := make(map[string][]byte)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("BindSignInfo deserialize public key error")
		}
		v, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("BindSignInfo deserialize byte error")
		}
		bindSignInfo[k] = v
	}
	this.BindSignInfo = bindSignInfo
	return nil
}

type ContractBinded struct {
	Contract []byte
	Ver      uint64
}

func (this *ContractBinded) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Contract)
	sink.WriteUint64(this.Ver)
}

func (this *ContractBinded) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Contract, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("BindContract deserialize contract error")
	}
	this.Ver, eof = source.NextUint64()
	if eof {
		return fmt.Errorf("BindContract deserialize version error")
	}
	return nil
}
