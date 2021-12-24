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

package proc

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	scommon "github.com/polynetwork/poly/core/store/common"
	tx "github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/errors"
	"github.com/polynetwork/poly/events/message"
	bactor "github.com/polynetwork/poly/http/base/actor"
	"github.com/polynetwork/poly/native/service/governance/relayer_manager"
	"github.com/polynetwork/poly/native/service/utils"
	tc "github.com/polynetwork/poly/txnpool/common"
	"github.com/polynetwork/poly/validator/types"
)

// NewTxActor creates an actor to handle the transaction-based messages from
// network and http
func NewTxActor(s *TXPoolServer) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

// NewTxPoolActor creates an actor to handle the messages from the consensus
func NewTxPoolActor(s *TXPoolServer) *TxPoolActor {
	a := &TxPoolActor{}
	a.setServer(s)
	return a
}

// NewVerifyRspActor creates an actor to handle the verified result from validators
func NewVerifyRspActor(s *TXPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(s)
	return a
}

func replyTxResult(txResultCh chan *tc.TxResult, hash common.Uint256,
	err errors.ErrCode, desc string) {
	result := &tc.TxResult{
		Err:  err,
		Hash: hash,
		Desc: desc,
	}
	select {
	case txResultCh <- result:
	default:
		log.Debugf("handleTransaction: duplicated result")
	}
}

// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	server *TXPoolServer
}

func (ta *TxActor) isValidSender(txn *tx.Transaction, permittedAddrMap map[common.Address]bool) (err error) {
	lock.RLock()
	defer lock.RUnlock()

	// Get txn's signature addresses
	addresses, err := txn.GetSignatureAddresses()
	if err != nil {
		return
	}

	// flag is set to true, meaning not any address in the signed addresses is permitted.
	flag := true
	for _, address := range addresses {
		key := append([]byte(relayer_manager.RELAYER), address[:]...)
		value, err := bactor.GetStorageItem(utils.RelayerManagerContractAddress, key)
		if err != nil {
			if err != scommon.ErrNotFound {
				return err
			}
		}
		// Check if address is registreed relayer
		if len(value) > 0 {
			// Here means address is registered relayer
			flag = false
			break
		}
		// Check if address is included in permittedAddrMap, if so, the txn is permitted
		if val, ok := permittedAddrMap[address]; val && ok {
			flag = false
			break
		}
	}
	if flag {
		// If flag is true, it means any address within addresses is not permitted address to send tx
		return fmt.Errorf("address is not registered")
	}
	return nil
}

var permittedAddrMap = make(map[common.Address]bool)
var lastTime int64
var lock sync.RWMutex

func updatePermittedAddrMap() (err error) {
	if lastTime == 0 || len(permittedAddrMap) == 0 || lastTime < time.Now().Add(-time.Minute).Unix() {
		lock.Lock()
		defer lock.Unlock()
		lastTime = time.Now().Unix()
		if err = bactor.UpdatePermittedAddrMap(permittedAddrMap); err != nil {
			log.Debugf("updatePermittedAddrMap failed")
			return
		}
	}
	return
}

// handleTransaction handles a transaction from network and http
func (ta *TxActor) handleTransaction(sender tc.SenderType, self *actor.PID,
	txn *tx.Transaction, txResultCh chan *tc.TxResult) {

	err := updatePermittedAddrMap()
	if err != nil {
		log.Debugf("handleTransaction: UpdatePermittedAddrMap failed for tx %x",
			txn.Hash())
		return
	}

	err = ta.isValidSender(txn, permittedAddrMap)
	if err != nil {
		log.Debugf("handleTransaction: invalid sender for tx %x",
			txn.Hash())
		return
	}
	ta.server.increaseStats(tc.RcvStats)
	if len(txn.ToArray()) > tc.MAX_TX_SIZE {
		log.Debugf("handleTransaction: reject a transaction due to size over 1M")
		if sender == tc.HttpSender && txResultCh != nil {
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, "size is over 1M")
		}
		return
	}

	if ta.server.getTransaction(txn.Hash()) != nil {
		log.Debugf("handleTransaction: transaction %x already in the txn pool",
			txn.Hash())

		ta.server.increaseStats(tc.DuplicateStats)
		if sender == tc.HttpSender && txResultCh != nil {
			replyTxResult(txResultCh, txn.Hash(), errors.ErrDuplicateInput,
				fmt.Sprintf("transaction %x is already in the tx pool", txn.Hash()))
		}
	} else if ta.server.getTransactionCount() >= tc.MAX_CAPACITY {
		log.Debugf("handleTransaction: transaction pool is full for tx %x",
			txn.Hash())

		ta.server.increaseStats(tc.FailureStats)
		if sender == tc.HttpSender && txResultCh != nil {
			replyTxResult(txResultCh, txn.Hash(), errors.ErrTxPoolFull,
				"transaction pool is full")
		}
	} else {
		<-ta.server.slots
		ta.server.assignTxToWorker(txn, sender, txResultCh)
	}
}

// Receive implements the actor interface
func (ta *TxActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool-tx actor started and be ready to receive tx msg")

	case *actor.Stopping:
		log.Warn("txpool-tx actor stopping")

	case *actor.Restarting:
		log.Warn("txpool-tx actor restarting")

	case *tc.TxReq:
		sender := msg.Sender

		log.Debugf("txpool-tx actor receives tx from %v ", sender.Sender())

		ta.handleTransaction(sender, context.Self(), msg.Tx, msg.TxResultCh)

	case *tc.GetTxnReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx req from %v", sender)

		res := ta.server.getTransaction(msg.Hash)
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: res},
				context.Self())
		}

	case *tc.GetTxnStats:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx stats from %v", sender)

		res := ta.server.getStats()
		if sender != nil {
			sender.Request(&tc.GetTxnStatsRsp{Count: res},
				context.Self())
		}

	case *tc.CheckTxnReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives checking tx req from %v", sender)

		res := ta.server.checkTx(msg.Hash)
		if sender != nil {
			sender.Request(&tc.CheckTxnRsp{Ok: res},
				context.Self())
		}

	case *tc.GetTxnStatusReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx status req from %v", sender)

		res := ta.server.getTxStatusReq(msg.Hash)
		if sender != nil {
			if res == nil {
				sender.Request(&tc.GetTxnStatusRsp{Hash: msg.Hash,
					TxStatus: nil}, context.Self())
			} else {
				sender.Request(&tc.GetTxnStatusRsp{Hash: res.Hash,
					TxStatus: res.Attrs}, context.Self())
			}
		}

	case *tc.GetTxnCountReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx count req from %v", sender)

		res := ta.server.getTxCount()
		if sender != nil {
			sender.Request(&tc.GetTxnCountRsp{Count: res},
				context.Self())
		}

	default:
		log.Debugf("txpool-tx actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

func (ta *TxActor) setServer(s *TXPoolServer) {
	ta.server = s
}

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	server *TXPoolServer
}

// Receive implements the actor interface
func (tpa *TxPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool actor started and be ready to receive txPool msg")

	case *actor.Stopping:
		log.Warn("txpool actor stopping")

	case *actor.Restarting:
		log.Warn("txpool actor Restarting")

	case *tc.GetTxnPoolReq:
		sender := context.Sender()

		log.Debugf("txpool actor receives getting tx pool req from %v", sender)

		res := tpa.server.getTxPool(msg.ByCount, msg.Height)
		if sender != nil {
			sender.Request(&tc.GetTxnPoolRsp{TxnPool: res}, context.Self())
		}

	case *tc.GetPendingTxnReq:
		sender := context.Sender()

		log.Debugf("txpool actor receives getting pedning tx req from %v", sender)

		res := tpa.server.getPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&tc.GetPendingTxnRsp{Txs: res}, context.Self())
		}

	case *tc.VerifyBlockReq:
		sender := context.Sender()

		log.Debugf("txpool actor receives verifying block req from %v", sender)

		tpa.server.verifyBlock(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()

		log.Debugf("txpool actor receives block complete event from %v", sender)

		if msg.Block != nil {
			tpa.server.cleanTransactionList(msg.Block.Transactions, msg.Block.Header.Height)
		}

	default:
		log.Debugf("txpool actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

func (tpa *TxPoolActor) setServer(s *TXPoolServer) {
	tpa.server = s
}

// VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *TXPoolServer
}

// Receive implements the actor interface
func (vpa *VerifyRspActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool-verify actor: started and be ready to receive validator's msg")

	case *actor.Stopping:
		log.Warn("txpool-verify actor: stopping")

	case *actor.Restarting:
		log.Warn("txpool-verify actor: Restarting")

	case *types.RegisterValidator:
		log.Debugf("txpool-verify actor:: validator %v connected", msg.Sender)
		vpa.server.registerValidator(msg)

	case *types.UnRegisterValidator:
		log.Debugf("txpool-verify actor:: validator %d:%v disconnected", msg.Type, msg.Id)

		vpa.server.unRegisterValidator(msg.Type, msg.Id)

	case *types.CheckResponse:
		log.Debug("txpool-verify actor:: Receives verify rsp message")

		vpa.server.assignRspToWorker(msg)

	default:
		log.Debugf("txpool-verify actor:Unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

func (vpa *VerifyRspActor) setServer(s *TXPoolServer) {
	vpa.server = s
}
