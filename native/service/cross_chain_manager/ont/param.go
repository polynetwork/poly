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
	"github.com/ontio/multi-chain/native/service/utils"
)

type CreateCrossChainTxParam struct {
	ToChainID uint64
	Fee       uint64
	ToAddress string
	Amount    uint64
}

func (this *CreateCrossChainTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ToChainID)
	sink.WriteUint64(this.Fee)
	sink.WriteVarBytes([]byte(this.ToAddress))
	sink.WriteUint64(this.Amount)
}

func (this *CreateCrossChainTxParam) Deserialization(source *common.ZeroCopySource) error {
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CreateCrossChainTxParam deserialize toChainID error")
	}
	fee, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CreateCrossChainTxParam deserialize fee error")
	}
	toAddress, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize address error")
	}
	amount, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CreateCrossChainTxParam deserialize amount error")
	}

	this.ToChainID = toChainID
	this.Fee = fee
	this.ToAddress = toAddress
	this.Amount = amount
	return nil
}

type ProcessCrossChainTxParam struct {
	Address     common.Address
	FromChainID uint64
	Height      uint32
	Proof       string
}

func (this *ProcessCrossChainTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.Address)
	sink.WriteUint64(this.FromChainID)
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes([]byte(this.Proof))
}

func (this *ProcessCrossChainTxParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize address error:%s", err)
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize fromChainID error")
	}
	height, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("ProcessCrossChainTxParam deserialize height error")
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

type OngUnlockParam struct {
	FromChainID uint64
	Address     common.Address
	Amount      uint64
}

func (this *OngUnlockParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.FromChainID)
	sink.WriteAddress(this.Address)
	sink.WriteUint64(this.Amount)
}

func (this *OngUnlockParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("OngLockParam deserialize fromChainID error")
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize address error:%s", err)
	}
	amount, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("OngLockParam deserialize amount error")
	}
	this.FromChainID = fromChainID
	this.Address = address
	this.Amount = amount
	return nil
}
