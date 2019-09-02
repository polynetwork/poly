package btc

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
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
