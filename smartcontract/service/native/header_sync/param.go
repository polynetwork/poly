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
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type SyncBlockHeaderParam struct {
	Address common.Address
	Headers [][]byte
}

func (this *SyncBlockHeaderParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarUint(sink, uint64(len(this.Headers)))
	for _, v := range this.Headers {
		utils.EncodeVarBytes(sink, v)
	}
}

func (this *SyncBlockHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error:%s", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error:%s", err)
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
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarBytes(sink, this.Header)
	utils.EncodeString(sink, this.Proof)
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
