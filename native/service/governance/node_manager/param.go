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

package node_manager

import (
	"fmt"
	"github.com/polynetwork/poly/common"
)

type RegisterPeerParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *RegisterPeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
}

func (this *RegisterPeerParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize peerPubkey error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}

	this.PeerPubkey = peerPubkey
	this.Address = addr
	return nil
}

type PeerParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *PeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
}

func (this *PeerParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize peerPubkey error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}

	this.PeerPubkey = peerPubkey
	this.Address = addr
	return nil
}

type PeerListParam struct {
	PeerPubkeyList []string
	Address        common.Address
}

func (this *PeerListParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.PeerPubkeyList)))
	for _, v := range this.PeerPubkeyList {
		sink.WriteString(v)
	}
	sink.WriteVarBytes(this.Address[:])
}

func (this *PeerListParam) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize PeerPubkeyList length error")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("source.NextString, deserialize peerPubkey error")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}

	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	this.PeerPubkeyList = peerPubkeyList
	this.Address = addr
	return nil
}

type UpdateConfigParam struct {
	Configuration *Configuration
}

func (this *UpdateConfigParam) Serialization(sink *common.ZeroCopySink) {
	this.Configuration.Serialization(sink)
}

func (this *UpdateConfigParam) Deserialization(source *common.ZeroCopySource) error {
	configuration := new(Configuration)
	err := configuration.Deserialization(source)
	if err != nil {
		return fmt.Errorf("configuration.Deserialization, deserialize configuration error: %s", err)
	}
	this.Configuration = configuration
	return nil
}
