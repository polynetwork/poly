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
package native

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/merkle"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/states"
	"github.com/ontio/multi-chain/native/storage"
)

type (
	Handler         func(native *NativeService) ([]byte, error)
	RegisterService func(native *NativeService)
)

var (
	Contracts = make(map[common.Address]RegisterService)
)

// Native service struct
// Invoke a native smart contract, new a native service
type NativeService struct {
	cacheDB       *storage.CacheDB
	serviceMap    map[string]Handler
	notifications []*event.NotifyEventInfo
	invokeParam   states.ContractInvokeParam
	input         []byte
	tx            *types.Transaction
	height        uint32
	time          uint32
	blockHash     common.Uint256
	crossHashes   *common.ZeroCopySink
	preExec       bool
}

func NewNativeService(cacheDB *storage.CacheDB, tx *types.Transaction,
	time, height uint32, blockHash common.Uint256, preExec bool) (*NativeService, error) {
	service := &NativeService{
		cacheDB:    cacheDB,
		tx:         tx,
		time:       time,
		height:     height,
		blockHash:  blockHash,
		serviceMap: make(map[string]Handler),
		preExec:    preExec,
	}
	return service, nil
}

func (this *NativeService) Register(methodName string, handler Handler) {
	this.serviceMap[methodName] = handler
}

func (this *NativeService) Invoke() (interface{}, error) {
	contract := this.invokeParam
	services, ok := Contracts[contract.Address]
	if !ok {
		return false, fmt.Errorf("[Invoke] Native contract address %x haven't been registered.", contract.Address)
	}
	services(this)
	service, ok := this.serviceMap[contract.Method]
	if !ok {
		return false, fmt.Errorf("[Invoke] Native contract %x doesn't support this function %s.",
			contract.Address, contract.Method)
	}
	this.input = contract.Args
	result, err := service(this)
	if err != nil {
		return result, fmt.Errorf("[Invoke] Native serivce function execute error!")
	}
	return result, nil
}

func (this *NativeService) PutMerkleVal(data []byte) {
	this.crossHashes.WriteHash(merkle.HashLeaf(data))
}

// CheckWitness check whether authorization correct
func (this *NativeService) CheckWitness(address common.Address) bool {
	addresses, err := this.tx.GetSignatureAddresses()
	if err != nil {
		log.Errorf("get signature address error:%v", err)
		return false
	}
	for _, v := range addresses {
		if v == address {
			return true
		}
	}
	return false
}

func (this *NativeService) AddNotify(notify *event.NotifyEventInfo) {
	this.notifications = append(this.notifications, notify)
}

func (this *NativeService) GetCacheDB() *storage.CacheDB {
	return this.cacheDB
}

func (this *NativeService) GetInput() []byte {
	return this.input
}

func (this *NativeService) GetTx() *types.Transaction {
	return this.tx
}