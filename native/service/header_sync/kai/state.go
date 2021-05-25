package kai

import (
	"fmt"

	"github.com/kardiachain/go-kardia/lib/bytes"
	"github.com/polynetwork/poly/common"
)

type EpochSwitchInfo struct {
	// The height where validators set changed last time. Poly only accept
	// header and proof signed by new validators. That means the header
	// can not be lower than this height.
	Height int64

	// Hash of the block at `Height`. Poly don't save the whole header.
	// So we can identify the content of this block by `BlockHash`.
	BlockHash bytes.HexBytes

	// The hash of new validators set which used to verify validators set
	// committed with proof.
	NextValidatorsHash bytes.HexBytes

	ChainID string
}

func (info *EpochSwitchInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteInt64(info.Height)
	sink.WriteVarBytes(info.BlockHash)
	sink.WriteVarBytes(info.NextValidatorsHash)
	sink.WriteString(info.ChainID)
}

func (info *EpochSwitchInfo) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	info.Height, eof = source.NextInt64()
	if eof {
		return fmt.Errorf("deserialize height of EpochSwitchInfo failed")
	}
	info.BlockHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("deserialize BlockHash of EpochSwitchInfo failed")
	}
	info.NextValidatorsHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("deserialize NextValidatorsHash of EpochSwitchInfo failed")
	}
	info.ChainID, eof = source.NextString()
	if eof {
		return fmt.Errorf("deserialize ChainID of EpochSwitchInfo failed")
	}
	return nil
}
