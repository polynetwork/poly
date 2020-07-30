/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */
package relayer_manager

import (
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/genesis"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var (
	acct     = account.NewAccount("")
	conAccts = func() []*account.Account {
		accts := make([]*account.Account, 0)
		for i := 0; i < 7; i++ {
			accts = append(accts, account.NewAccount(strconv.FormatUint(uint64(i), 10)))
		}
		return accts
	}
	getNativeFunc = func() *native.NativeService {
		store, _ :=
			leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		ns, _ := native.NewNativeService(cacheDB, new(types.Transaction), 0, 200, common.Uint256{}, 0, nil, false)
		return ns
	}

	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}

	nativeService *native.NativeService
)

func init() {
	setBKers()
}
func putPeerMapPoolAndView(db *storage.CacheDB, conAccts []*account.Account) {
	peerPoolMap := new(node_manager.PeerPoolMap)
	peerPoolMap.PeerPoolMap = make(map[string]*node_manager.PeerPoolItem)
	for i, conAcct := range conAccts {
		pkStr := vconfig.PubkeyID(conAcct.PublicKey)
		peerPoolMap.PeerPoolMap[pkStr] = &node_manager.PeerPoolItem{
			Index:      uint32(i),
			PeerPubkey: pkStr,
			Address:    conAcct.Address,
			Status:     node_manager.ConsensusStatus,
		}
	}
	viewBytes := utils.GetUint32Bytes(0)
	sink := common.NewZeroCopySink(nil)
	peerPoolMap.Serialization(sink)
	db.Put(utils.ConcatKey(utils.NodeManagerContractAddress, []byte(node_manager.PEER_POOL), viewBytes), cstates.GenRawStorageItem(sink.Bytes()))

	govView := node_manager.GovernanceView{
		0, 10, common.UINT256_EMPTY,
	}
	sink = common.NewZeroCopySink(nil)
	govView.Serialization(sink)
	db.Put(utils.ConcatKey(utils.NodeManagerContractAddress, []byte(node_manager.GOVERNANCE_VIEW)), cstates.GenRawStorageItem(sink.Bytes()))
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns, _ := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	return ns
}

func TestRegisterRelayer(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = []common.Address{{1, 2, 4, 6}, {1, 4, 5, 7}, {1, 3, 5, 7, 9}}
	params.Address = acct.Address
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)
	putPeerMapPoolAndView(nativeService.GetCacheDB(), conAccts())

	res, err := RegisterRelayer(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	relayerListParam, err := getRelayerApply(nativeService, 0)
	assert.Nil(t, err)
	assert.Equal(t, params, relayerListParam)

	// none consensus acct should not be able to approve register relayer
	notConAcct := account.NewAccount("x")
	arp := &ApproveRelayerParam{
		0,
		notConAcct.Address,
	}
	sink = common.NewZeroCopySink(nil)
	arp.Serialization(sink)
	tx = &types.Transaction{
		SignedAddr: []common.Address{notConAcct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err = ApproveRegisterRelayer(nativeService)
	assert.Nil(t, err)
	assert.Equal(t, utils.BYTE_TRUE, res)

	for i, conAcct := range conAccts() {
		arp := &ApproveRelayerParam{
			0,
			conAcct.Address,
		}
		sink := common.NewZeroCopySink(nil)
		arp.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{conAcct.Address},
		}
		nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
		res, err := ApproveRegisterRelayer(nativeService)
		assert.Nil(t, err)
		assert.Equal(t, utils.BYTE_TRUE, res)
		if i < (2*len(conAccts())+2)/3 {
			ok, err := node_manager.CheckConsensusSigns(nativeService, APPROVE_REGISTER_RELAYER, utils.GetUint64Bytes(0), conAcct.Address)
			assert.Nil(t, err)
			assert.Equal(t, false, ok)
		}
	}
}

func TestRemoveRelayer(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = []common.Address{{1, 2, 4, 6}, {1, 4, 5, 7}}
	params.Address = acct.Address
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)
	putPeerMapPoolAndView(nativeService.GetCacheDB(), conAccts())

	res, err := RemoveRelayer(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	relayerRaw1, err := getRelayerRemove(nativeService, 0)
	assert.Nil(t, err)
	assert.Equal(t, params, relayerRaw1)

	for i, conAcct := range conAccts() {
		arp := &ApproveRelayerParam{
			0,
			conAcct.Address,
		}
		sink := common.NewZeroCopySink(nil)
		arp.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{conAcct.Address},
		}
		nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
		res, err := ApproveRemoveRelayer(nativeService)
		assert.Nil(t, err)
		assert.Equal(t, utils.BYTE_TRUE, res)
		if i < (2*len(conAccts())+2)/3 {
			ok, err := node_manager.CheckConsensusSigns(nativeService, APPROVE_REGISTER_RELAYER, utils.GetUint64Bytes(0), conAcct.Address)
			assert.Nil(t, err)
			assert.Equal(t, false, ok)
		}
	}
}
