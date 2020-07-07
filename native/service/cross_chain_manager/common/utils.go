/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */
package common

import (
	"fmt"
	"math/big"
	"strings"

	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
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
