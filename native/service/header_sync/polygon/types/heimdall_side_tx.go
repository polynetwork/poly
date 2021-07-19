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
package types

import (
	"encoding/binary"
	"fmt"

	"github.com/polynetwork/poly/native/service/header_sync/polygon/types/common"
)

// SideTxResult side tx result for vote
type SideTxResult struct {
	TxHash []byte `json:"tx_hash"`
	Result int32  `json:"result"`
	Sig    []byte `json:"sig"`
}

func (sp *SideTxResult) String() string {
	if sp == nil {
		return ""
	}

	return fmt.Sprintf("SideTxResult{%X (%v) %X}",
		common.Fingerprint(sp.TxHash),
		sp.Result,
		common.Fingerprint(sp.Sig),
	)
}

// SideTxResultWithData side tx result with data for vote
type SideTxResultWithData struct {
	SideTxResult

	Data []byte `json:"data"`
}

// GetBytes returns data bytes for sign
func (sp *SideTxResultWithData) GetBytes() []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(sp.Result))

	data := make([]byte, 0)
	data = append(data, bs[3]) // use last byte as result
	if len(sp.Data) > 0 {
		data = append(data, sp.Data...)
	}
	return data
}

func (sp *SideTxResultWithData) String() string {
	if sp == nil {
		return ""
	}

	return fmt.Sprintf("SideTxResultWithData {%s %X}",
		sp.SideTxResult.String(),
		common.Fingerprint(sp.Data),
	)
}
