package polygon

import (
	"encoding/json"
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
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"gotest.tools/assert"
)

var (
	acct     = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
	heimdalChainID = uint64(2)
	borChainID     = uint64(3)
)

func init() {
	setBKers()
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) (service *native.NativeService, err error) {
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
	if db == nil {
		err = side_chain_manager.PutSideChain(service, &side_chain_manager.SideChain{
			ChainId: heimdalChainID,
		})
		if err != nil {
			return
		}
		extraInfo := ExtraInfo{
			Sprint:           64,
			Period:           2,
			ProducerDelay:    6,
			BackupMultiplier: 2,
			HeimdallChainID:  80001,
		}
		extraInfoBytes, _ := json.Marshal(extraInfo)
		err = side_chain_manager.PutSideChain(service, &side_chain_manager.SideChain{
			ChainId:   borChainID,
			ExtraInfo: extraInfoBytes,
		})
		if err != nil {
			return
		}
	}
	return
}

func TestPolygon(t *testing.T) {

	var (
		native *native.NativeService
		err    error
	)

	heimdallGenesisHeaderBytes := []byte("")
	handler := NewHeimdallHandler()

	{
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = heimdalChainID
		param.GenesisHeader = heimdallGenesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, nil)
		assert.NilError(t, err)
		err = handler.SyncGenesisHeader(native)
		if err != nil {
			t.Fatal("SyncGenesisHeader fail", err)
		}
	}

	{
		var n1Bytes []byte
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = heimdalChainID
		param.Address = acct.Address
		param.Headers = append(param.Headers, n1Bytes)
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		if err != nil {
			t.Fatal("NewNative fail", err)
		}
		err = handler.SyncBlockHeader(native)
		if err != nil {
			t.Fatal("SyncBlockHeader fail", err)
		}
	}

	borHandler := NewBorHandler()
	borGenesisHeaderBytes := []byte("")
	{
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = borChainID
		param.GenesisHeader = borGenesisHeaderBytes
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		assert.NilError(t, err)
		err = borHandler.SyncGenesisHeader(native)
		if err != nil {
			t.Fatal("SyncGenesisHeader fail", err)
		}
	}

	{
		var n1Bytes []byte
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = borChainID
		param.Address = acct.Address
		param.Headers = append(param.Headers, n1Bytes)
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)
		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native, err = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		if err != nil {
			t.Fatal("NewNative fail", err)
		}
		err = borHandler.SyncBlockHeader(native)
		if err != nil {
			t.Fatal("SyncBlockHeader fail", err)
		}
	}
}
