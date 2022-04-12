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

package ripple

import (
	"fmt"
	"github.com/polynetwork/poly/common"
)

type MultiSignParam struct {
	ToChainId    uint64
	AssetAddress []byte
	FromChainId  uint64
	TxHash       []byte
	TxJson       string
}

func (this *MultiSignParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.ToChainId)
	sink.WriteVarBytes(this.AssetAddress)
	sink.WriteVarUint(this.FromChainId)
	sink.WriteVarBytes(this.TxHash)
	sink.WriteString(this.TxJson)
}

func (this *MultiSignParam) Deserialization(source *common.ZeroCopySource) error {
	toChainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize toChainId error")
	}
	assetAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize assetAddress error")
	}
	fromChainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize fromChainId error")
	}
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize tx hash error")
	}
	txJson, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize txJson error")
	}

	this.ToChainId = toChainId
	this.AssetAddress = assetAddress
	this.FromChainId = fromChainId
	this.TxHash = txHash
	this.TxJson = txJson
	return nil
}

type ReconstructTxParam struct {
	FromChainId  uint64
	TxHash       []byte
	ToChainId    uint64
}

func (this *ReconstructTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.FromChainId)
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarUint(this.ToChainId)
}

func (this *ReconstructTxParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("ReconstructTxParam deserialize fromChainId error")
	}
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("ReconstructTxParam deserialize tx hash error")
	}
	toChainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("ReconstructTxParam deserialize toChainId error")
	}

	this.FromChainId = fromChainId
	this.TxHash = txHash
	this.ToChainId = toChainId
	return nil
}
