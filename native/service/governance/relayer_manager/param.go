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

package relayer_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
)

type RegisterRelayerParam struct {
	Pubkey  string
	Address []byte
}

func (this *RegisterRelayerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.Pubkey)
	sink.WriteVarBytes(this.Address)
}

func (this *RegisterRelayerParam) Deserialization(source *common.ZeroCopySource) error {
	pubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize pubkey error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}

	this.Pubkey = pubkey
	this.Address = address
	return nil
}

type RelayerParam struct {
	Pubkey string
}

func (this *RelayerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.Pubkey)
}

func (this *RelayerParam) Deserialization(source *common.ZeroCopySource) error {
	pubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize pubkey error")
	}
	this.Pubkey = pubkey
	return nil
}

type RelayerListParam struct {
	PubkeyList []string
}

func (this *RelayerListParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.PubkeyList)))
	for _, v := range this.PubkeyList {
		sink.WriteString(v)
	}
}

func (this *RelayerListParam) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize PeerPubkeyList length error")
	}
	pubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("source.NextString, deserialize pubkeyList error")
		}
		pubkeyList = append(pubkeyList, k)
	}
	this.PubkeyList = pubkeyList
	return nil
}
