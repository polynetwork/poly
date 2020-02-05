package relayer_manager

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

func TestRegisterRelayer(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = [][]byte{{1, 2, 4, 6}, {1, 4, 5, 7}, {1, 3, 5, 7, 9}}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nil)

	res, err := RegisterRelayer(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	res, err = RegisterRelayer(nativeService)
	assert.Equal(t, res, []byte{0})
	assert.NotNil(t, err)
}

func TestRemoveRelayer(t *testing.T) {
	params := new(RelayerListParam)
	params.AddressList = [][]byte{{1, 2, 4, 6}, {1, 4, 5, 7}}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	tx := &types.Transaction{
		SignedAddr: []common.Address{acct.Address},
	}
	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
	res, err := RemoveRelayer(nativeService)
	assert.Equal(t, res, []byte{1})
	assert.Nil(t, err)

	address1 := []byte{1, 3, 5, 7, 9}
	relayerRaw1, err := GetRelayerRaw(nativeService, address1)
	assert.Nil(t, err)
	assert.NotNil(t, relayerRaw1)

	address2 := []byte{1, 4, 5, 7}
	relayerRaw2, err := GetRelayerRaw(nativeService, address2)
	assert.Nil(t, err)
	assert.Nil(t, relayerRaw2)
}
