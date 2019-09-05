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

package header_sync

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

type SyncBlockHeaderParam struct {
	Address common.Address
	Headers [][]byte
}

func (this *SyncBlockHeaderParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.Address)
	sink.WriteUint64(uint64(len(this.Headers)))
	for _, v := range this.Headers {
		sink.WriteVarBytes(v)
	}
}

func (this *SyncBlockHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error:%s", err)
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error")
	}
	var headers [][]byte
	for i := 0; uint64(i) < n; i++ {
		header, err := utils.DecodeVarBytes(source)
		if err != nil {
			return fmt.Errorf("utils.DecodeVarBytes, deserialize header error: %v", err)
		}
		headers = append(headers, header)
	}
	this.Address = address
	this.Headers = headers
	return nil
}

type SyncConsensusPeerParam struct {
	Address common.Address
	Header  []byte
	Proof   string
}

func (this *SyncConsensusPeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.Address)
	sink.WriteVarBytes(this.Header)
	sink.WriteVarBytes([]byte(this.Proof))
}

func (this *SyncConsensusPeerParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error:%s", err)
	}
	header, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize header error:%s", err)
	}
	proof, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize proof count error:%s", err)
	}
	this.Address = address
	this.Header = header
	this.Proof = proof
	return nil
}

type SyncGenesisHeaderParam struct {
	GenesisHeader []byte
}

func (this *SyncGenesisHeaderParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(this.GenesisHeader)
	return nil
}

func (this *SyncGenesisHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	genesisHeader, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisHeader count error:%s", err)
	}
	this.GenesisHeader = genesisHeader
	return nil
}
