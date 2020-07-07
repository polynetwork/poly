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

// Package actor privides communication with other actor
package actor

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/core/genesis"
	scommon "github.com/polynetwork/poly/core/store/common"
	"github.com/polynetwork/poly/core/types"
	ontErrors "github.com/polynetwork/poly/errors"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/relayer_manager"
	"github.com/polynetwork/poly/native/service/utils"
	nutils "github.com/polynetwork/poly/native/service/utils"
	tcomn "github.com/polynetwork/poly/txnpool/common"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
)

var txnPid *actor.PID
var txnPoolPid *actor.PID
var DisableSyncVerifyTx = false

func SetTxPid(actr *actor.PID) {
	txnPid = actr
}
func SetTxnPoolPid(actr *actor.PID) {
	txnPoolPid = actr
}

//append transaction to pool to txpool actor
func AppendTxToPool(txn *types.Transaction) (ontErrors.ErrCode, string) {
	//check if registered relayer
	flag := true
	addresses, err := txn.GetSignatureAddresses()
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}

	//get consensus node address
	governanceViewBytes, err := GetStorageItem(utils.NodeManagerContractAddress, []byte(node_manager.GOVERNANCE_VIEW))
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}
	governanceView := new(node_manager.GovernanceView)
	err = governanceView.Deserialization(common.NewZeroCopySource(governanceViewBytes))
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}
	viewBytes := nutils.GetUint32Bytes(governanceView.View)
	peerPoolMapBytes, err := GetStorageItem(utils.NodeManagerContractAddress, append([]byte(node_manager.PEER_POOL), viewBytes...))
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}
	peerMap := &node_manager.PeerPoolMap{
		PeerPoolMap: make(map[string]*node_manager.PeerPoolItem),
	}
	err = peerMap.Deserialization(common.NewZeroCopySource(peerPoolMapBytes))
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}

	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}
	for _, address := range addresses {
		key := append([]byte(relayer_manager.RELAYER), address[:]...)
		value, err := GetStorageItem(utils.RelayerManagerContractAddress, key)
		if err != nil {
			if err != scommon.ErrNotFound {
				return ontErrors.ErrUnknown, err.Error()
			}
		}
		if value != nil || address == operatorAddress {
			flag = false
			break
		}

		for k := range peerMap.PeerPoolMap {
			kb, err := hex.DecodeString(k)
			if err != nil {
				return ontErrors.ErrUnknown, err.Error()
			}
			pk, err := keypair.DeserializePublicKey(kb)
			if err != nil {
				return ontErrors.ErrUnknown, err.Error()
			}
			addr := types.AddressFromPubKey(pk)
			if address == addr {
				flag = false
				break
			}
		}
		if !flag {
			break
		}
	}
	if flag {
		return ontErrors.ErrUnknown, "address is not registered"
	}

	if DisableSyncVerifyTx {
		txReq := &tcomn.TxReq{txn, tcomn.HttpSender, nil}
		txnPid.Tell(txReq)
		return ontErrors.ErrNoError, ""
	}
	//add Pre Execute Contract
	_, err = PreExecuteContract(txn)
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}
	ch := make(chan *tcomn.TxResult, 1)
	txReq := &tcomn.TxReq{txn, tcomn.HttpSender, ch}
	txnPid.Tell(txReq)
	if msg, ok := <-ch; ok {
		return msg.Err, msg.Desc
	}
	return ontErrors.ErrUnknown, ""
}

//GetTxsFromPool from txpool actor
func GetTxsFromPool(byCount bool) map[common.Uint256]*types.Transaction {
	future := txnPoolPid.RequestFuture(&tcomn.GetTxnPoolReq{ByCount: byCount}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return nil
	}
	txpool, ok := result.(*tcomn.GetTxnPoolRsp)
	if !ok {
		return nil
	}
	txMap := make(map[common.Uint256]*types.Transaction)
	for _, v := range txpool.TxnPool {
		txMap[v.Tx.Hash()] = v.Tx
	}
	return txMap

}

//GetTxFromPool from txpool actor
func GetTxFromPool(hash common.Uint256) (tcomn.TXEntry, error) {

	future := txnPid.RequestFuture(&tcomn.GetTxnReq{hash}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return tcomn.TXEntry{}, err
	}
	rsp, ok := result.(*tcomn.GetTxnRsp)
	if !ok {
		return tcomn.TXEntry{}, errors.New("fail")
	}
	if rsp.Txn == nil {
		return tcomn.TXEntry{}, errors.New("fail")
	}

	future = txnPid.RequestFuture(&tcomn.GetTxnStatusReq{hash}, REQ_TIMEOUT*time.Second)
	result, err = future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return tcomn.TXEntry{}, err
	}
	txStatus, ok := result.(*tcomn.GetTxnStatusRsp)
	if !ok {
		return tcomn.TXEntry{}, errors.New("fail")
	}
	txnEntry := tcomn.TXEntry{rsp.Txn, txStatus.TxStatus}
	return txnEntry, nil
}

//GetTxnCount from txpool actor
func GetTxnCount() ([]uint32, error) {
	future := txnPid.RequestFuture(&tcomn.GetTxnCountReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return []uint32{}, err
	}
	txnCnt, ok := result.(*tcomn.GetTxnCountRsp)
	if !ok {
		return []uint32{}, errors.New("fail")
	}
	return txnCnt.Count, nil
}
