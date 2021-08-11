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

package req

import (
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/core/types"
	p2pcommon "github.com/polynetwork/poly/p2pserver/common"
	tc "github.com/polynetwork/poly/txnpool/common"
)

const txnPoolReqTimeout = p2pcommon.ACTOR_TIMEOUT * time.Second

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

//add txn to txnpool
func AddTransaction(transaction *types.Transaction) {
	if txnPoolPid == nil {
		log.Error("[p2p]net_server AddTransaction(): txnpool pid is nil")
		return
	}
	txReq := &tc.TxReq{
		Tx:         transaction,
		Sender:     tc.NetSender,
		TxResultCh: nil,
	}
	txnPoolPid.Tell(txReq)
}
