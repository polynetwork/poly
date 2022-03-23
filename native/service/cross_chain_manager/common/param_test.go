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
package common

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestMakeTxParamWithSender(t *testing.T) {
	txParam := MakeTxParam{TxHash: []byte("hash"),
		CrossChainID:        []byte("1"),
		FromContractAddress: []byte("from addr"),
		ToChainID:           1,
		ToContractAddress:   []byte("to addr"),
		Method:              "test",
		Args:                []byte("args")}

	value := MakeTxParamWithSender{Sender: ethcommon.HexToAddress("abc"), MakeTxParam: txParam}
	data, err := value.Serialization()

	assert.Nil(t, err)

	var decoded MakeTxParamWithSender
	err = decoded.Deserialization(data)
	assert.Nil(t, err)

	assert.Equal(t, value, decoded)
}
