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
	ca, _ := hex.DecodeString("9702640a6b971CA18EFC20AD73CA4e8bA390C910")
	redeem, _ := hex.DecodeString("552102dec9a415b6384ec0a9331d0cdf02020f0f1e5731c327b86e2b5a92455a289748210365b1066bcfa21987c3e207b92e309b95ca6bee5f1133cf04d6ed4ed265eafdbc21031104e387cd1a103c27fdc8a52d5c68dec25ddfb2f574fbdca405edfd8c5187de21031fdb4b44a9f20883aff505009ebc18702774c105cb04b1eecebcb294d404b1cb210387cda955196cc2b2fc0adbbbac1776f8de77b563c6d2a06a77d96457dc3d0d1f2102dd7767b6a7cc83693343ba721e0f5f4c7b4b8d85eeb7aec20d227625ec0f59d321034ad129efdab75061e8d4def08f5911495af2dae6d3e9a4b6e7aeb5186fa432fc57ae")
	sigStr := strings.Split("304402207cf1b8bf2d7234c77a84250a79d07a87b9fb09378096d34a5459b79afa414c57022015308108b6ec07df3b286c0fe20fe10b23e77377959d7160c339508ec1759da8,3045022100d6731dd8a0ee9e32423a25ed4638882d9ffcb259cdb03a3f75b8f1e3cd23540c02204d2511f9b748d5e356a9dfe20cfdda49a2de631637cc980ac7416cc7b6954466,3045022100a1e43664faafe50e429ad5c246266122dbc7df835f3758603c390a75019bb581022023d34b4c8bed500ea67ef5e9cbe259d7406487cc06a7da8d0661e9e58a0bbd52,3045022100e9716af38afd49fae2951c87ceec8add41d2915befee4962f3babb4e9b88897302207b870953ca1bde8edf1417ec7cb3e07d9c7862583aae92c9b8c408b746e14987,304402201bf226994026d060ddae579108bd5b1b06aeba4a313be6875b7ccf5482618ba602200ee71faa98c49f6d5120b6e8dc73663838be546cf47ca9c7a7c6bf2e672c8cfa", ",")
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
	param.Address = ""

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	ns := getNativeFunc(sink.Bytes())
	ok, err := RegisterRedeem(ns)
	assert.NoError(t, err)
	assert.Equal(t, utils.BYTE_TRUE, ok)
	states := ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, "c330431496364497d7257839737b5e4596f5ac06", hex.EncodeToString(states[1].([]byte)))
	assert.Equal(t, strings.ToLower("9702640a6b971CA18EFC20AD73CA4e8bA390C910"), states[2].(string))
}
