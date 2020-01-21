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

package node_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
)

type RegisterPeerParam struct {
	PeerPubkey string
	Address    []byte
	Pos        uint64
}

func (this *RegisterPeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address)
	sink.WriteVarUint(this.Pos)
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
	pos, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize pos error")
	}

	this.PeerPubkey = peerPubkey
	this.Address = address
	this.Pos = pos
	return nil
}

type PeerParam struct {
	PeerPubkey string
}

func (this *PeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
}

func (this *PeerParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize peerPubkey error")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type PeerListParam struct {
	PeerPubkeyList []string
}

func (this *PeerListParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.PeerPubkeyList)))
	for _, v := range this.PeerPubkeyList {
		sink.WriteString(v)
	}
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
	this.PeerPubkeyList = peerPubkeyList
	return nil
}

type Configuration struct {
	N                    uint32
	C                    uint32
	K                    uint32
	L                    uint32
	BlockMsgDelay        uint32
	HashMsgDelay         uint32
	PeerHandshakeTimeout uint32
	MaxBlockChangeView   uint32
}

func (this *Configuration) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.N)
	sink.WriteUint32(this.C)
	sink.WriteUint32(this.K)
	sink.WriteUint32(this.L)
	sink.WriteUint32(this.BlockMsgDelay)
	sink.WriteUint32(this.HashMsgDelay)
	sink.WriteUint32(this.PeerHandshakeTimeout)
	sink.WriteUint32(this.MaxBlockChangeView)
}

func (this *Configuration) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize n error")
	}
	c, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize c error")
	}
	k, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize k error")
	}
	l, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize l error")
	}
	blockMsgDelay, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize blockMsgDelay error")
	}
	hashMsgDelay, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize hashMsgDelay error")
	}
	peerHandshakeTimeout, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize peerHandshakeTimeout error")
	}
	maxBlockChangeView, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize maxBlockChangeView error")
	}

	this.N = n
	this.C = c
	this.K = k
	this.L = l
	this.BlockMsgDelay = blockMsgDelay
	this.HashMsgDelay = hashMsgDelay
	this.PeerHandshakeTimeout = peerHandshakeTimeout
	this.MaxBlockChangeView = maxBlockChangeView
	return nil
}

type PreConfig struct {
	Configuration *Configuration
	SetView       uint32
}

func (this *PreConfig) Serialization(sink *common.ZeroCopySink) {
	this.Configuration.Serialization(sink)
	sink.WriteUint32(this.SetView)
}

func (this *PreConfig) Deserialization(source *common.ZeroCopySource) error {
	config := new(Configuration)
	err := config.Deserialization(source)
	if err != nil {
		return fmt.Errorf("config.Deserialization, deserialize configuration error: %v", err)
	}
	setView, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize setView error: %v", err)
	}
	this.Configuration = config
	this.SetView = setView
	return nil
}

type GlobalParam struct {
	MinInitStake uint32 //min init pos
	CandidateNum uint32 //num of candidate and consensus node
}

func (this *GlobalParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.MinInitStake)
	sink.WriteUint32(this.CandidateNum)
}

func (this *GlobalParam) Deserialization(source *common.ZeroCopySource) error {
	minInitStake, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("utils.ReadVarUint, deserialize minInitStake error")
	}
	candidateNum, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateNum error")
	}
	this.MinInitStake = minInitStake
	this.CandidateNum = candidateNum
	return nil
}
