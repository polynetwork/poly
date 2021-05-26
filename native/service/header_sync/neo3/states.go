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

package neo3

import (
	"fmt"
	"github.com/joeqian10/neo3-gogogo/block"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/helper"
	"github.com/joeqian10/neo3-gogogo/io"
	"github.com/joeqian10/neo3-gogogo/mpt"
	"github.com/polynetwork/poly/common"
)

type NeoConsensus struct {
	ChainID       uint64
	Height        uint32
	NextConsensus *helper.UInt160
}

func (this *NeoConsensus) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes(this.NextConsensus.ToByteArray())
}

func (this *NeoConsensus) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	if this.ChainID, eof = source.NextUint64(); eof {
		return fmt.Errorf("NeoConsensus.Deserialization, ChainID NextUint64 error")
	}
	if this.Height, eof = source.NextUint32(); eof {
		return fmt.Errorf("NeoConsensus.Deserialization, Height NextUint32 error")
	}

	nextConsensusBs, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("NeoConsensus.Deserialization, NextConsensus NextVarBytes error")
	}

	this.NextConsensus = helper.UInt160FromBytes(nextConsensusBs);
	return nil
}

type NeoBlockHeader struct {
	*block.Header
}

func (this *NeoBlockHeader) Deserialization(source *common.ZeroCopySource) error {
	this.Header = block.NewBlockHeader()
	br := io.NewBinaryReaderFromBuf(source.Bytes())
	this.Header.Deserialize(br)
	if br.Err != nil {
		return fmt.Errorf("neo3-gogogo Header Deserialize error: %s", br.Err)
	}
	return nil
}

func (this *NeoBlockHeader) Serialization(sink *common.ZeroCopySink) error {
	bw := io.NewBufBinaryWriter()
	this.Serialize(bw.BinaryWriter)
	if bw.Err != nil {
		return fmt.Errorf("neo3-gogogo Header Serialize error: %s", bw.Err)
	}
	sink.WriteBytes(bw.Bytes())
	return nil
}

func (this *NeoBlockHeader) GetMessage(magic uint32) ([]byte, error) {
	buff2 := io.NewBufBinaryWriter()
	this.SerializeUnsigned(buff2.BinaryWriter)
	if buff2.Err != nil {
		return nil, fmt.Errorf("neo3-gogogo Header SerializeUnsigned error: %s", buff2.Err)
	}
	hash := helper.UInt256FromBytes(crypto.Sha256(buff2.Bytes()))

	buf := io.NewBufBinaryWriter()
	buf.BinaryWriter.WriteLE(magic)
	buf.BinaryWriter.WriteLE(hash)
	if buf.Err != nil {
		return nil, fmt.Errorf("NeoBlockHeader.GetMessage write hash error: %s", buf.Err)
	}
	return buf.Bytes(), nil
}

type NeoCrossChainMsg struct {
	*mpt.StateRoot
}

func (this *NeoCrossChainMsg) Deserialization(source *common.ZeroCopySource) error {
	this.StateRoot = new(mpt.StateRoot)
	br := io.NewBinaryReaderFromBuf(source.Bytes())
	this.Deserialize(br)
	if br.Err != nil {
		return fmt.Errorf("neo3-gogogo mpt.StateRoot Deserialize error: %s", br.Err)
	}
	return nil
}

func (this *NeoCrossChainMsg) Serialization(sink *common.ZeroCopySink) error {
	bw := io.NewBufBinaryWriter()
	this.Serialize(bw.BinaryWriter)
	if bw.Err != nil {
		return fmt.Errorf("neo3-gogogo mpt.StateRoot Serialize error: %s", bw.Err)
	}
	sink.WriteBytes(bw.Bytes())
	return nil
}

func (this *NeoCrossChainMsg) GetScriptHash() (*helper.UInt160, error) {
	if len(this.Witnesses) == 0 {
		return nil, fmt.Errorf("NeoCrossChainMsg.Witness incorrect length")
	}
	verificationScriptBs, err := crypto.Base64Decode(this.Witnesses[0].Verification) // base64
	if err != nil {
		return nil, fmt.Errorf("NeoCrossChainMsg.Witness.Verification decode error: %s", err)
	}
	if len(verificationScriptBs) == 0 {
		return nil, fmt.Errorf("NeoCrossChainMsg.Witness.VerificationScript is empty")
	}
	scriptHash := helper.UInt160FromBytes(crypto.Hash160(verificationScriptBs))
	return scriptHash, nil
}

func (this *NeoCrossChainMsg) GetMessage(magic uint32) ([]byte, error) {
	buff2 := io.NewBufBinaryWriter()
	this.SerializeUnsigned(buff2.BinaryWriter)
	if buff2.Err != nil {
		return nil, fmt.Errorf("neo3-gogogo mpt.StateRoot SerializeUnsigned error: %s", buff2.Err)
	}
	hash := helper.UInt256FromBytes(crypto.Sha256(buff2.Bytes()))

	buf := io.NewBufBinaryWriter()
	buf.BinaryWriter.WriteLE(magic)
	buf.BinaryWriter.WriteLE(hash)
	if buf.Err != nil {
		return nil, fmt.Errorf("NeoCrossChainMsg.GetMessage write hash error: %s", buf.Err)
	}
	return buf.Bytes(), nil
}
