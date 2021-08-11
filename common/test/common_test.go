/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */
package test

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHexAndBytesTransfer(t *testing.T) {
	testBytes := []byte("10, 11, 12, 13, 14, 15, 16, 17, 18, 19")
	stringAfterTrans := common.ToHexString(testBytes)
	bytesAfterTrans, err := common.HexToBytes(stringAfterTrans)
	assert.Nil(t, err)
	assert.Equal(t, testBytes, bytesAfterTrans)
}

func TestGetNonce(t *testing.T) {
	nonce1 := common.GetNonce()
	nonce2 := common.GetNonce()
	assert.NotEqual(t, nonce1, nonce2)
}

func TestFileExisted(t *testing.T) {
	assert.True(t, common.FileExisted("common_test.go"))
	assert.True(t, common.FileExisted("common.go"))
	assert.False(t, common.FileExisted("../log/log.og"))
	assert.False(t, common.FileExisted("../log/log.go"))
	assert.True(t, common.FileExisted("./log/log.go"))
}

func TestBase58(t *testing.T) {
	addr := common.ADDRESS_EMPTY
	fmt.Println("emtpy addr:", addr.ToBase58())
}
