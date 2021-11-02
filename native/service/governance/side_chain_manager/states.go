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

package side_chain_manager

import (
	"fmt"
	"sort"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/core/ledger"
)

type SideChain struct {
	Address      common.Address
	ChainId      uint64
	Router       uint64
	Name         string
	BlocksToWait uint64
	CCMCAddress  []byte
	ExtraInfo    []byte
}

func (this *SideChain) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(this.Address[:])
	sink.WriteVarUint(this.ChainId)
	sink.WriteVarUint(this.Router)
	sink.WriteVarBytes([]byte(this.Name))
	sink.WriteVarUint(this.BlocksToWait)
	sink.WriteVarBytes(this.CCMCAddress)
	height := config.GetExtraInfoHeight(config.DefConfig.P2PNode.NetworkId)
	if !config.EXTRA_INFO_HEIGHT_FORK_CHECK || ledger.DefLedger.GetCurrentBlockHeight() >= height {
		sink.WriteVarBytes(this.ExtraInfo)
	}
	return nil
}

func (this *SideChain) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("utils.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	chainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize chainid error")
	}
	router, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize router error")
	}
	name, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize name error")
	}
	blocksToWait, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize blocksToWait error")
	}
	CCMCAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize CCMCAddress error")
	}
	ExtraInfo, _ := source.NextVarBytes()

	this.Address = addr
	this.ChainId = chainId
	this.Router = router
	this.Name = name
	this.BlocksToWait = blocksToWait
	this.CCMCAddress = CCMCAddress
	this.ExtraInfo = ExtraInfo
	return nil
}

type BindSignInfo struct {
	BindSignInfo map[string][]byte
}

func (this *BindSignInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.BindSignInfo)))
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
	n, eof := source.NextVarUint()
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
