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
package fisco

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
	common3 "github.com/polynetwork/poly/native/service/header_sync/common"
	fisco2 "github.com/polynetwork/poly/native/service/header_sync/fisco"
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
			root := &fisco2.FiscoRoot{
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

	// sm2P256 for pubkey
	rootCA = `-----BEGIN CERTIFICATE-----
MIIEFzCCAwGgAwIBAgIIb28QP9OdRXwwCwYJKoZIhvcNAQELMIGFMQswCQYDVQQG
EwJDTjEQMA4GA1UECAwHQmVpamluZzEQMA4GA1UEBwwHQmVpamluZzEMMAoGA1UE
CgwDQnNuMRAwDgYDVQQLDAdyc2Fyb290MRAwDgYDVQQLDAdic25iYXNlMQwwCgYD
VQQLDANjb20xEjAQBgNVBAMMCXJzYXJvb3RjYTAgFw0yMDA5MTgwOTE2NTdaGA8y
MTIwMDkxODA5MTY1N1owgYUxCzAJBgNVBAYTAkNOMRAwDgYDVQQIDAdCZWlqaW5n
MRAwDgYDVQQHDAdCZWlqaW5nMQwwCgYDVQQKDANCc24xEDAOBgNVBAsMB3JzYXJv
b3QxEDAOBgNVBAsMB2JzbmJhc2UxDDAKBgNVBAsMA2NvbTESMBAGA1UEAwwJcnNh
cm9vdGNhMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqPClB+uMUpX1
E7OlOteNaN/QSLKH8XUGB9t0TSFvLnRIpGWrAbseVTANjKcqrY5rcOdioWme6AXL
+2kUkD0JWXWc1QmHNG8gOnfK7NM7oZpXLW+lwI/Oe25gyuHuIwrhGB07sbZffKCU
VtY04uFICf1s5C/QcYvT7QAABCk8y2BDn+LaYbLsMDfoMJcJEjeYbkI+Cqnyc4Xt
WriUlQDo3E2T6XBRUC8arrU43WffqCyEY9qn7XzKyyacGPGd/jwsXkB5OiWKBWGl
ZotT9kWApGmpw+X5l8IUS32f+BzQNlW1/yMWWgl66WkSvSAYOrlhMmcDjN/SmN7x
bfT73TF97QIDAQABo4GKMIGHMBIGA1UdEwEB/wQIMAYBAf8CAQEwDgYDVR0PAQH/
BAQDAgEGMB0GA1UdDgQWBBSfRCi8cEGSzzmWv9uYb2pBHvvsDDAfBgNVHSMEGDAW
gBSfRCi8cEGSzzmWv9uYb2pBHvvsDDAhBgNVHREEGjAYghZjYS5yc2Fyb290LmJz
bmJhc2UuY29tMAsGCSqGSIb3DQEBCwOCAQEAPVOBfhrL89ZjUgJX6mLkRiAQDSJN
8hIWOGeeiE+upAFRIPJgr6aqz/HHA56qw9UgNEsW728Z1cUbchxjsyiwL04dU1Jz
9EjWrp3qAR3Qx5vPOvUMd2iqSgyccc26ep20Tan5jMOn7Sxzv9+CsmJKQVhI2dFr
O7IHLQbJN21CmF6Ffie9kaHxMMM4v3241QqSPFHCx3WFCWPmpl6XxZcDOQJmL3MN
qpoYAs9z71j/+bR+LTIdUTk3gyzWI6QzFJ5//1VfeuVWJYZcb+Oos25NtI8Gsspd
k4XUKm4pgIdDPAfdtpV2OAaO28E38yoqRGgxex5it87BKcV2SnFHFklLKw==
-----END CERTIFICATE-----
`

	agencyCA = `-----BEGIN CERTIFICATE-----
MIIDVTCCAj2gAwIBAgIJAKNA7Qsboii7MA0GCSqGSIb3DQEBCwUAMIGFMQswCQYD
VQQGEwJDTjEQMA4GA1UECAwHQmVpamluZzEQMA4GA1UEBwwHQmVpamluZzEMMAoG
A1UECgwDQnNuMRAwDgYDVQQLDAdyc2Fyb290MRAwDgYDVQQLDAdic25iYXNlMQww
CgYDVQQLDANjb20xEjAQBgNVBAMMCXJzYXJvb3RjYTAeFw0yMDA5MjEwODI1NTFa
Fw0zMDA5MTkwODI1NTFaMDwxFDASBgNVBAMMC3Rlc3RuZXRub2RlMRMwEQYDVQQK
DApmaXNjby1iY29zMQ8wDQYDVQQLDAZhZ2VuY3kwggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQCiWGlRpP0NKihqyu7lNXS+uG+cxF/vaAIrHaBI9iH6HmEE
6APKQ39YadusRXNaT96iHt5GuUHKFnFyG33j5/vVS92u4PwWHeVSMtU2WrGVxApS
aK+heZJxYohCmpcSym+gXodv+0XNRZRq/MQYHStdbWjuYgVLa3gjByyBiVkpqPt5
w5ceu8t3C652JHGip4Vf0SUZAPv8pTgWjHD5K3pESUkby2ur090vf8B0440dtfKg
o74m0aGGuPYrrb8vkgySgk5IQI7hJp9Vq/KyL3RdbS7Vbj3SF+AKRnCCnWaINQpd
iBdPwqrXH6a/nPQL4UVM85Wf4GyQmgmhWh2aK5QpAgMBAAGjEDAOMAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQELBQADggEBAAYGkXe5oLrw+CFMATFtdyg3ZWMNAOBy
Sww0YmWroAK1dJ2r4gk9GCPHiWhFCIGUCb4qPJoFyYLcBEL5sorWDy71bs3fEkZh
m+d2/xslIartZPeWMBPrfbKnti/wkQi9q+9hspJzFENmMN/9D1RgBoqVpCxFrxsH
APO8gVmVLkHEHI8iBPdsFRK1XVejH7dmEdJVkVFIzee608pSKC2GCFjSZhvcU/su
GzuN1yluba5YIupZ8GT2s/HbqDUBicvERLQ4xT1mCgHYDAF9Lb2FhehKsX5Papai
dsJ1yXHNNComRc2B9rRD+Zad/JnUAHEZyufoCoeCYcTf42R7e558XMU=
-----END CERTIFICATE-----`

	nodeCA = `-----BEGIN CERTIFICATE-----
MIICPjCCASagAwIBAgIJAKNA7Qsboii8MA0GCSqGSIb3DQEBCwUAMDwxFDASBgNV
BAMMC3Rlc3RuZXRub2RlMRMwEQYDVQQKDApmaXNjby1iY29zMQ8wDQYDVQQLDAZh
Z2VuY3kwIBcNMjAwOTIxMDgyNTUxWhgPMjEyMDA4MjgwODI1NTFaMDExDDAKBgNV
BAMMA3NkazETMBEGA1UECgwKZmlzY28tYmNvczEMMAoGA1UECwwDc2RrMFYwEAYH
KoZIzj0CAQYFK4EEAAoDQgAEfyhbyfvLBncLQhY8qjGBiylyOr+twOM0f3YoWLEm
AwWQPE6tNrgPzjFm7or/LiBt8FJHfkBhHSnkuZGbdtOj86MaMBgwCQYDVR0TBAIw
ADALBgNVHQ8EBAMCBeAwDQYJKoZIhvcNAQELBQADggEBAFytxOqjCLsL6VPi9jbR
K4NbTKgQ1SjfUjOogYkNdV/Y5WcOci6GmQWAftAVIsXN/30PVczm7dO331zTfvBc
iZiGUVGn9HRryNVAHW5F/OhnGAQ7dLY9ZSHmW9FXAK9FsGw3tSVp06PFYBWw7PwZ
DGe4kchwF4sZp7zn+pAMcI1LgIt6+BozNxCVOHER2jIAPi2V+is+EbNLkXAQDE/4
bJUcoeAldeAJuAAw+SZ4UpjWOcthvY0Kx8D7Tn9/q9fa6pi+w6xn7eJlAuQR2Tk1
6+mKiRnqMFfgGne6MzF7Ei6OegP5QzmtcpJVE94nj+fRNOwCthC173Tgq2R0eK39
DY0=
-----END CERTIFICATE-----`

	nodeK = `-----BEGIN PRIVATE KEY-----
MIGEAgEAMBAGByqGSM49AgEGBSuBBAAKBG0wawIBAQQgklUCTj3fWHQEIkAL3G0/
+Kqf3YF1Iauav00ES8RfKjWhRANCAAR/KFvJ+8sGdwtCFjyqMYGLKXI6v63A4zR/
dihYsSYDBZA8Tq02uA/OMWbuiv8uIG3wUkd+QGEdKeS5kZt206Pz
-----END PRIVATE KEY-----`

	// secp256k1 for pubkey
	rootCAGM = `-----BEGIN CERTIFICATE-----
MIIBxTCCAWqgAwIBAgIJALAiHhwloxDDMAoGCCqBHM9VAYN1MDcxEDAOBgNVBAMM
B2dtY2hhaW4xEzARBgNVBAoMCmZpc2NvLWJjb3MxDjAMBgNVBAsMBWNoYWluMCAX
DTIwMTAwOTE0MTk0OVoYDzIxMjAwOTE1MTQxOTQ5WjA3MRAwDgYDVQQDDAdnbWNo
YWluMRMwEQYDVQQKDApmaXNjby1iY29zMQ4wDAYDVQQLDAVjaGFpbjBZMBMGByqG
SM49AgEGCCqBHM9VAYItA0IABGuAAKUbkxLdUYhfenU+3u0Qr0thA6rcQ7wPjjHq
MQjKq7ET0NsVqU6gvbp6cvUd7HrYYu7GZbnyRyY/FLg8WOWjXTBbMB0GA1UdDgQW
BBQPWUqHyIJw+nuIY/n6BEyylGfWiTAfBgNVHSMEGDAWgBQPWUqHyIJw+nuIY/n6
BEyylGfWiTAMBgNVHRMEBTADAQH/MAsGA1UdDwQEAwIBBjAKBggqgRzPVQGDdQNJ
ADBGAiEAtKeE0E90QK2wKc9iYt9f53em+syFpyGJnm/jemMt3n0CIQC3dwivnZWX
1RvSpV/sTJwW2PNcheXPpy2dmxmNqwJr/A==
-----END CERTIFICATE-----`

	agencyCAGM = `-----BEGIN CERTIFICATE-----
MIIBxzCCAWygAwIBAgIJAOtUq4KceHzZMAoGCCqBHM9VAYN1MDcxEDAOBgNVBAMM
B2dtY2hhaW4xEzARBgNVBAoMCmZpc2NvLWJjb3MxDjAMBgNVBAsMBWNoYWluMB4X
DTIwMTAwOTE0MTk0OVoXDTMwMTAwNzE0MTk0OVowOzETMBEGA1UEAwwKYWdlbmN5
X3NvbjETMBEGA1UECgwKZmlzY28tYmNvczEPMA0GA1UECwwGYWdlbmN5MFkwEwYH
KoZIzj0CAQYIKoEcz1UBgi0DQgAE9QMCyGfOPX9MiIz36ch8f6VyHVqWJRG9a3/J
9VzkA5NGIvh1LVD0BmlPckHC74+RM8IRJCj7ui8PIH7I6xv0R6NdMFswHQYDVR0O
BBYEFJG3RCInvqrhJdSBoMex761aDq3aMB8GA1UdIwQYMBaAFA9ZSofIgnD6e4hj
+foETLKUZ9aJMAwGA1UdEwQFMAMBAf8wCwYDVR0PBAQDAgEGMAoGCCqBHM9VAYN1
A0kAMEYCIQC6Nz7a/c+XYUel+tczFriKb44OMO2xCKkTg4O8olEHdQIhAJ9ueoY+
hlZ1Rdlj5zp54i2VaZXiykOVIb17YxcbrdCg
-----END CERTIFICATE-----`

	nodeCAGM = `-----BEGIN CERTIFICATE-----
MIIBgTCCASigAwIBAgIJALuk6bFVg00+MAoGCCqBHM9VAYN1MDsxEzARBgNVBAMM
CmFnZW5jeV9zb24xEzARBgNVBAoMCmZpc2NvLWJjb3MxDzANBgNVBAsMBmFnZW5j
eTAgFw0yMDEwMDkxNDE5NDlaGA8yMTIwMDkxNTE0MTk0OVowNDEOMAwGA1UEAwwF
bm9kZTAxEzARBgNVBAoMCmZpc2NvLWJjb3MxDTALBgNVBAsMBG5vZGUwWTATBgcq
hkjOPQIBBggqgRzPVQGCLQNCAATO76uJQiWgZuM+NJBikLfpP8lL2xQUeYpcsE+g
dVzFv8H+wSnEZPvVBoAL6aYemQ95+4hc8wd1kyQOZSh4Pwg/oxowGDAJBgNVHRME
AjAAMAsGA1UdDwQEAwIGwDAKBggqgRzPVQGDdQNHADBEAiALFb1aWBHxEzyBbryh
RbxmUzuqam2J5RecvmvdlazhdgIgKEQ5Hl3bGkp19Pvnk3Npyy8RhDzY5Ex2cnVK
U8F8RPo=
-----END CERTIFICATE-----`

	nodeKGM = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBG0wawIBAQQgUKmXlaBz+QYcr/jb
sEkwCDOEobPTXXDCw8u6xJsbFd6hRANCAATO76uJQiWgZuM+NJBikLfpP8lL2xQU
eYpcsE+gdVzFv8H+wSnEZPvVBoAL6aYemQ95+4hc8wd1kyQOZSh4Pwg/
-----END PRIVATE KEY-----`
)

func TestFiscoHandler_MakeDepositProposal_GM(t *testing.T) {
	param := common2.MakeTxParam{}
	param.TxHash = common.UINT256_EMPTY.ToArray()
	param.Method = "test"
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	blk, _ := pem.Decode([]byte(nodeKGM))
	key, err := sm2.ParsePKCS8UnecryptedPrivateKey(blk.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	priv := key

	val := sink.Bytes()
	sig, err := priv.Sign(rand.Reader, val, nil)
	if err != nil {
		t.Fatal(err)
	}

	caSet := &common3.CertTrustChain{
		Certs: make([]*sm2.Certificate, 2),
	}
	blk, _ = pem.Decode([]byte(agencyCAGM))
	caSet.Certs[0], _ = sm2.ParseCertificate(blk.Bytes)
	blk, _ = pem.Decode([]byte(nodeCAGM))
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

	ns := getNativeFunc(sink.Bytes(), nil, rootCAGM)
	h := NewFiscoHandler()
	p, err := h.MakeDepositProposal(ns)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, p.Method, "test")
}

func TestFiscoHandler_MakeDepositProposal(t *testing.T) {
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
	h := NewFiscoHandler()
	p, err := h.MakeDepositProposal(ns)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, p.Method, "test")
}
