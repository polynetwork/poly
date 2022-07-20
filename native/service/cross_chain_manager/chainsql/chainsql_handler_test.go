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
	"crypto/ecdsa"
	"crypto/rand"
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
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	chainsql2 "github.com/polynetwork/poly/native/service/header_sync/chainsql"
	common3 "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/tjfoc/gmsm/pkcs12"
	"github.com/tjfoc/gmsm/sm2"
	"gotest.tools/assert"
	"testing"
	"time"
)

var (
	acct = account.NewAccount("")

	getNativeFunc = func(args []byte, db *storage.CacheDB, cert string) *native.NativeService {
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

			blk, _ := pem.Decode([]byte(cert))
			cert, _ := sm2.ParseCertificate(blk.Bytes)
			root := &chainsql2.ChainsqlRoot{
				RootCA: cert,
			}
			sink = common.NewZeroCopySink(nil)
			root.Serialization(sink)
			db.Put(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(common3.ROOT_CERT), utils.GetUint64Bytes(6)),
				states.GenRawStorageItem(sink.Bytes()))
		}
		signAddr, _ := types.AddressFromBookkeepers([]keypair.PublicKey{acct.PublicKey})
		ns, _ := native.NewNativeService(db, &types.Transaction{SignedAddr: []common.Address{signAddr}}, uint32(time.Now().Unix()), 0, common.Uint256{0}, 0, args, false)
		return ns
	}

	// secp256k1 for pubkey
	rootCA = `-----BEGIN CERTIFICATE-----
MIIBYDCCAQUCCQDG5aP//Iy2ITAKBggqhkjOPQQDAjA5MRwwGgYDVQQDDBN3d3cu
cGVlcnNhZmUuY25DPVVTMRkwFwYDVQQHDBBCZWlqaW5nIENIQU9ZQU5HMB4XDTIy
MDcxMjExMTUxM1oXDTIzMDcxMjExMTUxM1owOTEcMBoGA1UEAwwTd3d3LnBlZXJz
YWZlLmNuQz1VUzEZMBcGA1UEBwwQQmVpamluZyBDSEFPWUFORzBWMBAGByqGSM49
AgEGBSuBBAAKA0IABOqcH14jODlTr/y2ji3U0Z8ENzF4I77Uc5wY8jKZ1+rgQKZ7
0XXEoc7YUKkh6DRrsC66TSbHKlOKePISpJ3xa2MwCgYIKoZIzj0EAwIDSQAwRgIh
AOJoC9uSfOGl/x2wnwW5xUpRTJoQA+zIwBih63HvvdyBAiEAqF3IlgFHPBAShWGk
2A7kpLBW2/MnT12QyC26hvth2gM=
-----END CERTIFICATE-----`

	agencyCA = `-----BEGIN CERTIFICATE-----
MIIBYDCCAQUCCQDG5aP//Iy2ITAKBggqhkjOPQQDAjA5MRwwGgYDVQQDDBN3d3cu
cGVlcnNhZmUuY25DPVVTMRkwFwYDVQQHDBBCZWlqaW5nIENIQU9ZQU5HMB4XDTIy
MDcxMjExMTUxM1oXDTIzMDcxMjExMTUxM1owOTEcMBoGA1UEAwwTd3d3LnBlZXJz
YWZlLmNuQz1VUzEZMBcGA1UEBwwQQmVpamluZyBDSEFPWUFORzBWMBAGByqGSM49
AgEGBSuBBAAKA0IABOqcH14jODlTr/y2ji3U0Z8ENzF4I77Uc5wY8jKZ1+rgQKZ7
0XXEoc7YUKkh6DRrsC66TSbHKlOKePISpJ3xa2MwCgYIKoZIzj0EAwIDSQAwRgIh
AOJoC9uSfOGl/x2wnwW5xUpRTJoQA+zIwBih63HvvdyBAiEAqF3IlgFHPBAShWGk
2A7kpLBW2/MnT12QyC26hvth2gM=
-----END CERTIFICATE-----`

	nodeCA = `-----BEGIN CERTIFICATE-----
MIICOjCCAeGgAwIBAgIJAK29Ojq6GKR/MAoGCCqGSM49BAMCMDkxHDAaBgNVBAMM
E3d3dy5wZWVyc2FmZS5jbkM9VVMxGTAXBgNVBAcMEEJlaWppbmcgQ0hBT1lBTkcw
HhcNMjIwNzEzMDIxNzI0WhcNMjMwNzEzMDIxNzI0WjCBgDELMAkGA1UEBhMCVVMx
EzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFuc2lzY28xETAP
BgNVBAoMCE1Mb3BzSHViMRUwEwYDVQQLDAxNbG9wc0h1YiBEZXYxGjAYBgNVBAMM
EWRlbW8ubWxvcHNodWIuY29tMFYwEAYHKoZIzj0CAQYFK4EEAAoDQgAEqzH+rlXB
v0H3f2WsTeU0p679KzPu+7qSRnorTrXjjJo9srNgNPiaa/9C/tX/up4TsWZvpAtz
QXoRfFUJ9eWBe6OBjDCBiTBTBgNVHSMETDBKoT2kOzA5MRwwGgYDVQQDDBN3d3cu
cGVlcnNhZmUuY25DPVVTMRkwFwYDVQQHDBBCZWlqaW5nIENIQU9ZQU5HggkAxuWj
//yMtiEwCQYDVR0TBAIwADALBgNVHQ8EBAMCBPAwGgYDVR0RBBMwEYIPd3d3LnBl
ZXJzYWZlLmNuMAoGCCqGSM49BAMCA0cAMEQCIANla2bAh2xbtY8e3MnVLXWBhSss
OWRHBAXdwwh8oJxGAiB2VzALkw3au6buj0M6yIRzRaPuY0z4jfM2klrpG23jmA==
-----END CERTIFICATE-----`

	nodeK = `-----BEGIN PRIVATE KEY-----
MIGEAgEAMBAGByqGSM49AgEGBSuBBAAKBG0wawIBAQQgZ/BF7dM6KiTq6Xvf1OM7
Are3ZZ38mMxTICL6k0hgRLShRANCAASrMf6uVcG/Qfd/ZaxN5TSnrv0rM+77upJG
eitOteOMmj2ys2A0+Jpr/0L+1f+6nhOxZm+kC3NBehF8VQn15YF7
-----END PRIVATE KEY-----`
)

func TestChainsqlHandler_MakeDepositProposal(t *testing.T) {
	param := common2.MakeTxParam{}
	param.TxHash = common.UINT256_EMPTY.ToArray()
	param.Method = "test"
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	blk, _ := pem.Decode([]byte(nodeK))
	key, err := pkcs12.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	priv := key.(*ecdsa.PrivateKey)

	// the choise of hash according to the algo for signing
	// here is SHA256
	hasher := sm2.SHA256.New()
	val := sink.Bytes()
	hasher.Write(val)
	raw := hasher.Sum(nil)
	sig, err := priv.Sign(rand.Reader, raw, nil)
	if err != nil {
		t.Fatal(err)
	}

	caSet := &common3.CertTrustChain{
		Certs: make([]*sm2.Certificate, 2),
	}
	blk, _ = pem.Decode([]byte(agencyCA))
	caSet.Certs[0], err = sm2.ParseCertificate(blk.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	blk, _ = pem.Decode([]byte(nodeCA))
	caSet.Certs[1], _ = sm2.ParseCertificate(blk.Bytes)
	sink = common.NewZeroCopySink(nil)
	caSet.Serialization(sink)

	params := new(common2.EntranceParam)
	params.Proof = sig
	params.Extra = val
	params.SourceChainID = 6
	params.HeaderOrCrossChainMsg = sink.Bytes()
	sink = common.NewZeroCopySink(nil)
	params.Serialization(sink)

	ns := getNativeFunc(sink.Bytes(), nil, rootCA)
	h := NewChainsqlHandler()
	p, err := h.MakeDepositProposal(ns)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, p.Method, "test")
}
