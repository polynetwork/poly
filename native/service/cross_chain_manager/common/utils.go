package common

import (
	"fmt"
	"github.com/ontio/multi-chain/core/genesis"
	"math/big"
	"strings"

	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
)

func ValidateVote(service *native.NativeService, vote *Vote) error {
	sum := 0
	for _, v := range genesis.GenesisBookkeepers {
		address := types.AddressFromPubKey(v)
		_, ok := vote.VoteMap[address.ToBase58()]
		if ok {
			sum = sum + 1
		}
	}

	if sum != (2*len(genesis.GenesisBookkeepers)+2)/3 {
		return fmt.Errorf("ValidateVote, not enough vote")
	}
	return nil
}

func Replace0x(s string) string {
	return strings.Replace(strings.ToLower(s), "0x", "", 1)
}

func ConverDecimal(fromDecimal int, toDecimal int, fromAmount *big.Int) *big.Int {
	diff := fromDecimal - toDecimal
	if diff > 0 {
		return new(big.Int).Div(fromAmount, ethmath.Exp(big.NewInt(10), big.NewInt(int64(diff))))
	} else if diff < 0 {
		return new(big.Int).Mul(fromAmount, ethmath.Exp(big.NewInt(10), big.NewInt(int64(-diff))))
	}
	return fromAmount
}
