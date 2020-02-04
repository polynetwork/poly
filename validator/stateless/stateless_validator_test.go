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
package stateless

import (
	"fmt"
	"github.com/ontio/multi-chain/core/payload"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/errors"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/core/signature"
	ctypes "github.com/ontio/multi-chain/core/types"
	types2 "github.com/ontio/multi-chain/validator/types"
	"github.com/ontio/ontology-crypto/keypair"
)

func signTransaction(signer *account.Account, tx *ctypes.MutableTransaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, ctypes.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func TestStatelessValidator(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	acc := account.NewAccount("")
	invoke := &payload.InvokeCode{}
	mutable := &types.MutableTransaction{
		TxType:  types.Invoke,
		Payload: invoke,
	}
	fmt.Print(mutable.Hash())

	err := signTransaction(acc, mutable)
	assert.Nil(t, err)

	tx, err := mutable.IntoImmutable()
	assert.Nil(t, err)

	validator := &validator{id: "test"}
	props := actor.FromProducer(func() actor.Actor {
		return validator
	})

	pid, err := actor.SpawnNamed(props, validator.id)
	assert.Nil(t, err)

	msg := &types2.CheckTx{WorkerId: 1, Tx: tx}
	fut := pid.RequestFuture(msg, time.Second)

	res, err := fut.Result()
	assert.Nil(t, err)

	result := res.(*types2.CheckResponse)
	assert.Equal(t, result.ErrCode, errors.ErrNoError)
	assert.Equal(t, mutable.Hash(), result.Hash)
}
