/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/common/serialization"
	"github.com/polynetwork/poly/core/signature"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Uint256
	Height          uint32
	BookkeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           keypair.PublicKey
	Signature       []byte
	PeerId          uint64
	hash            common.Uint256
}

//get the consensus payload hash
func (this *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

//Check whether header is correct
func (this *ConsensusPayload) Verify() error {
	buf := new(bytes.Buffer)
	err := this.SerializeUnsigned(buf)
	if err != nil {
		return err
	}
	err = signature.Verify(this.Owner, buf.Bytes(), this.Signature)
	if err != nil {
		return fmt.Errorf("signature verify error. buf:%v", buf)
	}
	return nil
}

//serialize the consensus payload
func (this *ConsensusPayload) ToArray() []byte {
	b := new(bytes.Buffer)
	err := this.Serialize(b)
	if err != nil {
		log.Errorf("consensus payload serialize error in ToArray(). payload:%v", this)
		return nil
	}
	return b.Bytes()
}

//return inventory type
func (this *ConsensusPayload) InventoryType() common.InventoryType {
	return common.CONSENSUS
}

func (this *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	//return sig.GetHashData(cp)
	return []byte{}
}

func (this *ConsensusPayload) Type() common.InventoryType {

	//TODO:Temporary add for Interface signature.SignableData use.
	return common.CONSENSUS
}

func (this *ConsensusPayload) Serialization(sink *common.ZeroCopySink) error {
	this.serializationUnsigned(sink)
	buf := keypair.SerializePublicKey(this.Owner)
	sink.WriteVarBytes(buf)
	sink.WriteVarBytes(this.Signature)
	return nil
}

//Serialize message payload
func (this *ConsensusPayload) Serialize(w io.Writer) error {
	err := this.SerializeUnsigned(w)
	if err != nil {
		return err
	}
	buf := keypair.SerializePublicKey(this.Owner)
	err = serialization.WriteVarBytes(w, buf)
	if err != nil {
		return fmt.Errorf("write publickey error. publickey buf:%v", buf)
	}
	err = serialization.WriteVarBytes(w, this.Signature)
	if err != nil {
		return fmt.Errorf("write Signature error. Signature:%v", this.Signature)
	}
	return nil
}

//Deserialize message payload
func (this *ConsensusPayload) Deserialization(source *common.ZeroCopySource) error {
	err := this.deserializationUnsigned(source)
	if err != nil {
		return err
	}
	buf, eof := source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Owner, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		return errors.New("deserialize publickey error")
	}
	this.Signature, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

//Deserialize message payload
func (this *ConsensusPayload) Deserialize(r io.Reader) error {
	err := this.DeserializeUnsigned(r)
	if err != nil {
		return err
	}
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {

		return errors.New("read buf error")
	}
	this.Owner, err = keypair.DeserializePublicKey(buf)
	if err != nil {

		return errors.New("deserialize publickey error")
	}
	this.Signature, err = serialization.ReadVarBytes(r)
	if err != nil {

		return errors.New("read Signature error")
	}
	return err
}

func (this *ConsensusPayload) serializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteHash(this.PrevHash)
	sink.WriteUint32(this.Height)
	sink.WriteUint16(this.BookkeeperIndex)
	sink.WriteUint32(this.Timestamp)
	sink.WriteVarBytes(this.Data)
}

//Serialize message payload
func (this *ConsensusPayload) SerializeUnsigned(w io.Writer) error {
	err := serialization.WriteUint32(w, this.Version)
	if err != nil {
		return fmt.Errorf("write error. version:%v", this.Version)
	}
	err = this.PrevHash.Serialize(w)
	if err != nil {
		return fmt.Errorf("serialize error. PrevHash:%v", this.PrevHash)
	}
	err = serialization.WriteUint32(w, this.Height)
	if err != nil {
		return fmt.Errorf("write error. Height:%v", this.Height)
	}
	err = serialization.WriteUint16(w, this.BookkeeperIndex)
	if err != nil {
		return fmt.Errorf("write error. BookkeeperIndex:%v", this.BookkeeperIndex)
	}
	err = serialization.WriteUint32(w, this.Timestamp)
	if err != nil {
		return fmt.Errorf("write error. Timestamp:%v", this.Timestamp)
	}
	err = serialization.WriteVarBytes(w, this.Data)
	if err != nil {
		return fmt.Errorf("write error. Data:%v", this.Data)
	}
	return nil
}

func (this *ConsensusPayload) deserializationUnsigned(source *common.ZeroCopySource) error {
	var eof bool
	this.Version, eof = source.NextUint32()
	this.PrevHash, eof = source.NextHash()
	this.Height, eof = source.NextUint32()
	this.BookkeeperIndex, eof = source.NextUint16()
	this.Timestamp, eof = source.NextUint32()
	this.Data, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

//Deserialize message payload
func (this *ConsensusPayload) DeserializeUnsigned(r io.Reader) error {
	var err error
	this.Version, err = serialization.ReadUint32(r)
	if err != nil {
		return errors.New("read version error")
	}
	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		return errors.New("read preBlock error")
	}
	this.PrevHash = *preBlock
	this.Height, err = serialization.ReadUint32(r)
	if err != nil {
		return errors.New("read Height error")
	}
	this.BookkeeperIndex, err = serialization.ReadUint16(r)
	if err != nil {
		return errors.New("read BookkeeperIndex error")
	}
	this.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		return errors.New("read Timestamp error")
	}

	this.Data, err = serialization.ReadVarBytes(r)
	if err != nil {
		return errors.New("read Data error")
	}
	return nil
}
