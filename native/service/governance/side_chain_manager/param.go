/*
 * Copyright (C) 2020 The poly network Authors
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
package side_chain_manager

import (
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/core/ledger"
)

type RegisterSideChainParam struct {
	Address      common.Address
	ChainId      uint64
	Router       uint64
	Name         string
	BlocksToWait uint64
	CCMCAddress  []byte
	ExtraInfo    []byte
}

func (this *RegisterSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(this.Address[:])
	sink.WriteVarUint(this.ChainId)
	sink.WriteVarUint(this.Router)
	sink.WriteVarBytes([]byte(this.Name))
	sink.WriteVarUint(this.BlocksToWait)
	sink.WriteVarBytes(this.CCMCAddress)

	height := config.GetExtraInfoHeight(config.DefConfig.P2PNode.NetworkId)
	if ledger.DefLedger.GetCurrentBlockHeight() >= height {
		sink.WriteVarBytes(this.ExtraInfo)
	}

	return nil
}

func (this *RegisterSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("utils.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	chainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize chainid error")
	}
	router, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize router error")
	}
	name, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize name error")
	}
	blocksToWait, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize blocksToWait error")
	}
	if blocksToWait == 0 {
		return fmt.Errorf("minimal value of BlocksToWait is 1")
	}
	CCMCAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize CCMCAddress error")
	}
	ExtraInfo, _ := source.NextVarBytes()
	this.Address = addr
	this.ChainId = chainId
	this.Router = router
	this.Name = name
	this.BlocksToWait = blocksToWait
	this.CCMCAddress = CCMCAddress
	this.ExtraInfo = ExtraInfo
	return nil
}

type ChainidParam struct {
	Chainid uint64
	Address common.Address
}

func (this *ChainidParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.Chainid)
	sink.WriteVarBytes(this.Address[:])
}

func (this *ChainidParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize chainid error")
	}

	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	this.Chainid = chainid
	this.Address = addr
	return nil
}

type RegisterRedeemParam struct {
	RedeemChainID   uint64
	ContractChainID uint64
	Redeem          []byte
	CVersion        uint64
	ContractAddress []byte
	Signs           [][]byte
}

func (this *RegisterRedeemParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.RedeemChainID)
	sink.WriteVarUint(this.ContractChainID)
	sink.WriteVarBytes(this.Redeem)
	sink.WriteVarUint(this.CVersion)
	sink.WriteVarBytes(this.ContractAddress)
	sink.WriteVarUint(uint64(len(this.Signs)))
	for _, v := range this.Signs {
		sink.WriteVarBytes(v)
	}
}

func (this *RegisterRedeemParam) Deserialization(source *common.ZeroCopySource) error {
	redeemChainID, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize redeemChainID error")
	}
	contractChainID, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize contractChainID error")
	}
	redeem, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize redeemKey error")
	}
	cver, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize contract version error")
	}
	contractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("RegisterRedeemParam deserialize contractAddress error")
	}
	n, eof := source.NextVarUint()
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
	this.CVersion = cver
	this.ContractAddress = contractAddress
	this.Signs = signs
	return nil
}

type BtcTxParamDetial struct {
	PVersion  uint64
	FeeRate   uint64
	MinChange uint64
}

func (this *BtcTxParamDetial) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.PVersion)
	sink.WriteVarUint(this.FeeRate)
	sink.WriteVarUint(this.MinChange)
}

func (this *BtcTxParamDetial) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.PVersion, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("BtcTxParamDetial deserialize version error")
	}
	this.FeeRate, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("BtcTxParamDetial deserialize fee rate error")
	}
	this.MinChange, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("BtcTxParamDetial deserialize min-change error")
	}
	return nil
}

type BtcTxParam struct {
	Redeem        []byte
	RedeemChainId uint64
	Sigs          [][]byte
	Detial        *BtcTxParamDetial
}

func (this *BtcTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Redeem)
	sink.WriteVarUint(this.RedeemChainId)
	sink.WriteVarUint(uint64(len(this.Sigs)))
	for _, v := range this.Sigs {
		sink.WriteVarBytes(v)
	}
	this.Detial.Serialization(sink)
}

func (this *BtcTxParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Redeem, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcFeeRateParam deserialize redeem error")
	}
	this.RedeemChainId, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("BtcFeeRateParam deserialize redeem chain-id error")
	}
	l, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("BtcFeeRateParam deserialize length of signature array error")
	}
	sigs := make([][]byte, l)
	for i := uint64(0); i < l; i++ {
		sigs[i], eof = source.NextVarBytes()
		if eof {
			return fmt.Errorf("BtcFeeRateParam deserialize no.%d signature error", i+1)
		}
	}
	this.Sigs = sigs
	detial := &BtcTxParamDetial{}
	if err := detial.Deserialization(source); err != nil {
		return fmt.Errorf("BtcFeeRateParam deserialize detail error: %v", err)
	}
	this.Detial = detial
	return nil
}
