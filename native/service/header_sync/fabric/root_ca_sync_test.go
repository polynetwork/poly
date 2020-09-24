package fabric

import (
	"encoding/pem"
	"fmt"
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
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/tjfoc/gmsm/sm2"
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
MIICKDCCAc+gAwIBAgIRAN4EisCV7Y+rbW2hHV7wI0wwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMjAxMDExMTg1NzAwWhcNMzAxMDA5MTg1NzAw
WjBqMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzENMAsGA1UECxMEcGVlcjEfMB0GA1UEAxMWcGVlcjAub3Jn
MS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABIlsw55yk3JX
yqtkpCrUsFK5X5wwcfaB3F2SggaW5PPTC0QWx3qIXLlPCK67bnX4w8fpG3ECE2qI
W3dJ9pFiN0KjTTBLMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1Ud
IwQkMCKAIDdSh00xsy2nqjtFAK5YMYIrU5CrVLzVMJTuIqBRnftjMAoGCCqGSM49
BAMCA0cAMEQCIE6oFsTk+feM0FgPyzrAXz6X6T67Tx9t4EkZT/OoezD7AiBFElLQ
09lFFYvdtoQ/6rTc8TugxcWIlwgM4w6W9996+g==
-----END CERTIFICATE-----`

	orderCA = `-----BEGIN CERTIFICATE-----
MIICHjCCAcWgAwIBAgIRAKU15UAdRc3gZQuCCdYE2SIwCgYIKoZIzj0EAwIwaTEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xFDASBgNVBAoTC2V4YW1wbGUuY29tMRcwFQYDVQQDEw5jYS5leGFt
cGxlLmNvbTAeFw0yMDEwMDkwMjQ5MDBaFw0zMDEwMDcwMjQ5MDBaMGoxCzAJBgNV
BAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNp
c2NvMRAwDgYDVQQLEwdvcmRlcmVyMRwwGgYDVQQDExNvcmRlcmVyLmV4YW1wbGUu
Y29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEA67IcH48n8fpLoT9MjyDT6Qh
QZqGe5KXHG9sqHJdIbJoYpnHMxkletVrqI35Y6sgp4w9Sy+8jTvReHc1+fchwKNN
MEswDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwKwYDVR0jBCQwIoAgfi+u
kqWiPFOtT8mCFDWk2Rbl5JDHW1dwJRmcEyihyqkwCgYIKoZIzj0EAwIDRwAwRAIg
HNzfr04Jzi4J/p1UZn1U14JM8S6ym65/BxmH9uqepM8CIA5/tfv6aZ53PpOVYsrs
zQW7eQxTo228awU1AIwsA95+
-----END CERTIFICATE-----`
)

func TestFabricHandler_SyncGenesisHeader(t *testing.T) {
	blk, _ := pem.Decode([]byte(orderCA))
	cert, err := sm2.ParseCertificate(blk.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	cert1, err := sm2.ParseCertificate(blk.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	ss := make(map[sm2.Certificate]bool)
	ss[*cert] = true

	fmt.Println(ss[*cert1])
	fmt.Println(cert.SignatureAlgorithm)
}
