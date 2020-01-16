package btc

import (
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/storage"
)

var (
	getNativeFunc = func(args []byte, db *storage.CacheDB) *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		if db == nil {
			db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		}
		return native.NewNativeService(db, nil, 0, 0, common.Uint256{0}, 0, args, false, nil)
	}
)
