package side_chain_manager

import (
	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/storage"
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

func TestRegisterSideChainManager(t *testing.T) {
	param := new(RegisterSideChainParam)
	param.Address = ""
	param.BlocksToWait = 4
	param.ChainId = 8
	param.Name = "mychain"
	param.Router = 3

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)
	res, err := RegisterSideChain(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	sideChain, err := getSideChainApply(nativeService, 8)
	assert.Equal(t, sideChain.Name, "mychain")
	assert.Nil(t, err)

	res, err = RegisterSideChain(nativeService)
	assert.Equal(t, res, []byte{0})
	assert.NotNil(t, err)
}

func TestApproveRegisterSideChain(t *testing.T) {
	param := new(ChainidParam)
	param.Chainid = 8

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err := ApproveRegisterSideChain(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)
}

func TestUpdateSideChain(t *testing.T) {
	param := new(RegisterSideChainParam)
	param.Address = ""
	param.BlocksToWait = 10
	param.ChainId = 8
	param.Name = "own"
	param.Router = 3

	sink := common.NewZeroCopySink(nil)
	err := param.Serialization(sink)
	assert.Nil(t, err)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err := UpdateSideChain(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)
}

func TestApproveUpdateSideChain(t *testing.T) {
	param := new(ChainidParam)
	param.Chainid = 8
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}

	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err := ApproveUpdateSideChain(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	sideChain, err := GetSideChain(nativeService, 8)
	assert.Equal(t, sideChain.Name, "own")
	assert.Nil(t, err)
}

func TestRemoveSideChain(t *testing.T) {
	param := new(ChainidParam)
	param.Chainid = 8
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err := RemoveSideChain(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	sideChain, err := GetSideChain(nativeService, 8)
	assert.Nil(t, sideChain)
	assert.Nil(t, err)
}