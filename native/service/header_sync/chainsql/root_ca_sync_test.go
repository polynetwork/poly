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
package chainsql

import (
	"encoding/pem"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	common2 "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	acct = account.NewAccount("")

	getNativeFunc = func(args []byte, db *storage.CacheDB) *native.NativeService {
		if db == nil {
			store, _ := leveldbstore.NewMemLevelDBStore()
			db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
			sink := common.NewZeroCopySink(nil)
			view := &node_manager.GovernanceView{
				TxHash: common.UINT256_EMPTY,
				Height: 0,
				View:   0,
			}
			view.Serialization(sink)
			db.Put(utils.ConcatKey(utils.NodeManagerContractAddress, []byte(node_manager.GOVERNANCE_VIEW)), states.GenRawStorageItem(sink.Bytes()))

			peerPoolMap := &node_manager.PeerPoolMap{
				PeerPoolMap: map[string]*node_manager.PeerPoolItem{
					vconfig.PubkeyID(acct.PublicKey): {
						Address:    acct.Address,
						Status:     node_manager.ConsensusStatus,
						PeerPubkey: vconfig.PubkeyID(acct.PublicKey),
						Index:      0,
					},
				},
			}
			sink.Reset()
			peerPoolMap.Serialization(sink)
			db.Put(utils.ConcatKey(utils.NodeManagerContractAddress,
				[]byte(node_manager.PEER_POOL), utils.GetUint32Bytes(0)), states.GenRawStorageItem(sink.Bytes()))
		}
		signAddr, _ := types.AddressFromBookkeepers([]keypair.PublicKey{acct.PublicKey})
		ns, _ := native.NewNativeService(db, &types.Transaction{SignedAddr: []common.Address{signAddr}}, 1600945402, 0, common.Uint256{0}, 0, args, false)
		return ns
	}

	rootCA = `-----BEGIN CERTIFICATE-----
MIIBxDCCAWqgAwIBAgIJAL0TiHo7dsqWMAoGCCqBHM9VAYN1MDcxEDAOBgNVBAMM
B2dtY2hhaW4xEzARBgNVBAoMCmZpc2NvLWJjb3MxDjAMBgNVBAsMBWNoYWluMCAX
DTIwMDkyNDExMDMyMVoYDzIxMjAwODMxMTEwMzIxWjA3MRAwDgYDVQQDDAdnbWNo
YWluMRMwEQYDVQQKDApmaXNjby1iY29zMQ4wDAYDVQQLDAVjaGFpbjBZMBMGByqG
SM49AgEGCCqBHM9VAYItA0IABEIf/hJjT6DAGYWCyP99sBoTF2cCqpbLsrOf+NwY
KY0zdXUA9BwYCs+HyoSLRtZBlfa5hO5S6wDbU1l9472aYFijXTBbMB0GA1UdDgQW
BBQILjttksRAgGbi4KMNrLHlwPhxkTAfBgNVHSMEGDAWgBQILjttksRAgGbi4KMN
rLHlwPhxkTAMBgNVHRMEBTADAQH/MAsGA1UdDwQEAwIBBjAKBggqgRzPVQGDdQNI
ADBFAiBKFFaclfd0IKplJgLXDdAxS1Cvwhl/ZOFwPq28V2wi8gIhAPaAT8qf1hUv
9FGgtdQbPr/lerRDGOETv5Zi5GEJBOpA
-----END CERTIFICATE-----`
)

func TestChainsqlHandler_SyncGenesisHeader(t *testing.T) {
	params := new(common2.SyncGenesisHeaderParam)
	params.ChainID = 6

	params.GenesisHeader = []byte(rootCA)
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)
	ns := getNativeFunc(sink.Bytes(), nil)

	h := NewChainsqlHandler()

	// sync ca
	if err := h.SyncGenesisHeader(ns); err != nil {
		t.Fatal(err)
	}

	root, err := GetChainsqlRoot(ns, 6)
	if err != nil {
		t.Fatal(err)
	}

	blk, _ := pem.Decode([]byte(rootCA))
	if blk == nil {
		t.Fatal("failed to decode pem")
	}
	assert.Equal(t, blk.Bytes, root.RootCA.Raw, "wrong base64-encoded bytes")

	// next successful to update
	params.GenesisHeader = []byte(rootCA)
	sink.Reset()
	params.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = h.SyncGenesisHeader(ns)
	if err != nil {
		t.Fatal(err)
	}
}
