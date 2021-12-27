package pixiechain

import (
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/polynetwork/poly/native/service/header_sync/eth"
)

// Handler ...
type Handler struct {
}

// GenesisHeader ...
type GenesisHeader struct {
	Header         eth.Header
	PrevValidators []HeightAndValidators
}

// ExtraInfo ...
type ExtraInfo struct {
	ChainID *big.Int // chainId of pixie chain. mainnet: 6626, testnet: 666
	Period  uint64
}

// Context ...
type Context struct {
	ExtraInfo ExtraInfo
	ChainID   uint64
}

// HeaderWithChainID ...
type HeaderWithChainID struct {
	Header  *HeaderWithDifficultySum
	ChainID uint64
}

// HeaderWithDifficultySum ...
type HeaderWithDifficultySum struct {
	Header          *eth.Header   `json:"header"`
	DifficultySum   *big.Int      `json:"difficultySum"`
	EpochParentHash *ecommon.Hash `json:"epochParentHash"`
}
