package harmony

import (
	"encoding/hex"
	"encoding/json"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"gotest.tools/assert"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/genesis"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
)

var (
	acct     = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
	harmonyChainID = uint64(2)
)

func init() {
	setBKers()
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) (service *native.NativeService, err error) {
	shouldInit := db == nil
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
	service, err = native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		return
	}
	if shouldInit {
		extraInfo := Context{
			NetworkID: 0, // Harmony Mainnet
		}
		extraInfoBytes, _ := json.Marshal(extraInfo)
		err = side_chain_manager.PutSideChain(service, &side_chain_manager.SideChain{
			ChainId:   harmonyChainID,
			ExtraInfo: extraInfoBytes,
		})
		if err != nil {
			return
		}
	}
	return
}

func TestHeaderSync(t *testing.T) {
	var (
		native *native.NativeService
		handler = NewHandler()
	)
	// Test genesis header
	{
		headerHex := "f905d9f9055487486d6e79546764827633f90546a0008ff03a0c79c1bd82be4591a48af64a424bda059f8610d54a4e5a48cffd615794f3c6adcbf25e4722826d8d350d6ea05e3b1015b9a05704ee532fddf54650cc5dd7833c02a12a7a29495198b02847b322a73eac9445a0f8341e59ba1ff1e17f009cc23c9f93103a8a3f99aa4eca4a98cfc1d3f2ae6b74a01905ee93688755235a2936498ce4ee4a4077d404b7dbbb75a885a7d196a35054a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b90100002148008000450921086002a100020800020400e40000101042008000000008002100200008800080020211000100001920000000000100040000008020000200000000064008602000000880200c300206004000010800000081001000800100000402b38040000040004402040820008100900408000000058010200008300000040003100010000020208023001100000000200800080b0050410000010282000000100000000001004202040000004000800000000001000000080200808002810200202002a000008008c020c00181800800000010000804012000202200108000000880a080012408200800040200001000000041000000200000000084016476b08404c4b40084010bacb184621756de80a000000000000000000000000000000000000000000000000000000000000000008401647e7382037580b8603cba685f42f5f45d79213488a4fb1580ad40e85ec683fe119d6ceb8d32617062156f76d919a97a26e6e74b733514bf10e7e3f3337ab696ba59787ed1e35b8320e15c93be77f6499352595bd196f7f2cd4cb4438aecf633b128f2e29aa66116989fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3fb880b99e4410deffbadcacd5f68382884b99716eab3f893d2b8bb5b3ffc5d5b1b5d1a12c46e6b75141f6fc62c7785f18ff8bc64382a93532f2326e6fa0049e53f82bd89ef4bd92991fa33f62ffc506056003299627c1e57be5f1b4d9959b47be1f14bfdd506905db38a66f54409fd54d1f459f900cfab1b6ec0adeaa28fea116d3888080b9021ff9021cf8b2a0106e4b4d687fe8f8f6fe22bb99b3b277f42a6d0067b6368b6858ff03e9da9e71840182e677840182f0fcb860a7717544e93125eed8ca4de85a129ba2d7c61d245e4f6d344d06a2e0cf1a31b841b443bdc33a8c722e78eead31bb4502f20dcaec24c1df686c311f458edf3bae3f22db9b6ae7b0c26f6436b8ab0f3e6fe0046e39e812abf6eda621320cd3b892a0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0f01820375f8b1a00eb691b60a46bf173c433b7547044d4b1724a18739091456aec7d3378f9a679f840184ffb0840185013db8605eb15dc0f01c02bb2922de582ed75e36a9e54adaeeeb0806ed9d04754c252f698e317955d0dea906d3ef8e2b9423c608a35207953e43fb7e34f3a89625b17f9ca055035f3ecd9e9b70437e4c9105716b61153f41e868e81489303d2711dc93099fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1f02820375f8b3a06fb87d4632cb0b6977ecef730bdd35ea37cd0e081e80314fcd55d058f9e0452e8401837ce3840183816cb8608ce66fc54523816c629c03760cdf0a0f54d49be2ecc4f2ca633d4fe1da469386f3f6e7f3eb9294640b1072b15ea095013b1909445a10fb3edf0529aee4ad7f2a8ef4a52f6870321a07dfee6aa7412be20585ff9d0a589705733025a08d66bf0ea1ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff010382037580b860b2a0ff8962f57499c633171db725f7c51c8f7f78ab4f4dad358f6aea960eac34855170dbefaf8df55a038da15d67450396f82f101325357ef9e00f56ad48fcb462f52489da51dd6a20293cc28da491f41a43ccfa0c9dab8789be5a9ffda360909fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3f"
		data, err := hex.DecodeString(headerHex)
		assert.NilError(t, err)
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = harmonyChainID
		param.GenesisHeader = data
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}
		native, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)
		err = handler.SyncGenesisHeader(native)
		assert.NilError(t, err)
	}
}

