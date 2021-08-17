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
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/payload"
	"github.com/polynetwork/poly/core/signature"
	"github.com/polynetwork/poly/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Serialize(t *testing.T) {
	tx := &types.Transaction{
		Version:    0,
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte("Chain Id"),
	}

	tx.Version = 1
	sink := common.NewZeroCopySink(nil)
	err := tx.SerializeUnsigned(sink)
	assert.Error(t, err)

	tx.Attributes = []byte{}
	err = tx.SerializeUnsigned(sink)
	assert.Error(t, err)

	tx.Version = 0
	err = tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	acc := account.NewAccount("123456")
	sigData, err := signature.Sign(acc, sink.Bytes())
	assert.NoError(t, err)

	tx.Sigs = []types.Sig{
		{
			SigData: [][]byte{sigData},
			PubKeys: []keypair.PublicKey{acc.PubKey()},
			M:       1,
		},
	}
	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx = new(types.Transaction)
	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}
