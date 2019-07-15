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

package cross_chain

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type CreateCrossChainTxParam struct {
	Fee     uint64
	Address common.Address

	ToChainID       uint64
	ContractAddress common.Address
	FunctionName    string
	Args            []byte
}

func (this *CreateCrossChainTxParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Fee)
	utils.EncodeAddress(sink, this.Address)

	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeAddress(sink, this.ContractAddress)
	utils.EncodeString(sink, this.FunctionName)
	utils.EncodeVarBytes(sink, this.Args)
}

func (this *CreateCrossChainTxParam) Deserialization(source *common.ZeroCopySource) error {
	fee, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize fee error:%s", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize address error:%s", err)
	}

	toChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize toChainID error:%s", err)
	}
	contractAddress, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize contractAddress error:%s", err)
	}
	functionName, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize functionName error:%s", err)
	}
	args, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize args error:%s", err)
	}
	this.Fee = fee
	this.Address = address

	this.ToChainID = toChainID
	this.ContractAddress = contractAddress
	this.FunctionName = functionName
	this.Args = args
	return nil
}

type ProcessCrossChainTxParam struct {
	Address     common.Address
	FromChainID uint64

	Height uint32
	Proof  string
}

func (this *ProcessCrossChainTxParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarUint(sink, this.FromChainID)
	utils.EncodeVarUint(sink, uint64(this.Height))
	utils.EncodeString(sink, this.Proof)
}

func (this *ProcessCrossChainTxParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize address error:%s", err)
	}
	fromChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize fromChainID error:%s", err)
	}
	height, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize height error:%s", err)
	}
	proof, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize proof error:%s", err)
	}
	this.Address = address
	this.FromChainID = fromChainID
	this.Height = uint32(height)
	this.Proof = proof
	return nil
}
