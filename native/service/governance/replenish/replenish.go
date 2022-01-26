/*
 * Copyright (C) 2021 The poly network Authors
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

package replenish

import (
	"fmt"
	"github.com/polynetwork/poly/native/event"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//function name
	REPLENISH_TX = "replenishTx"
)

//Register methods
func RegisterReplenishContract(native *native.NativeService) {
	native.Register(REPLENISH_TX, ReplenishTx)
}

func ReplenishTx(native *native.NativeService) ([]byte, error) {
	params := new(ReplenishTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ReplenishTx, contract params deserialize error: %v", err)
	}

	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.ReplenishContractAddress,
			States:          []interface{}{"ReplenishTx", params.TxHashes, params.ChainId},
		})
	return utils.BYTE_TRUE, nil
}
