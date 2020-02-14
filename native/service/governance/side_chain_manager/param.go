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

package side_chain_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
)

type RegisterSideChainParam struct {
	Address      string
	ChainId      uint64
	Router       uint64
	Name         string
	BlocksToWait uint64
}

func (this *RegisterSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteUint64(this.ChainId)
	sink.WriteUint64(this.Router)
	sink.WriteVarBytes([]byte(this.Name))
	sink.WriteUint64(this.BlocksToWait)
	return nil
}

func (this *RegisterSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize address error")
	}
	chainId, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	router, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize router error")
	}
	name, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize name error")
	}
	blocksToWait, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize blocksToWait error")
	}
	this.Address = address
	this.ChainId = chainId
	this.Router = router
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type ChainidParam struct {
	Chainid uint64
}

func (this *ChainidParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.Chainid)
}

func (this *ChainidParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	this.Chainid = chainid
	return nil
}

type RegisterRedeemParam struct {
	RedeemChainID   uint64
	ContractChainID uint64
	Redeem       []byte
	ContractAddress []byte
	Address         string
	Signs           [][]byte
}

func (this *RegisterRedeemParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.RedeemChainID)
	sink.WriteUint64(this.ContractChainID)
	sink.WriteVarBytes(this.Redeem)
	sink.WriteVarBytes([]byte(this.ContractAddress))
	sink.WriteString(this.Address)
	sink.WriteUint64(uint64(len(this.Signs)))
	for _, v := range this.Signs {
		sink.WriteVarBytes(v)
	}
}

func (this *RegisterRedeemParam) Deserialization(source *common.ZeroCopySource) error {
	redeemChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize redeemChainID error")
	}
	contractChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize contractChainID error")
	}
	redeem, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize redeemKey error")
	}
	contractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize contractAddress error")
	}
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize signs length error")
	}
	signs := make([][]byte, 0)
	for i := 0; uint64(i) < n; i++ {
		v, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("deserialize Signs error")
		}
		signs = append(signs, v)
	}

	this.RedeemChainID = redeemChainID
	this.ContractChainID = contractChainID
	this.Redeem = redeem
	this.ContractAddress = contractAddress
	this.Address = address
	this.Signs = signs
	return nil
}
