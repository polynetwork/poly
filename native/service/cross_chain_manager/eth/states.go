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
package eth

import (
	"fmt"
	"math/big"

	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/polynetwork/poly/common"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
)

type ToMerkleValue struct {
	TxHash            common.Uint256
	ToContractAddress string
	MakeTxParam       *scom.MakeTxParam
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
	makeTxParam := new(scom.MakeTxParam)
	err := makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize makeTxParam error:%s", err)
	}

	this.TxHash = txHash
	this.ToContractAddress = toContractAddress
	this.MakeTxParam = makeTxParam
	return nil
}

type ProofAccount struct {
	Nounce   *big.Int
	Balance  *big.Int
	Storage  ethcomm.Hash
	Codehash ethcomm.Hash
}
