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
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type FromMerkleValue struct {
	RequestID                uint64
	CreateCrossChainTxMerkle *CreateCrossChainTxMerkle
}

func (this *FromMerkleValue) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.RequestID)
	this.CreateCrossChainTxMerkle.Serialization(sink)
}

func (this *FromMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	requestID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize requestID error:%s", err)
	}
	createCrossChainTxMerkle := new(CreateCrossChainTxMerkle)
	err = createCrossChainTxMerkle.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize createCrossChainTxMerkle error:%s", err)
	}

	this.RequestID = requestID
	this.CreateCrossChainTxMerkle = createCrossChainTxMerkle
	return nil
}

type ToMerkleValue struct {
	RequestID   uint64
	MakeTxParam *inf.MakeTxParam
}

func (this *ToMerkleValue) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.RequestID)
	this.MakeTxParam.Serialization(sink)
}

func (this *ToMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	requestID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize requestID error:%s", err)
	}
	makeTxParam := new(inf.MakeTxParam)
	err = makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize makeTxParam error:%s", err)
	}

	this.RequestID = requestID
	this.MakeTxParam = makeTxParam
	return nil
}

type CreateCrossChainTxMerkle struct {
	FromChainID         uint64
	FromContractAddress string
	ToChainID           uint64
	Fee                 uint64
	ToAddress           string
	Amount              uint64
}

func (this *CreateCrossChainTxMerkle) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FromChainID)
	utils.EncodeString(sink, this.FromContractAddress)
	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeVarUint(sink, this.Fee)
	utils.EncodeString(sink, this.ToAddress)
	utils.EncodeVarUint(sink, this.Amount)
}

func (this *CreateCrossChainTxMerkle) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize fromChainID error:%s", err)
	}
	fromContractAddress, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize fromContractAddress error:%s", err)
	}
	toChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize toChainID error:%s", err)
	}
	fee, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize fee error:%s", err)
	}
	toAddress, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize toAddress error:%s", err)
	}
	amount, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxMerkle deserialize amount error:%s", err)
	}

	this.FromChainID = fromChainID
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.Fee = fee
	this.ToAddress = toAddress
	this.Amount = amount
	return nil
}
