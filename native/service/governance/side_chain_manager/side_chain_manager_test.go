package side_chain_manager

import (
	"encoding/hex"
	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/ontio/multi-chain/native/storage"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var (
	acct          *account.Account = account.NewAccount("")
	getNativeFunc                  = func(input []byte) *native.NativeService {
		store, _ :=
			leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, input, false)
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

func TestRegisterRedeem(t *testing.T) {
	ca, _ := hex.DecodeString("9a20bEd97360d28AE93c21750e9492ea8f85989f")
	redeem, _ := hex.DecodeString("552102dec9a415b6384ec0a9331d0cdf02020f0f1e5731c327b86e2b5a92455a289748210365b1066bcfa21987c3e207b92e309b95ca6bee5f1133cf04d6ed4ed265eafdbc21031104e387cd1a103c27fdc8a52d5c68dec25ddfb2f574fbdca405edfd8c5187de21031fdb4b44a9f20883aff505009ebc18702774c105cb04b1eecebcb294d404b1cb210387cda955196cc2b2fc0adbbbac1776f8de77b563c6d2a06a77d96457dc3d0d1f2102dd7767b6a7cc83693343ba721e0f5f4c7b4b8d85eeb7aec20d227625ec0f59d321034ad129efdab75061e8d4def08f5911495af2dae6d3e9a4b6e7aeb5186fa432fc57ae")
	sigStr := strings.Split("3045022100c0803a32d2a83d342ebf5902a820ebefac44bf1d205645a322bd1046c2162b62022041bafcc24fbbc2d57f13c2853abadcd84636e7e61f707d8fa73dfb413e54a0ed,3045022100b5697f6d426b5cd60a6b66dc388fbe9caee6fb2d9c40e7a59ba4db98b9283aab022007dc301d0808ecc29a0ed60a688431fa9857bd0954fddafbb6ee3be3e74e331d,30450221009c86460a9448c1e98d8e56885561a3c6b28b0f047d165aaedf00e4115c5563c702202c6d188a1116e359ea8448a0ac052c885037553b2c33dca5e96da412b7c315a7,304402206ebc639a396a25f5e6a153d364f7e1224d2ee0d2e659ea3388619875d1b5b4250220347368d42a819eab23c82d41a4d6010ce7e1b69179e79e73499fe0f698a2e004,3045022100aee0b9ce6d361dccb79065e53de3c38f2f1df675e7f6b5d4edca3ed056abdc960220093dbafc7171441d6d743d4343c70c53ddf96941466c5fb0066aecb711a3454a", ",")
	sigs := make([][]byte, len(sigStr))
	for i, s := range sigStr {
		t, _ := hex.DecodeString(s)
		sigs[i] = t
	}

	param := new(RegisterRedeemParam)
	param.ContractAddress = ca
	param.ContractChainID = 2
	param.RedeemChainID = 1
	param.Signs = sigs
	param.Redeem = redeem
	param.CVersion = 0

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	ns := getNativeFunc(sink.Bytes())
	ok, err := RegisterRedeem(ns)
	assert.NoError(t, err)
	assert.Equal(t, utils.BYTE_TRUE, ok)
	states := ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, "c330431496364497d7257839737b5e4596f5ac06", states[1].(string))
	assert.Equal(t, strings.ToLower("9a20bEd97360d28AE93c21750e9492ea8f85989f"), states[2].(string))
}
