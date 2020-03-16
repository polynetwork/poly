package cross_chain_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
)

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
