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
MIIBvjCCAWSgAwIBAgIUaL9yTsQ8JPW5aBqqO+BZf+sHU4gwCgYIKoZIzj0EAwIw
NTEOMAwGA1UEAwwFY2hhaW4xEzARBgNVBAoMCmZpc2NvLWJjb3MxDjAMBgNVBAsM
BWNoYWluMCAXDTIwMDkyMTA4MTU0MVoYDzIxMjAwODI4MDgxNTQxWjA1MQ4wDAYD
VQQDDAVjaGFpbjETMBEGA1UECgwKZmlzY28tYmNvczEOMAwGA1UECwwFY2hhaW4w
VjAQBgcqhkjOPQIBBgUrgQQACgNCAASJfBp5Kgh5KE188ecrrm6p9tnNZfidpYvy
zc0VvsQMBlRrYCzue202Bk8p2jQAnUjEh5VxDl6vS8ZH3LvgbzlAo1MwUTAdBgNV
HQ4EFgQUorN7Ml8AOhKYufvM02hGnR21QlswHwYDVR0jBBgwFoAUorN7Ml8AOhKY
ufvM02hGnR21QlswDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiAn
CMbUvn+M3dDGuK0vl0xKNZORBGpyID7DYzU96L7+owIhAIDpE+vK4Q6iqOvOebna
q6fx4KHTXocOimE0GXp6KDsw
-----END CERTIFICATE-----`

	agencyCA = `-----BEGIN CERTIFICATE-----
MIIBezCCASGgAwIBAgIUFcYPMOdZfLnS4OMiXIxb0kWuTtUwCgYIKoZIzj0EAwIw
NTEOMAwGA1UEAwwFY2hhaW4xEzARBgNVBAoMCmZpc2NvLWJjb3MxDjAMBgNVBAsM
BWNoYWluMB4XDTIwMDkyMTA4MTU0MVoXDTMwMDkxOTA4MTU0MVowNzEPMA0GA1UE
AwwGYWdlbmN5MRMwEQYDVQQKDApmaXNjby1iY29zMQ8wDQYDVQQLDAZhZ2VuY3kw
VjAQBgcqhkjOPQIBBgUrgQQACgNCAASM2BvX80G3bnhIwgmy6BJUW3mAEEhwWlZZ
OMHEUokYLxjcRm5zYDFSLm3qZxABxDZfbQ/CkWZNKr1EngQgygKsoxAwDjAMBgNV
HRMEBTADAQH/MAoGCCqGSM49BAMCA0gAMEUCIQD9UpNolGmJRnDFfB7YnT7gxqK4
bjTR1raaCTaWDtdmyAIgMuE9JU4zaSjCtrKT5hCmK/oogcble6CUcW2Mb6Zqk2g=
-----END CERTIFICATE-----`

	nodeCA = `-----BEGIN CERTIFICATE-----
MIIBhTCCASygAwIBAgIUcZdBSk8AarFW4NnCYTrfs9faSwEwCgYIKoZIzj0EAwIw
NzEPMA0GA1UEAwwGYWdlbmN5MRMwEQYDVQQKDApmaXNjby1iY29zMQ8wDQYDVQQL
DAZhZ2VuY3kwIBcNMjAwOTIxMDgxNTQxWhgPMjEyMDA4MjgwODE1NDFaMDQxDjAM
BgNVBAMMBW5vZGUwMRMwEQYDVQQKDApmaXNjby1iY29zMQ0wCwYDVQQLDARub2Rl
MFYwEAYHKoZIzj0CAQYFK4EEAAoDQgAEjcvtwNS9er6JESnlCL5WMep3fgFU8F4U
MHAZRuhp3LDOIiuEAqu3GUKto5B39gWoOObWVKDIcNRoKgIiCh65q6MaMBgwCQYD
VR0TBAIwADALBgNVHQ8EBAMCBeAwCgYIKoZIzj0EAwIDRwAwRAIgazCvifJM5qzW
DJq5c/7vsPY7O5koR1/IbaNjKHfEhhQCIDPEtj+cfgS2m8wL8xehEpTXaBTzTgbS
W2ImfYm1IQck
-----END CERTIFICATE-----`

	nodeK = `-----BEGIN PRIVATE KEY-----
MIGEAgEAMBAGByqGSM49AgEGBSuBBAAKBG0wawIBAQQgsPoI7jD9Pb+hp6F1yETj
hG57ZP6bTnK5hRoZ/gD/p5GhRANCAASNy+3A1L16vokRKeUIvlYx6nd+AVTwXhQw
cBlG6GncsM4iK4QCq7cZQq2jkHf2Bag45tZUoMhw1GgqAiIKHrmr
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
