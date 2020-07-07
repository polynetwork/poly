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
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/genesis"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/storage"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	acct          *account.Account = account.NewAccount("")
	getNativeFunc                  = func() *native.NativeService {
		store, _ :=
			leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, nil, false)
		return service
	}

	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}

	nativeService *native.NativeService
)

func init() {
	setBKers()
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	return native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
}

func TestRegisterRelayer(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = [][]byte{{1, 2, 4, 6}, {1, 4, 5, 7}, {1, 3, 5, 7, 9}}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)

	res, err := RegisterRelayer(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	res, err = RegisterRelayer(nativeService)
	assert.Equal(t, res, []byte{0})
	assert.NotNil(t, err)
}

func TestRemoveRelayer(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = [][]byte{{1, 2, 4, 6}, {1, 4, 5, 7}}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err := RemoveRelayer(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	address1 := []byte{1, 3, 5, 7, 9}
	relayerRaw1, err := GetRelayerRaw(nativeService, address1)
	assert.Nil(t, err)
	assert.NotNil(t, relayerRaw1)

	address2 := []byte{1, 4, 5, 7}
	relayerRaw2, err := GetRelayerRaw(nativeService, address2)
	assert.Nil(t, err)
	assert.Nil(t, relayerRaw2)
}
