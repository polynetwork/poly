package common

import (
	"fmt"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/utils"
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

func NotifyMakeProof(native *native.NativeService, txHash string, toChainID uint64, key string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{NotifyMakeProofInfo[toChainID], txHash, toChainID, native.GetHeight(), key},
		})
}

func PutDoneTx(native *native.NativeService, txHash, proof []byte, chainID uint64) error {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, txHash, proof), states.GenRawStorageItem(txHash))
	return nil
}

func CheckDoneTx(native *native.NativeService, txHash, proof []byte, chainID uint64) error {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	value, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, txHash, proof))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.GetCacheDB().Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}
