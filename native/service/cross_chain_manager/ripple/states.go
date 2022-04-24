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
	"sort"
)

type MultisignInfo struct {
	Status bool
	SigMap map[string]bool
}

func (this *MultisignInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBool(this.Status)

	var sigList []string
	for k := range this.SigMap {
		sigList = append(sigList, k)
	}
	sort.SliceStable(sigList, func(i, j int) bool {
		return sigList[i] > sigList[j]
	})

	sink.WriteVarUint(uint64(len(this.SigMap)))
	for _, sig := range sigList {
		sink.WriteString(sig)
		sink.WriteBool(this.SigMap[sig])
	}
}

func (this *MultisignInfo) Deserialization(source *common.ZeroCopySource) error {
	status, eof := source.NextBool()
	if eof {
		return fmt.Errorf("MultisignInfo deserialize status error")
	}

	l, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("MultisignInfo deserialize length of sig map error")
	}
	sigMap := make(map[string]bool, l)
	for i := uint64(0); i < l; i++ {
		sig, eof := source.NextString()
		if eof {
			return fmt.Errorf("MultisignInfo deserialize no.%d sig error", i+1)
		}
		v, eof := source.NextBool()
		if eof {
			return fmt.Errorf("MultisignInfo deserialize no.%d bool value error", i+1)
		}
		sigMap[sig] = v
	}

	this.Status = status
	this.SigMap = sigMap
	return nil
}

type TxJsonInfo struct {
	TxJson string
	FromChainId uint64
}

func (this *TxJsonInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.TxJson)
	sink.WriteUint64(this.FromChainId)
}

func (this *TxJsonInfo) Deserialization(source *common.ZeroCopySource) error {
	txJson, eof := source.NextString()
	if eof {
		return fmt.Errorf("TxJsonInfo deserialize txjson error")
	}
	fromChainId, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("TxJsonInfo deserialize fromChainId error")
	}

	this.TxJson = txJson
	this.FromChainId = fromChainId
	return nil
}