/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package types

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/polynetwork/poly/common"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
)

func TestHeader_Serialize(t *testing.T) {
	header := Header{}
	header.Height = 321
	header.Bookkeepers = make([]keypair.PublicKey, 0)
	header.SigData = make([][]byte, 0)
	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)
	bs := sink.Bytes()

	var h2 Header
	source := common.NewZeroCopySource(bs)
	err := h2.Deserialization(source)
	assert.Equal(t, fmt.Sprint(header), fmt.Sprint(h2))

	assert.Nil(t, err)
}

func TestHeader(t *testing.T) {
	h := Header{
		Version:          0,
		ChainID:          123,
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStateRoot:   common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        123,
		Height:           123,
		ConsensusData:    123,
		ConsensusPayload: []byte{123},
		NextBookkeeper:   common.ADDRESS_EMPTY,
	}
	sink := common.NewZeroCopySink(nil)
	err := h.Serialization(sink)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	err = h.Serialize(buf)

	assert.NoError(t, err)
	assert.Equal(t, sink.Bytes(), buf.Bytes())

	var header1 Header
	err = header1.Deserialize(buf)
	assert.NoError(t, err)

	var header2 Header

	err = header2.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	assert.NoError(t, err)

	assert.Equal(t, header1, header2)

}
