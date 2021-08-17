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
package neo3

import (
	"fmt"
	"github.com/joeqian10/neo3-gogogo/helper"
	"github.com/joeqian10/neo3-gogogo/mpt"
	"github.com/polynetwork/poly/common"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/header_sync/neo3"
)

func verifyFromNeoTx(proof []byte, crossChainMsg *neo3.NeoCrossChainMsg, contractId int) (*scom.MakeTxParam, error) {
	crossStateProofRoot, err := helper.UInt256FromString(crossChainMsg.StateRoot.RootHash)
	if err != nil {
		return nil, fmt.Errorf("verifyFromNeoTx, decode cross state proof root from string error: %s", err)
	}
	value, err := VerifyNeoCrossChainProof(proof, crossStateProofRoot.ToByteArray(), contractId)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromNeoTx, Verify Neo cross chain proof error: %v", err)
	}
	source := common.NewZeroCopySource(value)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(source); err != nil {
		return nil, fmt.Errorf("VerifyFromNeoTx, deserialize merkleValue error: %s", err)
	}
	return txParam, nil
}

func VerifyNeoCrossChainProof(proof []byte, stateRoot []byte, contractId int) ([]byte, error) {
	id, key, proofs, err := mpt.ResolveProof(proof)
	if err != nil {
		return nil, fmt.Errorf("VerifyNeoCrossChainProof, neo3-gogogo mpt.ResolveProof error: %v", err)
	}
	if id != contractId {
		return nil, fmt.Errorf("VerifyNeoCrossChainProof, error: id is not CCMC contract id, expected: %d, but got: %d", contractId, id)
	}
	root := helper.UInt256FromBytes(stateRoot)
	value, err := mpt.VerifyProof(root, contractId, key, proofs)
	if err != nil {
		return nil, fmt.Errorf("VerifyNeoCrossChainProof, neo3-gogogo mpt.VerifyProof error: %v", err)
	}
	return value, nil
}
