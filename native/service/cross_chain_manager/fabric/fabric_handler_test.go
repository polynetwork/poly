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
package fabric

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
	common3 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	common2 "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/stretchr/testify/assert"
	"github.com/tjfoc/gmsm/pkcs12"
	"github.com/tjfoc/gmsm/sm2"
	"testing"
	"time"
)

var (
	acct = account.NewAccount("")

	getNativeFunc = func(args []byte, db *storage.CacheDB, certArr []string) *native.NativeService {
		signAddr, _ := types.AddressFromBookkeepers([]keypair.PublicKey{acct.PublicKey})
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

			ctc := &common2.CertTrustChain{
				Certs: make([]*sm2.Certificate, len(certArr)),
			}
			for i, cert := range certArr {
				blk, _ := pem.Decode([]byte(cert))
				ctc.Certs[i], _ = sm2.ParseCertificate(blk.Bytes)
			}
			sink = common.NewZeroCopySink(nil)
			ctc.Serialization(sink)
			db.Put(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(common2.MULTI_ROOT_CERT), utils.GetUint64Bytes(7)),
				states.GenRawStorageItem(sink.Bytes()))

			sc := &side_chain_manager.SideChain{
				BlocksToWait: uint64(side_chain_manager.JustOne),
				Address:      signAddr,
			}
			sink.Reset()
			_ = sc.Serialization(sink)
			db.Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(7)), states.GenRawStorageItem(sink.Bytes()))
		}

		ns, _ := native.NewNativeService(db, &types.Transaction{SignedAddr: []common.Address{signAddr}}, uint32(time.Now().Unix()), 0, common.Uint256{0}, 0, args, false)
		return ns
	}

	rootCAOrg1 = `-----BEGIN CERTIFICATE-----
MIICUTCCAfigAwIBAgIRANS0C96GioU5ecb1JTV/ObkwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMjAxMTA1MDcwMzAwWhcNMzAxMTAzMDcwMzAw
WjBzMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UE
AxMTY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BKoS/9aLE1mMtLOrSlt+DH9SU3J3efRw3NFlSRL1xvuFuZG/jt/dGWFvpkyWdGNg
Fa/qp0SrmsJ8gIvnUhQ19fSjbTBrMA4GA1UdDwEB/wQEAwIBpjAdBgNVHSUEFjAU
BggrBgEFBQcDAgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB/zApBgNVHQ4EIgQg
nQoGBKFOnc3Tqo8za8jmjtqdupaunMSFSJoSQH+3C3EwCgYIKoZIzj0EAwIDRwAw
RAIgcgN9GTvO946M7gpnhIcTXuzep01u61BVe9xexL7+YDcCIEjOGfqfzTFDP1aZ
Pou8TmZ2fkcbuYYSapwKDQ7nVmbj
-----END CERTIFICATE-----`

	rootCAOrg2 = `-----BEGIN CERTIFICATE-----
MIICUDCCAfegAwIBAgIQXQKsTgAHTf33l2JX63cTaDAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMi5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMi5leGFtcGxlLmNvbTAeFw0yMDExMDUwNzAzMDBaFw0zMDExMDMwNzAzMDBa
MHMxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMRkwFwYDVQQKExBvcmcyLmV4YW1wbGUuY29tMRwwGgYDVQQD
ExNjYS5vcmcyLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
yV6ghkj8U+MZf4SXN/D97mx6cVMaYYtVkdYAKQwy5nAwvUI1qYVhIOh0Os5siZlT
MtxCBLPiIwaVm/ixgg9hoKNtMGswDgYDVR0PAQH/BAQDAgGmMB0GA1UdJQQWMBQG
CCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdDgQiBCB0
+CjJvg2YGAdeUxXeQJQD4X3nOA2X+fLHjNEM4zj7kDAKBggqhkjOPQQDAgNHADBE
AiBJK6QJJb7nXmj9+oK8QcEx6Qp9yWmuK17ibl387xTOmQIgejb3xQn85uZwR6RA
oAtkjGvf3mBgb3Ur7KT8fLyvtjI=
-----END CERTIFICATE-----`

	admin = `-----BEGIN CERTIFICATE-----
MIIDQDCCAuegAwIBAgIUf58D9bMhbyjNZ8eigcQVnc1a8DcwCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMjAxMTE5MDgwMzAwWhcNMjExMTE5MDgw
ODAwWjCBmDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQw
EgYDVQQKEwtIeXBlcmxlZGdlcjE4MAoGA1UECxMDY29tMAsGA1UECxMEb3JnMTAN
BgNVBAsTBmNsaWVudDAOBgNVBAsTB2V4YW1wbGUxIDAeBgNVBAMMF0NsaWVudEBv
cmcxLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/BuMwWXr
veTLt3EKicgqJ/kCFQ/8tW98v6/wd7kIf99r0/UYsNxgnKcRVOAyQXg9pNmez2Gy
lSA3gbn8ybtS5aOCATEwggEtMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAA
MB0GA1UdDgQWBBT97zHQGJE4j5okk88gEIrnIjjeTTArBgNVHSMEJDAigCCdCgYE
oU6dzdOqjzNryOaO2p26lq6cxIVImhJAf7cLcTAZBgNVHREEEjAQgg5iZXN0d29y
ay5sb2NhbDCBpQYIKgMEBQYHCAEEgZh7ImF0dHJzIjp7ImNjbV9jYWxsZXIiOiJ0
cnVlIiwiaGYuQWZmaWxpYXRpb24iOiJvcmcxLmV4YW1wbGUuY29tIiwiaGYuRW5y
b2xsbWVudElEIjoiQ2xpZW50QG9yZzEuZXhhbXBsZS5jb20iLCJoZi5UeXBlIjoi
Y2xpZW50IiwicG9seV9yZWxheWVyIjoidHJ1ZSJ9fTAKBggqhkjOPQQDAgNHADBE
AiBH/DB0xFhkDbjUhs5N3EypdhCmr5wqJrobrVjqhZpMrgIgegAAmiETGCDU0GoT
1T40W+oFlt0TjB6rnCkl9AY5crU=
-----END CERTIFICATE-----
`

	privk = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgtTRFeyDyWcSG4Yro
b6xAiZI48pJQmU8VnHEnedzIXfWhRANCAAT8G4zBZeu95Mu3cQqJyCon+QIVD/y1
b3y/r/B3uQh/32vT9Riw3GCcpxFU4DJBeD2k2Z7PYbKVIDeBufzJu1Ll
-----END PRIVATE KEY-----`
)

func TestFabricHandler_MakeDepositProposal(t *testing.T) {
	param := common3.MakeTxParam{}
	param.TxHash = common.UINT256_EMPTY.ToArray()
	param.Method = "test"
	param.FromContractAddress = utils.CrossChainManagerContractAddress[:]
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	blk, _ := pem.Decode([]byte(privk))
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

	caSet := &common2.CertTrustChain{
		Certs: make([]*sm2.Certificate, 2),
	}

	blk, _ = pem.Decode([]byte(rootCAOrg1))
	caSet.Certs[0], _ = sm2.ParseCertificate(blk.Bytes)
	blk, _ = pem.Decode([]byte(admin))
	caSet.Certs[1], _ = sm2.ParseCertificate(blk.Bytes)

	mctc := common2.MultiCertTrustChain([]*common2.CertTrustChain{caSet})
	sink = common.NewZeroCopySink(nil)
	mctc.Serialization(sink)

	params := new(common3.EntranceParam)
	params.Proof = SigArrSerialize([][]byte{sig})
	params.Extra = val
	params.SourceChainID = 7
	params.HeaderOrCrossChainMsg = sink.Bytes()

	sink = common.NewZeroCopySink(nil)
	params.Serialization(sink)

	ns := getNativeFunc(sink.Bytes(), nil, []string{rootCAOrg1, rootCAOrg2})
	h := &FabricHandler{}
	p, err := h.MakeDepositProposal(ns)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, p.Method, "test")
}
