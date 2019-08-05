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

package ong

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type OngLockParam struct {
	Fee       uint64 `json:"fee"`
	ToChainID uint64 `json:"toChainId"`
	Address   string `json:"address"`
	Amount    uint64 `json:"amount"`
}

func (this *OngLockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Fee)
	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeString(sink, this.Address)
	utils.EncodeVarUint(sink, this.Amount)
}

func (this *OngLockParam) Deserialization(source *common.ZeroCopySource) error {
	fee, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize fee error:%s", err)
	}
	toChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize toChainID error:%s", err)
	}
	address, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize address error:%s", err)
	}
	amount, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize amount error:%s", err)
	}
	this.Fee = fee
	this.ToChainID = toChainID
	this.Address = address
	this.Amount = amount
	return nil
}

type OngUnlockParam struct {
	FromChainID uint64
	Address     common.Address
	Amount      uint64
}

func (this *OngUnlockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FromChainID)
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarUint(sink, this.Amount)
}

func (this *OngUnlockParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize fromChainID error:%s", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize address error:%s", err)
	}
	amount, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngLockParam deserialize amount error:%s", err)
	}
	this.FromChainID = fromChainID
	this.Address = address
	this.Amount = amount
	return nil
}
