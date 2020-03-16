package common

import (
	"fmt"
	"math/big"
	"strings"

	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/utils"
)

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

func NotifyMakeProof(native *native.NativeService, fromChainID, toChainID uint64, txHash string, key string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{NOTIFY_MAKE_PROOF, fromChainID, toChainID, txHash, native.GetHeight(), key},
		})
}

func PutDoneTx(native *native.NativeService, crossChainID []byte, chainID uint64) error {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, crossChainID),
		states.GenRawStorageItem(crossChainID))
	return nil
}

func CheckDoneTx(native *native.NativeService, crossChainID []byte, chainID uint64) error {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	value, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, crossChainID))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.GetCacheDB().Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}
