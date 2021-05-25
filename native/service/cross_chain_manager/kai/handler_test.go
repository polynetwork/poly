package kai

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
)

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns, err := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		panic(fmt.Errorf("NewNativeService error: %+v", err))
	}

	contaractAddr, _ := hex.DecodeString("48A77F43C0D7A6D6f588c4758dbA22bf6C5D95a0")
	side := &side_chain_manager.SideChain{
		Name:         "kai",
		ChainId:      138,
		BlocksToWait: 1,
		Router:       1,
		CCMCAddress:  contaractAddr,
	}
	sink := common.NewZeroCopySink(nil)
	_ = side.Serialization(sink)
	ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(2)), cstates.GenRawStorageItem(sink.Bytes()))
	return ns
}

func TestMakeDepositProposal(t *testing.T) {
	//handler := NewHandler()

}
