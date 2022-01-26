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
	"github.com/polynetwork/poly/common"
)

type ReplenishTxParam struct {
	ChainId  uint64
	TxHashes []string
}

func (this *ReplenishTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.ChainId)
	sink.WriteVarUint(uint64(len(this.TxHashes)))
	for _, v := range this.TxHashes {
		sink.WriteString(v)
	}
}

func (this *ReplenishTxParam) Deserialization(source *common.ZeroCopySource) error {
	chainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize chain id error")
	}

	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize txhashes length error")
	}
	txHashes := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		txHash, eof := source.NextString()
		if eof {
			return fmt.Errorf("source.NextString, deserialize tx hash error")
		}
		txHashes = append(txHashes, txHash)
	}

	this.ChainId = chainId
	this.TxHashes = txHashes
	return nil
}
