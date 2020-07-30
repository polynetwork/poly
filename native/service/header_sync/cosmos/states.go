package cosmos

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"
)

type CosmosEpochSwitchInfo struct {
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

	// The cosmos chain-id of this chain basing Cosmos-sdk.
	ChainID string
}

func (info *CosmosEpochSwitchInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteInt64(info.Height)
	sink.WriteVarBytes(info.BlockHash)
	sink.WriteVarBytes(info.NextValidatorsHash)
	sink.WriteString(info.ChainID)
}

func (info *CosmosEpochSwitchInfo) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	info.Height, eof = source.NextInt64()
	if eof {
		return fmt.Errorf("deserialize height of CosmosEpochSwitchInfo failed")
	}
	info.BlockHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("deserialize BlockHash of CosmosEpochSwitchInfo failed")
	}
	info.NextValidatorsHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("deserialize NextValidatorsHash of CosmosEpochSwitchInfo failed")
	}
	info.ChainID, eof = source.NextString()
	if eof {
		return fmt.Errorf("deserialize ChainID of CosmosEpochSwitchInfo failed")
	}
	return nil
}

type CosmosHeader struct {
	Header  types.Header
	Commit  *types.Commit
	Valsets []*types.Validator
}
