/*
 * Copyright (C) 2021 The poly network Authors
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

package ont

import (
	"fmt"
	otypes "github.com/ontio/ontology/core/types"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/merkle"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
)

func VerifyFromOntTx(proof []byte, crossChainMsg *otypes.CrossChainMsg) (*scom.MakeTxParam, error) {
	v, err := merkle.MerkleProve(proof, crossChainMsg.StatesRoot.ToArray())
	if err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, merkle.MerkleProve verify merkle proof error")
	}

	s := common.NewZeroCopySource(v)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyFromOntTx, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}
