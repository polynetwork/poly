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

package chain_manager

import (
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegisterSideChainParam struct {
	Ratio              uint32
	Deposit            uint64
	OngPool            uint64
	GenesisBlockHeader []byte
	Caller             []byte
	KeyNo              uint32
}

func (this *RegisterSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, uint64(this.Ratio))
	utils.EncodeVarUint(sink, this.Deposit)
	utils.EncodeVarUint(sink, this.OngPool)
	utils.EncodeVarBytes(sink, this.Caller)
	utils.EncodeVarBytes(sink, this.GenesisBlockHeader)
	utils.EncodeVarUint(sink, uint64(this.KeyNo))
	return nil
}

func (this *RegisterSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	ratio, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize ratio error: %v", err)
	}
	if ratio > math.MaxUint32 {
		return fmt.Errorf("ratio larger than max of uint32")
	}
	deposit, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize deposit error: %v", err)
	}
	ongPool, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize ongPool error: %v", err)
	}
	genesisBlockHeader, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisBlockHeader error: %v", err)
	}
	caller, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize caller error: %v", err)
	}
	keyNo, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize keyNo error: %v", err)
	}
	if keyNo > math.MaxUint32 {
		return fmt.Errorf("initPos larger than max of uint32")
	}
	this.Ratio = uint32(ratio)
	this.Deposit = deposit
	this.OngPool = ongPool
	this.GenesisBlockHeader = genesisBlockHeader
	this.Caller = caller
	this.KeyNo = uint32(keyNo)
	return nil
}

type RegisterMainChainParam struct {
	GenesisHeader []byte
}

func (this *RegisterMainChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarBytes(sink, this.GenesisHeader)
	return nil
}

func (this *RegisterMainChainParam) Deserialization(source *common.ZeroCopySource) error {
	genesisHeader, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisHeader count error:%s", err)
	}
	this.GenesisHeader = genesisHeader
	return nil
}

type ChainIDParam struct {
	ChainID uint64
}

func (this *ChainIDParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	return nil
}

func (this *ChainIDParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	this.ChainID = chainID
	return nil
}

type QuitSideChainParam struct {
	ChainID uint64
}

func (this *QuitSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	return nil
}

func (this *QuitSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	this.ChainID = chainID
	return nil
}

type InflationParam struct {
	ChainID    uint64
	DepositAdd uint64
	OngPoolAdd uint64
}

func (this *InflationParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	utils.EncodeVarUint(sink, this.DepositAdd)
	utils.EncodeVarUint(sink, this.OngPoolAdd)
	return nil
}

func (this *InflationParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	depositAdd, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize depositAdd error: %v", err)
	}
	ongPoolAdd, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize ongPoolAdd error: %v", err)
	}
	this.ChainID = chainID
	this.DepositAdd = depositAdd
	this.OngPoolAdd = ongPoolAdd
	return nil
}

type NodeToSideChainParams struct {
	PeerPubkey string
	Address    common.Address
	ChainID    uint64
}

func (this *NodeToSideChainParams) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeString(sink, this.PeerPubkey)
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarUint(sink, this.ChainID)
	return nil
}

func (this *NodeToSideChainParams) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error: %v", err)
	}
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.ChainID = chainID
	return nil
}

type BlackSideChainParam struct {
	ChainID uint64
	Address common.Address
}

func (this *BlackSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	utils.EncodeAddress(sink, this.Address)
	return nil
}

func (this *BlackSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error: %v", err)
	}
	this.ChainID = chainID
	this.Address = address
	return nil
}

type StakeSideChainParam struct {
	ChainID uint64
	Pubkey  string
	Amount  uint64
}

func (this *StakeSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	utils.EncodeString(sink, this.Pubkey)
	utils.EncodeVarUint(sink, this.Amount)
	return nil
}

func (this *StakeSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	pubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize pubkey error: %v", err)
	}
	amount, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize amount error: %v", err)
	}
	this.ChainID = chainID
	this.Pubkey = pubkey
	this.Amount = amount
	return nil
}

type GovernanceEpoch struct {
	ChainID uint64
	Epoch   uint32
}

func (this *GovernanceEpoch) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	if this.Epoch > math.MaxUint32 {
		return fmt.Errorf("serialize GovernanceEpoch error: Epoch more than MaxUint32")
	}
	utils.EncodeVarUint(sink, uint64(this.Epoch))
	return nil
}

func (this *GovernanceEpoch) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	epoch, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize epoch error: %v", err)
	}
	if epoch > math.MaxUint32 {
		return fmt.Errorf("deserialize GovernanceEpoch error: Epoch more than MaxUint32")
	}
	this.ChainID = chainID
	this.Epoch = uint32(epoch)
	return nil
}
