package eth

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
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
