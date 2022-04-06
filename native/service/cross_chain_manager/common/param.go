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
package common

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
)

const (
	IMPORT_OUTER_TRANSFER_NAME = "ImportOuterTransfer"
	MULTI_SIGN                 = "MultiSign"
	BLACK_CHAIN                = "BlackChain"
	WHITE_CHAIN                = "WhiteChain"

	BLACKED_CHAIN = "BlackedChain"
)

var (
	KEY_PREFIX_BTC = "btc"

	KEY_PREFIX_BTC_VOTE = "btcVote"
	REQUEST             = "request"
	DONE_TX             = "doneTx"

	NOTIFY_MAKE_PROOF = "makeProof"
)

type ChainHandler interface {
	MakeDepositProposal(service *native.NativeService) (*MakeTxParam, error)
}

type InitRedeemScriptParam struct {
	RedeemScript string
}

func (this *InitRedeemScriptParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.RedeemScript)
}

func (this *InitRedeemScriptParam) Deserialization(source *common.ZeroCopySource) error {
	redeemScript, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize redeemScript error")
	}

	this.RedeemScript = redeemScript
	return nil
}

type EntranceParam struct {
	SourceChainID         uint64 `json:"sourceChainId"`
	Height                uint32 `json:"height"`
	Proof                 []byte `json:"proof"`
	RelayerAddress        []byte `json:"relayerAddress"`
	Extra                 []byte `json:"extra"`
	HeaderOrCrossChainMsg []byte `json:"headerOrCrossChainMsg"`
}

func (this *EntranceParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.SourceChainID)
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes(this.Proof)
	sink.WriteVarBytes(this.RelayerAddress)
	sink.WriteVarBytes(this.Extra)
	sink.WriteVarBytes(this.HeaderOrCrossChainMsg)
}

func (this *EntranceParam) Deserialization(source *common.ZeroCopySource) error {
	sourceChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("EntranceParam deserialize sourcechainid error")
	}

	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("EntranceParam deserialize height error")
	}
	proof, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize proof error")
	}
	relayerAddr, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize relayerAddr error")
	}
	extra, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize txdata error")
	}
	headerOrCrossChainMsg, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize headerOrCrossChainMsg error")
	}
	this.SourceChainID = sourceChainID
	this.Height = height
	this.Proof = proof
	this.RelayerAddress = relayerAddr
	this.Extra = extra
	this.HeaderOrCrossChainMsg = headerOrCrossChainMsg
	return nil
}

type MakeTxParamWithSender struct {
	Sender ethcommon.Address
	MakeTxParam
}

// this method is only used in test
func (this *MakeTxParamWithSender) Serialization() (data []byte, err error) {
	sink := common.NewZeroCopySink(nil)
	sink.WriteAddress(common.Address(this.Sender))
	this.MakeTxParam.Serialization(sink)
	data = sink.Bytes()
	return
}

func (this *MakeTxParamWithSender) Deserialization(data []byte) (err error) {

	source := common.NewZeroCopySource(data)

	addr, eof := source.NextAddress()
	if eof {
		err = fmt.Errorf("MakeTxParamWithSender NextAddress fail")
		return
	}
	this.Sender = ethcommon.Address(addr)
	return this.MakeTxParam.Deserialization(source)
}

type MakeTxParam struct {
	TxHash              []byte
	CrossChainID        []byte
	FromContractAddress []byte
	ToChainID           uint64
	ToContractAddress   []byte
	Method              string
	Args                []byte
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarBytes(this.CrossChainID)
	sink.WriteVarBytes(this.FromContractAddress)
	sink.WriteUint64(this.ToChainID)
	sink.WriteVarBytes(this.ToContractAddress)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes(this.Args)
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize txHash error")
	}
	crossChainID, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize crossChainID error")
	}
	fromContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize fromContractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize toChainID error")
	}
	toContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize toContractAddress error")
	}
	method, eof := source.NextString()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize method error")
	}
	args, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize args error")
	}

	this.TxHash = txHash
	this.CrossChainID = crossChainID
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.ToContractAddress = toContractAddress
	this.Method = method
	this.Args = args
	return nil
}

type MultiSignParam struct {
	ChainID   uint64
	RedeemKey string
	TxHash    []byte
	Address   string
	Signs     [][]byte
}

func (this *MultiSignParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteString(this.RedeemKey)
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteUint64(uint64(len(this.Signs)))
	for _, v := range this.Signs {
		sink.WriteVarBytes(v)
	}
}

func (this *MultiSignParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize txHash error")
	}
	redeemKey, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize redeemKey error")
	}
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize txHash error")
	}
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize signs length error")
	}
	signs := make([][]byte, 0)
	for i := 0; uint64(i) < n; i++ {
		v, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("deserialize Signs error")
		}
		signs = append(signs, v)
	}

	this.ChainID = chainID
	this.RedeemKey = redeemKey
	this.TxHash = txHash
	this.Address = address
	this.Signs = signs
	return nil
}

type ToMerkleValue struct {
	TxHash      []byte
	FromChainID uint64
	MakeTxParam *MakeTxParam
}

func (this *ToMerkleValue) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	sink.WriteUint64(this.FromChainID)
	this.MakeTxParam.Serialization(sink)
}

func (this *ToMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MerkleValue deserialize txHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MerkleValue deserialize fromChainID error")
	}

	makeTxParam := new(MakeTxParam)
	err := makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize makeTxParam error:%s", err)
	}

	this.TxHash = txHash
	this.FromChainID = fromChainID
	this.MakeTxParam = makeTxParam
	return nil
}


type BlackChainParam struct {
	ChainID uint64
}

func (this *BlackChainParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.ChainID)
}

func (this *BlackChainParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("BlackChainParam deserialize chainID error")
	}

	this.ChainID = chainID
	return nil
}

