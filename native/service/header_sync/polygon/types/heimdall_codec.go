package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/polynetwork/poly/native/service/header_sync/polygon/types/secp256k1"
	"github.com/tendermint/tendermint/crypto"
)

// NewCDC ...
func NewCDC() *codec.Codec {
	cdc := codec.New()

	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoName, nil)

	return cdc
}

var cdc = NewCDC()
