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

package stateful

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/core/ledger"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/errors"
	"github.com/polynetwork/poly/validator/db"
	vatypes "github.com/polynetwork/poly/validator/types"
	"reflect"
)

// Validator is an interface for tx validation actor
type Validator interface {
	Register(poolId *actor.PID)
	UnRegister(poolId *actor.PID)
	VerifyType() vatypes.VerifyType
}

type validator struct {
	pid       *actor.PID
	id        string
	bestBlock db.BestBlock
}

// NewValidator returns Validator for stateful check of tx
func NewValidator(id string) (Validator, error) {

	validator := &validator{id: id}
	props := actor.FromProducer(func() actor.Actor {
		return validator
	})

	pid, err := actor.SpawnNamed(props, id)
	validator.pid = pid
	return validator, err
}

func (self *validator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("stateful-validator: started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("stateful-validator: stopping")
	case *actor.Restarting:
		log.Info("stateful-validator: restarting")
	case *vatypes.CheckTx:
		log.Debugf("stateful-validator: receive tx %x", msg.Tx.Hash())
		sender := context.Sender()
		height := ledger.DefLedger.GetCurrentBlockHeight()

		errCode := errors.ErrNoError
		hash := msg.Tx.Hash()

		exist, err := ledger.DefLedger.IsContainTransaction(hash)
		if err != nil {
			log.Warn("query db error:", err)
			errCode = errors.ErrUnknown
		} else if exist {
			errCode = errors.ErrDuplicatedTx
		}

		response := &vatypes.CheckResponse{
			WorkerId: msg.WorkerId,
			Type:     self.VerifyType(),
			Hash:     msg.Tx.Hash(),
			Height:   height,
			ErrCode:  errCode,
		}

		sender.Tell(response)
	case *vatypes.UnRegisterAck:
		context.Self().Stop()
	case *types.Block:

		//bestBlock, _ := self.db.GetBestBlock()
		//if bestBlock.Height+1 < msg.Header.Height {
		//	// add sync block request
		//} else if bestBlock.Height+1 == msg.Header.Height {
		//	self.db.PersistBlock(msg)
		//}

	default:
		log.Info("stateful-validator: unknown msg ", msg, "type", reflect.TypeOf(msg))
	}

}

func (self *validator) VerifyType() vatypes.VerifyType {
	return vatypes.Stateful
}

func (self *validator) Register(poolId *actor.PID) {
	poolId.Tell(&vatypes.RegisterValidator{
		Sender: self.pid,
		Type:   self.VerifyType(),
		Id:     self.id,
	})
}

func (self *validator) UnRegister(poolId *actor.PID) {
	poolId.Tell(&vatypes.UnRegisterValidator{
		Id: self.id,
	})
}
