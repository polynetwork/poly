package inf

import (
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
	"math/big"
)

type ChainHandler interface {
	Verify(service *native.NativeService) (*MakeTxParam, error)
	MakeTransaction(service *native.NativeService, param *MakeTxParam) error
}

type EntranceParam struct {
	SourceChainID  uint64	`json:"sourceChainId"`
	TxData         string	`json:"txData"`
	Height         uint32	`json:"height"`
	Proof          string	`json:"proof"`
	RelayerAddress string	`json:"relayerAddress"`
	TargetChainID  uint64	`json:"targetChainId"`
}

func (this *EntranceParam) Deserialization(source *common.ZeroCopySource) error {
	sourcechainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	txdata, err := utils.DecodeString(source)
	if err != nil {
		return err
	}
	height, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	proof, err := utils.DecodeString(source)
	if err != nil {
		return err
	}
	relayerAddr, err := utils.DecodeString(source)
	if err != nil {
		return err
	}
	targetchainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}

	this.SourceChainID = sourcechainid
	this.TxData = txdata
	this.Height = uint32(height)
	this.Proof = proof
	this.RelayerAddress = relayerAddr
	this.TargetChainID = targetchainid

	return nil
}

func (this *EntranceParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.SourceChainID)
	utils.EncodeString(sink, this.TxData)
	utils.EncodeVarUint(sink, uint64(this.Height))
	utils.EncodeString(sink, this.Proof)
	utils.EncodeString(sink, this.RelayerAddress)
	utils.EncodeVarUint(sink, this.TargetChainID)
}

type MakeTxParam struct {
	FromChainID         uint64
	FromContractAddress string
	ToChainID           uint64
	Address             string
	Amount              *big.Int
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FromChainID)
	utils.EncodeString(sink, this.FromContractAddress)
	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeString(sink, this.Address)
	utils.EncodeVarBytes(sink, this.Amount.Bytes())
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	fromContractAddress, err := utils.DecodeString(source)
	if err != nil {
		return err
	}
	toChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	address, err := utils.DecodeString(source)
	if err != nil {
		return err
	}
	amount, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}

	this.FromChainID = fromChainID
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.Address = address
	this.Amount = new(big.Int).SetBytes(amount)
	return nil
}
