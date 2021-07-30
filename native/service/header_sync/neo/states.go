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

package neo

import (
	"encoding/hex"
	"fmt"
	"github.com/joeqian10/neo-gogogo/block"
	"github.com/joeqian10/neo-gogogo/crypto"
	"github.com/joeqian10/neo-gogogo/helper"
	"github.com/joeqian10/neo-gogogo/helper/io"
	"github.com/joeqian10/neo-gogogo/mpt"
	"github.com/polynetwork/poly/common"
)

type NeoConsensus struct {
	ChainID       uint64
	Height        uint32
	NextConsensus helper.UInt160
}

func (this *NeoConsensus) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes(this.NextConsensus.Bytes())
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

	var err error
	if this.NextConsensus, err = helper.UInt160FromBytes(nextConsensusBs); err != nil {
		return fmt.Errorf("NeoConsensus.Deserialization, NextConsensus UInt160FromBytes error: %s", err)
	}
	return nil
}

type NeoBlockHeader struct {
	*block.BlockHeader
}

func (this *NeoBlockHeader) Deserialization(source *common.ZeroCopySource) error {
	this.BlockHeader = new(block.BlockHeader)
	br := io.NewBinaryReaderFromBuf(source.Bytes())
	this.BlockHeader.Deserialize(br)
	if br.Err != nil {
		return fmt.Errorf("joeqian10/neo-gogogo/block.BlockHeader Deserialize error: %s", br.Err)
	}
	return nil
}

func (this *NeoBlockHeader) Serialization(sink *common.ZeroCopySink) error {
	bw := io.NewBufBinaryWriter()
	this.Serialize(bw.BinaryWriter)
	if bw.Err != nil {
		return fmt.Errorf("joeqian10/neo-gogogo/block.BlockHeader Serialize error: %s", bw.Err)
	}
	sink.WriteBytes(bw.Bytes())
	return nil
}

func (this *NeoBlockHeader) GetMessage() ([]byte, error) {
	buf := io.NewBufBinaryWriter()
	this.SerializeUnsigned(buf.BinaryWriter)
	if buf.Err != nil {
		return nil, fmt.Errorf("GetHashData of NeoBlockHeader neo-gogogo block.BlockHeader SerializeUnsigned error: %s", buf.Err)
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
		return fmt.Errorf("neo-gogogo mpt.StateRoot Deserialize error: %s", br.Err)
	}
	return nil
}

func (this *NeoCrossChainMsg) Serialization(sink *common.ZeroCopySink) error {
	bw := io.NewBufBinaryWriter()
	this.Serialize(bw.BinaryWriter)
	if bw.Err != nil {
		return fmt.Errorf("neo-gogogo mpt.StateRoot Serialize error: %s", bw.Err)
	}
	sink.WriteBytes(bw.Bytes())
	return nil
}

func (this *NeoCrossChainMsg) GetScriptHash() (helper.UInt160, error) {
	verificationScriptBs, err := hex.DecodeString(this.Witness.VerificationScript)
	if err != nil {
		return helper.UInt160{}, fmt.Errorf("NeoCrossChainMsg.Witness.VerificationScript decode to bytes error: %s", err)
	}
	if len(verificationScriptBs) == 0 {
		return helper.UInt160{}, fmt.Errorf("NeoCrossChainMsg.Witness.VerificationScript length is 0 ")
	}
	scriptHash, err := helper.UInt160FromBytes(crypto.Hash160(verificationScriptBs))
	if err != nil {
		return helper.UInt160{}, fmt.Errorf("neo-gogogo tx.Witness GetScriptHash error: %s", err)
	}
	return scriptHash, nil
}

func (this *NeoCrossChainMsg) GetMessage() ([]byte, error) {
	buf := io.NewBufBinaryWriter()
	this.SerializeUnsigned(buf.BinaryWriter)
	if buf.Err != nil {
		return nil, fmt.Errorf("GetHashData of NeoBlockHeader neo-gogogo mpt.StateRoot SerializeUnsigned error: %s", buf.Err)
	}
	return buf.Bytes(), nil
}
