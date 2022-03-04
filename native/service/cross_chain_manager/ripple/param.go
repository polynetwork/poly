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

type TxJson struct {
	Account         string
	Amount          string
	Destination     string
	Fee             string
	Sequence        string
	SigningPubKey   string
	TransactionType string
	hash            string
	Signers         []*Signer
}

type Signer struct {
	Account       string
	SigningPubKey string
	TxnSignature  string
}

type MultiSignParam struct {
	ChainId      uint64
	AssetAddress []byte

	Id     []byte
	TxJson string
}

func (this *MultiSignParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.ChainId)
	sink.WriteVarBytes(this.AssetAddress)
	sink.WriteVarBytes(this.Id)
	sink.WriteString(this.TxJson)
}

func (this *MultiSignParam) Deserialization(source *common.ZeroCopySource) error {
	chainId, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize chainId error")
	}
	assetAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize assetAddress error")
	}
	id, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize Id error")
	}
	txJson, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize txJson error")
	}

	this.ChainId = chainId
	this.AssetAddress = assetAddress
	this.Id = id
	this.TxJson = txJson
	return nil
}
