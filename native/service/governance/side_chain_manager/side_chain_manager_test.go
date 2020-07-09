/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package side_chain_manager

import (
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/genesis"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
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
		service, _ := native.NewNativeService(cacheDB, new(types.Transaction), 0, 200, common.Uint256{}, 0, input, false)
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
	ns, err := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		panic("NewNativeService error")
	}
	return ns
}

func TestRegisterSideChainManager(t *testing.T) {
	param := new(RegisterSideChainParam)
	param.Address = acct.Address
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
	param.Address = common.Address{1, 2, 3}
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

//func TestRemoveSideChain(t *testing.T) {
//	param := new(ChainidParam)
//	param.Chainid = 8
//	sink := common.NewZeroCopySink(nil)
//	param.Serialization(sink)
//
//	tx := &types.Transaction{
//		SignedAddr: []common.Address{acct.Address},
//	}
//	nativeService = NewNative(sink.Bytes(), tx, nativeService.GetCacheDB())
//	res, err := RemoveSideChain(nativeService)
//	assert.Equal(t, res, []byte{1})
//	assert.Nil(t, err)
//
//	sideChain, err := GetSideChain(nativeService, 8)
//	assert.Nil(t, sideChain)
//	assert.Nil(t, err)
//}

func TestRegisterRedeem(t *testing.T) {
	ca, _ := hex.DecodeString("9a20bEd97360d28AE93c21750e9492ea8f85989f")
	redeem, _ := hex.DecodeString("552102dec9a415b6384ec0a9331d0cdf02020f0f1e5731c327b86e2b5a92455a289748210365b1066bcfa21987c3e207b92e309b95ca6bee5f1133cf04d6ed4ed265eafdbc21031104e387cd1a103c27fdc8a52d5c68dec25ddfb2f574fbdca405edfd8c5187de21031fdb4b44a9f20883aff505009ebc18702774c105cb04b1eecebcb294d404b1cb210387cda955196cc2b2fc0adbbbac1776f8de77b563c6d2a06a77d96457dc3d0d1f2102dd7767b6a7cc83693343ba721e0f5f4c7b4b8d85eeb7aec20d227625ec0f59d321034ad129efdab75061e8d4def08f5911495af2dae6d3e9a4b6e7aeb5186fa432fc57ae")
	sigStr := strings.Split("30450221009539bdeec25d289eaad308b0a79a1920b49784814b737499b9475c07e6ad82ab022051d435074e83b20b3c4d75f883a60ac1e21593763425055034e3f5670d0f5bf4,304402207c33e35b8dbf32d6b9ce52ef8ee9f60918ffb6ce2afe4f49d0e4e4f9cf031dc2022079756e511fe6971e13dabf91b1796dbc6aab1857c1f119fc74b34a1a413ba8c4,30440220794a0643bd5310ffda72aac380c4c7580a472fae7d583a08b041d67d2d0c68b602203ceed8d7e34885168c64acae1e323a96d19f0336f797a8fb88d5d29e414aecb9,304402203556fe453027af7814347cb2c5675685dbe8817224237759fc7a7c43271ef6ac02206f04ed4ab6963ccf57f5578381a38d695eadd58183fed1c357c7c291b75d98b9,3044022061b4a9152f7c128b672feebc04aad6341a286c889b137437d67526349b814b900220219427e0a57e26fb9413cc99813964a5156bf802028523191c919895733bc293", ",")
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

	ok, err = RegisterRedeem(ns)
	assert.Error(t, err)
	assert.Equal(t, utils.BYTE_FALSE, ok)
}

func TestSetBtcTxParam(t *testing.T) {
	redeem, _ := hex.DecodeString("552102dec9a415b6384ec0a9331d0cdf02020f0f1e5731c327b86e2b5a92455a289748210365b1066bcfa21987c3e207b92e309b95ca6bee5f1133cf04d6ed4ed265eafdbc21031104e387cd1a103c27fdc8a52d5c68dec25ddfb2f574fbdca405edfd8c5187de21031fdb4b44a9f20883aff505009ebc18702774c105cb04b1eecebcb294d404b1cb210387cda955196cc2b2fc0adbbbac1776f8de77b563c6d2a06a77d96457dc3d0d1f2102dd7767b6a7cc83693343ba721e0f5f4c7b4b8d85eeb7aec20d227625ec0f59d321034ad129efdab75061e8d4def08f5911495af2dae6d3e9a4b6e7aeb5186fa432fc57ae")
	sigStr := strings.Split("3045022100dbd452e851efbe8ae56c9a7da38d8ba59bf9fa5baefd439383271dba8998d4a00220227313c17e1438c5f679f10d520a5b7a1e56cbf6f0e6a824537a963ffb4d27f8,3045022100f22368985fbb00e3649b6d36e85b849c91dd57d3fc762fd63bbb7cdd478e3aeb022029aee48ed40473dcb9d1c634fe93b182b3ea6df400f09c00040339dabe67fd30,30450221009d3f736b27c991f78b84856c70e30813448d33284d04bb4dd0560b3bd250a15a02201e722d0087d537b2609e9140b398ce992b89ed583f2beb793ac78ded9bd0a207,3044022018f6cc301029843332794745b020dcb3fc768198c04026eb2efcbf0ba7febc0d02203c463b1b96eb6b7dcfff61bb6ccb87bffd6388d43ed464a9537e09b9c58c7e88,30440220301e5e1c37e699d8a0f999796b481a0611c11da7ea16389a760dccf5a8b659e8022072b4e0d48ef80db976bbd8f249a9b2aa0fb7fc995363a3ee9c629401cffe05b7", ",")
	sigs := make([][]byte, len(sigStr))
	for i, s := range sigStr {
		t, _ := hex.DecodeString(s)
		sigs[i] = t
	}
	param := new(BtcTxParam)
	param.Redeem = redeem
	param.Sigs = sigs
	param.RedeemChainId = 1
	param.Detial = &BtcTxParamDetial{
		PVersion:  0,
		MinChange: 2000,
		FeeRate:   2,
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	ns := getNativeFunc(sink.Bytes())
	ok, err := SetBtcTxParam(ns)
	assert.NoError(t, err)
	assert.Equal(t, utils.BYTE_TRUE, ok)
	states := ns.GetNotify()[0].States.([]interface{})
	assert.Equal(t, "c330431496364497d7257839737b5e4596f5ac06", states[1])
	assert.Equal(t, uint64(1), states[2])
	assert.Equal(t, param.Detial.FeeRate, states[3])
	assert.Equal(t, param.Detial.MinChange, states[4])

	ok, err = SetBtcTxParam(ns)
	assert.Error(t, err)
	assert.Equal(t, utils.BYTE_FALSE, ok)
}
