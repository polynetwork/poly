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
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type MerkleValue struct {
	RequestID               uint64
	CreateCrossChainTxParam *CreateCrossChainTxParam
}

func (this *MerkleValue) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.RequestID)
	this.CreateCrossChainTxParam.Serialization(sink)
}

func (this *MerkleValue) Deserialization(source *common.ZeroCopySource) error {
	requestID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize requestID error:%s", err)
	}
	createCrossChainTxParam := new(CreateCrossChainTxParam)
	err = createCrossChainTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize createCrossChainTxParam error:%s", err)
	}

	this.RequestID = requestID
	this.CreateCrossChainTxParam = createCrossChainTxParam
	return nil
}
