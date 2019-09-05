package btc

import (
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

type BtcProof struct {
	Tx           []byte
	Proof        []byte
	Height       uint32
	BlocksToWait uint64
}

func (this *BtcProof) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.Tx)
	utils.EncodeVarBytes(sink, this.Proof)
	utils.EncodeVarUint(sink, uint64(this.Height))
	utils.EncodeVarUint(sink, this.BlocksToWait)
}

func (this *BtcProof) Deserialization(source *common.ZeroCopySource) error {
	tx, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("BtcProof deserialize tx error:%s", err)
	}
	proof, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("BtcProof deserialize proof error:%s", err)
	}
	height, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("BtcProof deserialize height error:%s", err)
	}
	blocksToWait, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("BtcProof deserialize blocksToWait error:%s", err)
	}

	this.Tx = tx
	this.Proof = proof
	this.Height = uint32(height)
	this.BlocksToWait = blocksToWait
	return nil
}

type Utxos struct {
	Utxos []*Utxo
}

func (this *Utxos) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(len(this.Utxos)))
	for _, v := range this.Utxos {
		v.Serialization(sink)
	}
}

func (this *Utxos) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize Utxos length error: %v", err)
	}
	utxos := make([]*Utxo, 0)
	for i := 0; uint64(i) < n; i++ {
		utxo := new(Utxo)
		if err := utxo.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize utxo error: %v", err)
		}
		utxos = append(utxos, utxo)
	}

	this.Utxos = utxos
	return nil
}

type Utxo struct {
	// Previous txid and output index
	Op *OutPoint

	// Block height where this tx was confirmed, 0 for unconfirmed
	AtHeight uint32

	// The higher the better
	Value uint64

	// Output script
	ScriptPubkey []byte
}

func (this *Utxo) Serialization(sink *common.ZeroCopySink) {
	this.Op.Serialization(sink)
	utils.EncodeVarUint(sink, uint64(this.AtHeight))
	utils.EncodeVarUint(sink, this.Value)
	utils.EncodeVarBytes(sink, this.ScriptPubkey)
}

func (this *Utxo) Deserialization(source *common.ZeroCopySource) error {
	op := new(OutPoint)
	err := op.Deserialization(source)
	if err != nil {
		return fmt.Errorf("Utxo deserialize OutPoint error:%s", err)
	}
	atHeight, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OutPoint deserialize atHeight error:%s", err)
	}
	value, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OutPoint deserialize value error:%s", err)
	}
	scriptPubkey, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("OutPoint deserialize scriptPubkey error:%s", err)
	}

	this.Op = op
	this.AtHeight = uint32(atHeight)
	this.Value = value
	this.ScriptPubkey = scriptPubkey
	return nil
}

type OutPoint struct {
	Hash  []byte
	Index uint32
}

func (this *OutPoint) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.Hash)
	utils.EncodeVarUint(sink, uint64(this.Index))
}

func (this *OutPoint) Deserialization(source *common.ZeroCopySource) error {
	hash, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("OutPoint deserialize hash error:%s", err)
	}
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OutPoint deserialize height error:%s", err)
	}

	this.Hash = hash
	this.Index = uint32(index)
	return nil
}
