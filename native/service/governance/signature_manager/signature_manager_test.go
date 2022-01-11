/*
 * Copyright (C) 2022 The poly network Authors
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
package signature_manager

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/zeebo/assert"
)

var (
	acct1     *account.Account   = account.NewAccount("")
	acct2     *account.Account   = account.NewAccount("")
	acct3     *account.Account   = account.NewAccount("")
	acct4     *account.Account   = account.NewAccount("")
	acct5     *account.Account   = account.NewAccount("")
	acct6     *account.Account   = account.NewAccount("")
	acct7     *account.Account   = account.NewAccount("")
	acctList1 []*account.Account = []*account.Account{acct1, acct2, acct3, acct4, acct5, acct6, acct7}
)

func Init(db *storage.CacheDB) {
	//put governance view
	var view uint32 = 1
	governanceView := &node_manager.GovernanceView{
		View: view, Height: 1, TxHash: common.Uint256{0},
	}
	contract := utils.NodeManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	governanceView.Serialization(sink)
	db.Put(utils.ConcatKey(contract, []byte(node_manager.GOVERNANCE_VIEW)), cstates.GenRawStorageItem(sink.Bytes()))

	//put peer pool map
	peerPoolMap := &node_manager.PeerPoolMap{
		PeerPoolMap: make(map[string]*node_manager.PeerPoolItem),
	}
	for i, acct := range acctList1 {
		peerPoolMap.PeerPoolMap[hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey))] =
			&node_manager.PeerPoolItem{
				Index:      uint32(i + 1),
				PeerPubkey: hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey)),
				Address:    acct.Address,
				Status:     node_manager.ConsensusStatus,
			}
	}
	contract = utils.NodeManagerContractAddress
	viewBytes := utils.GetUint32Bytes(view)
	sink = common.NewZeroCopySink(nil)
	peerPoolMap.Serialization(sink)
	db.Put(utils.ConcatKey(contract, []byte(node_manager.PEER_POOL), viewBytes), cstates.GenRawStorageItem(sink.Bytes()))
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns, err := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		panic(fmt.Sprintf("NewNativeService error: %+v", err))
	}
	return ns
}

func TestConsensusSig(t *testing.T) {
	store, _ := leveldbstore.NewMemLevelDBStore()
	db := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	Init(db)

	quorum := (len(acctList1)*2 + 2) / 3
	for i, acct := range acctList1 {
		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		param := AddSignatureParam{
			Address:   acct.Address,
			Subject:   []byte("demo"),
			Signature: []byte("sig"),
		}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		nativeService := NewNative(sink.Bytes(), tx, db)
		res, err := AddSignature(nativeService)
		t.Log(err)
		assert.Equal(t, res, []byte{1})
		assert.Nil(t, err)

		if i < quorum-1 {
			assert.Equal(t, 0, len(nativeService.GetNotify()))
		} else {
			assert.Equal(t, 1, len(nativeService.GetNotify()))
			break
		}
	}

}
