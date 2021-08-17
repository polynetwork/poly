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

// Package actor privides communication with other actor
package actor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	scommon "github.com/polynetwork/poly/core/store/common"
	"github.com/polynetwork/poly/core/types"
	polyErrors "github.com/polynetwork/poly/errors"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/relayer_manager"
	"github.com/polynetwork/poly/native/service/utils"
	nutils "github.com/polynetwork/poly/native/service/utils"
	tcomn "github.com/polynetwork/poly/txnpool/common"
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
func AppendTxToPool(txn *types.Transaction) (polyErrors.ErrCode, string) {
	// Get txn's signature addresses
	addresses, err := txn.GetSignatureAddresses()
	if err != nil {
		return polyErrors.ErrUnknown, err.Error()
	}

	permittedAddrMap := make(map[common.Address]bool)
	// flag is set to true, meaning not any address in the signed addresses is permitted.
	flag := true
	for _, address := range addresses {
		key := append([]byte(relayer_manager.RELAYER), address[:]...)
		value, err := GetStorageItem(utils.RelayerManagerContractAddress, key)
		if err != nil {
			if err != scommon.ErrNotFound {
				return polyErrors.ErrUnknown, err.Error()
			}
		}
		// Check if address is registreed relayer
		if value != nil {
			// Here means address is registered relayer
			flag = false
			break
		}
		// Check if permittedAddrMap is empty
		if len(permittedAddrMap) == 0 {
			// If empty, permittedAddrMap will be updated only one time
			if err := UpdatePermittedAddrMap(permittedAddrMap); err != nil {
				return polyErrors.ErrUnknown, err.Error()
			}
		}
		// Check if address is included in permittedAddrMap, if so, the txn is permitted
		if val, ok := permittedAddrMap[address]; val && ok {
			flag = false
			break
		}
	}
	if flag {
		// If flag is true, it means any address within addresses is not permitted address to send tx
		return polyErrors.ErrUnknown, "address is not registered"
	}
	if DisableSyncVerifyTx {
		txReq := &tcomn.TxReq{txn, tcomn.HttpSender, nil}
		txnPid.Tell(txReq)
		return polyErrors.ErrNoError, ""
	}
	//add Pre Execute Contract
	_, err = PreExecuteContract(txn)
	if err != nil {
		return polyErrors.ErrUnknown, err.Error()
	}
	ch := make(chan *tcomn.TxResult, 1)
	txReq := &tcomn.TxReq{txn, tcomn.HttpSender, ch}
	txnPid.Tell(txReq)
	if msg, ok := <-ch; ok {
		return msg.Err, msg.Desc
	}
	return polyErrors.ErrUnknown, ""
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

func UpdatePermittedAddrMap(permittedAddrMap map[common.Address]bool) error {
	//get consensus node address
	governanceViewBytes, err := GetStorageItem(utils.NodeManagerContractAddress, []byte(node_manager.GOVERNANCE_VIEW))
	if err != nil {
		return fmt.Errorf("UpdatePermittedAddrMap, get governance view bytes from storage error: %v", err)
	}
	governanceView := new(node_manager.GovernanceView)
	err = governanceView.Deserialization(common.NewZeroCopySource(governanceViewBytes))
	if err != nil {
		return fmt.Errorf("UpdatePermittedAddrMap, governanceView.Deserialization error: %v", err)
	}
	viewBytes := nutils.GetUint32Bytes(governanceView.View)
	peerPoolMapBytes, err := GetStorageItem(utils.NodeManagerContractAddress, append([]byte(node_manager.PEER_POOL), viewBytes...))
	if err != nil {
		return fmt.Errorf("UpdatePermittedAddrMap, get peerPoolMap bytes from storage error: %v", err)
	}
	peerMap := &node_manager.PeerPoolMap{
		PeerPoolMap: make(map[string]*node_manager.PeerPoolItem),
	}
	err = peerMap.Deserialization(common.NewZeroCopySource(peerPoolMapBytes))
	if err != nil {
		return fmt.Errorf("UpdatePermittedAddrMap, peerPoolMap.Deserialization error: %v", err)
	}
	// Update permittedAddrMap
	publicKeys := make([]keypair.PublicKey, 0)
	for k := range peerMap.PeerPoolMap {
		kb, err := hex.DecodeString(k)
		if err != nil {
			return fmt.Errorf("UpdatePermittedAddrMap, DecodeString PeerPoolMap public key error: %v", err)
		}
		pk, err := keypair.DeserializePublicKey(kb)
		if err != nil {
			return fmt.Errorf("UpdatePermittedAddrMap, DeserializePublicKey error: %v", err)
		}
		permittedAddrMap[types.AddressFromPubKey(pk)] = true
		publicKeys = append(publicKeys, pk)
	}
	operatorAddress, err := types.AddressFromBookkeepers(publicKeys)
	if err != nil {
		return fmt.Errorf("UpdatePermittedAddrMap, AddressFromBookkeepers error: %v", err)
	}
	permittedAddrMap[operatorAddress] = true
	return nil
}
