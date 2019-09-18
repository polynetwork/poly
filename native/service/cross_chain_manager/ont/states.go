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

package ont

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
)

type FromMerkleValue struct {
	TxHash                   common.Uint256
	CreateCrossChainTxMerkle *CreateCrossChainTxMerkle
}

func (this *FromMerkleValue) Serialization(sink *common.ZeroCopySink) {
	sink.WriteHash(this.TxHash)
	this.CreateCrossChainTxMerkle.Serialization(sink)
}

func (this *FromMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextHash()
	if eof {
		return fmt.Errorf("MerkleValue deserialize txHash error")
	}
	createCrossChainTxMerkle := new(CreateCrossChainTxMerkle)
	err := createCrossChainTxMerkle.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize createCrossChainTxMerkle error:%s", err)
	}

	this.TxHash = txHash
	this.CreateCrossChainTxMerkle = createCrossChainTxMerkle
	return nil
}

type ToMerkleValue struct {
	TxHash            common.Uint256
	ToContractAddress string
	MakeTxParam       *crosscommon.MakeTxParam
}

func (this *ToMerkleValue) Serialization(sink *common.ZeroCopySink) {
	sink.WriteHash(this.TxHash)
	sink.WriteVarBytes([]byte(this.ToContractAddress))
	this.MakeTxParam.Serialization(sink)
}

func (this *ToMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextHash()
	if eof {
		return fmt.Errorf("MerkleValue deserialize txHash error")
	}
	toContractAddress, eof := source.NextString()
	if eof {
		return fmt.Errorf("MerkleValue deserialize toContractAddress error")
	}
	makeTxParam := new(crosscommon.MakeTxParam)
	err := makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize makeTxParam error:%s", err)
	}

	this.TxHash = txHash
	this.ToContractAddress = toContractAddress
	this.MakeTxParam = makeTxParam
	return nil
}

type CreateCrossChainTxMerkle struct {
	FromChainID         uint64
	FromContractAddress string
	ToChainID           uint64
	Fee                 uint64
	Method              string
	Args                []byte
}

func (this *CreateCrossChainTxMerkle) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.FromChainID)
	sink.WriteVarBytes([]byte(this.FromContractAddress))
	sink.WriteUint64(this.ToChainID)
	sink.WriteUint64(this.Fee)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes(this.Args)
}

func (this *CreateCrossChainTxMerkle) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize fromChainID error")
	}
	fromContractAddress, eof := source.NextString()
	if eof {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize fromContractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize toChainID error")
	}
	fee, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize fee error")
	}
	method, eof := source.NextString()
	if eof {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize method error")
	}
	args, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize args error")
	}

	this.FromChainID = fromChainID
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.Fee = fee
	this.Method = method
	this.Args = args
	return nil
}
