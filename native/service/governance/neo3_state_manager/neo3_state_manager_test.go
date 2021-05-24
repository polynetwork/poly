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

package neo3_state_manager

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
	"log"
	"strconv"
	"testing"
)

var (
	acct     = account.NewAccount("")
	conAccts = func() []*account.Account {
		accts := make([]*account.Account, 0, 7)
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

func TestRegisterStateValidator(t *testing.T) {
	params := new(StateValidatorListParam)
	// four state validators
	params.StateValidators = []string{
		"023e9b32ea89b94d066e649b124fd50e396ee91369e8e2a6ae1b11c170d022256d",
		"03009b7540e10f2562e5fd8fac9eaec25166a58b26e412348ff5a86927bfac22a2",
		"02ba2c70f5996f357a43198705859fae2cfea13e1172962800772b3d588a9d4abd",
		"03408dcd416396f64783ac587ea1e1593c57d9fea880c8a6a1920e92a259477806",
	}
	params.Address = acct.Address
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)
	consensusAccounts := conAccts()
	putPeerMapPoolAndView(nativeService.GetCacheDB(), consensusAccounts)
	// register
	res, err := RegisterStateValidator(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	svListParam, err := getStateValidatorApply(nativeService, 0)
	assert.Nil(t, err)
	assert.Equal(t, params, svListParam)

	// none consensus acct should not be able to approve register
	notConAcct := account.NewAccount("x")
	asvp := &ApproveStateValidatorParam{
		0,
		notConAcct.Address,
	}
	sink = common.NewZeroCopySink(nil)
	asvp.Serialization(sink)
	tx = &types.Transaction{
		SignedAddr: []common.Address{notConAcct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err = ApproveRegisterStateValidator(nativeService)
	assert.Nil(t, err)
	assert.Equal(t, utils.BYTE_FALSE, res)

	for i, conAcct := range consensusAccounts {
		asvp := &ApproveStateValidatorParam{
			0,
			conAcct.Address,
		}
		sink := common.NewZeroCopySink(nil)
		asvp.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{conAcct.Address},
		}
		nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
		res, err := ApproveRegisterStateValidator(nativeService)
		assert.Nil(t, err)

		if i < (2*len(consensusAccounts)+2)/3-1 { // not enough sig
			assert.Equal(t, utils.BYTE_FALSE, res)
		} else {
			assert.Equal(t, utils.BYTE_TRUE, res)
			break
		}
	}
}

func TestGetCurrentStateValidator(t *testing.T) {
	params := new(StateValidatorListParam)
	// four state validators
	params.StateValidators = []string{
		"023e9b32ea89b94d066e649b124fd50e396ee91369e8e2a6ae1b11c170d022256d",
		"03009b7540e10f2562e5fd8fac9eaec25166a58b26e412348ff5a86927bfac22a2",
		"02ba2c70f5996f357a43198705859fae2cfea13e1172962800772b3d588a9d4abd",
		"03408dcd416396f64783ac587ea1e1593c57d9fea880c8a6a1920e92a259477806",
	}
	params.Address = acct.Address
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)
	consensusAccounts := conAccts()
	putPeerMapPoolAndView(nativeService.GetCacheDB(), consensusAccounts)

	contract := utils.Neo3StateManagerContractAddress
	// clear storage
	nativeService.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(STATE_VALIDATOR)))

	// register
	res, err := RegisterStateValidator(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	// approve
	for i, conAcct := range consensusAccounts {
		log.Println(i)
		asvp := &ApproveStateValidatorParam{
			0,
			conAcct.Address,
		}
		sink := common.NewZeroCopySink(nil)
		asvp.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{conAcct.Address},
		}
		nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
		res, err := ApproveRegisterStateValidator(nativeService)
		assert.Nil(t, err)
		if i < (2*len(consensusAccounts)+2)/3-1 { // not enough sig
			assert.Equal(t, utils.BYTE_FALSE, res)
		} else {
			assert.Equal(t, utils.BYTE_TRUE, res)
			break
		}
		//if i < (2*len(consensusAccounts)+2)/3 - 1 { // not enough sig
		//	ok, err := node_manager.CheckConsensusSigns(nativeService, APPROVE_REGISTER_STATE_VALIDATOR, utils.GetUint64Bytes(0), conAcct.Address)
		//	assert.Nil(t, err)
		//	assert.Equal(t, false, ok)
		//}
	}
	// get
	svBytes, err := GetCurrentStateValidator(nativeService)
	assert.Nil(t, err)
	assert.Equal(t, 269, len(svBytes))
}

func TestRemoveStateValidator(t *testing.T) {
	params := new(StateValidatorListParam)
	// four state validators
	params.StateValidators = []string{
		"023e9b32ea89b94d066e649b124fd50e396ee91369e8e2a6ae1b11c170d022256d",
		"03009b7540e10f2562e5fd8fac9eaec25166a58b26e412348ff5a86927bfac22a2",
		"02ba2c70f5996f357a43198705859fae2cfea13e1172962800772b3d588a9d4abd",
		"03408dcd416396f64783ac587ea1e1593c57d9fea880c8a6a1920e92a259477806",
	}
	params.Address = acct.Address
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil) // for register
	consensusAccounts := conAccts()
	putPeerMapPoolAndView(nativeService.GetCacheDB(), consensusAccounts)

	// first register
	res, err := RegisterStateValidator(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	// approve register
	for i, conAcct := range consensusAccounts {
		asvp := &ApproveStateValidatorParam{
			0,
			conAcct.Address,
		}
		sink := common.NewZeroCopySink(nil)
		asvp.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{conAcct.Address},
		}
		nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
		res, err := ApproveRegisterStateValidator(nativeService)
		assert.Nil(t, err)

		if i < (2*len(consensusAccounts)+2)/3-1 { // not enough sig
			assert.Equal(t, utils.BYTE_FALSE, res)
		} else {
			assert.Equal(t, utils.BYTE_TRUE, res)
			break
		}
	}

	// confirm state validators in storage
	svBytes, err := GetCurrentStateValidator(nativeService)
	assert.Nil(t, err)
	assert.Equal(t, 269, len(svBytes))

	// remove
	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB()) // for remove
	res, err = RemoveStateValidator(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	svRaw1, err := getStateValidatorRemove(nativeService, 0)
	assert.Nil(t, err)
	assert.Equal(t, params, svRaw1)

	// approve remove
	for i, conAcct := range consensusAccounts {
		asvp := &ApproveStateValidatorParam{
			0,
			conAcct.Address,
		}
		sink := common.NewZeroCopySink(nil)
		asvp.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{conAcct.Address},
		}
		nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
		res, err := ApproveRemoveStateValidator(nativeService)
		assert.Nil(t, err)

		if i < (2*len(consensusAccounts)+2)/3-1 { // not enough sig
			assert.Equal(t, utils.BYTE_FALSE, res)
		} else {
			assert.Equal(t, utils.BYTE_TRUE, res)
			break
		}
	}
}
