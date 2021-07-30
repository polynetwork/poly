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

package neo3

import (
	"encoding/hex"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/helper"
	tx2 "github.com/joeqian10/neo3-gogogo/tx"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/genesis"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	hscom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/neo3"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

var (
	acct          *account.Account = account.NewAccount("")
	getNativeFunc                  = func() *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service, _ := native.NewNativeService(cacheDB, new(types.Transaction), 0, 200, common.Uint256{}, 0, nil, false)
		return service
	}
	getNeoHanderFunc = func() *Neo3Handler {
		return NewNeo3Handler()
	}
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
)

func init() {
	setBKers()
}
func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns, _ := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	return ns
}
func SetContract(ns *native.NativeService, contractAddr string) {
	contaractAddr, _ := hex.DecodeString(contractAddr)
	side := &side_chain_manager.SideChain{
		Name:         "neo",
		ChainId:      4,
		BlocksToWait: 1,
		Router:       4,
		CCMCAddress:  contaractAddr,
	}
	sink := common.NewZeroCopySink(nil)
	_ = side.Serialization(sink)
	ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(side.ChainId)), cstates.GenRawStorageItem(sink.Bytes()))
}

func Test_Neo_MakeDepositProposal(t *testing.T) {
	var native *native.NativeService
	{
		prevHash, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
		merkleRoot, _ := helper.UInt256FromString("0x0000000000000000000000000000000000000000000000000000000000000000")
		nextConsensus, _ := crypto.AddressToScriptHash("NVg7LjGcUSrgxgjX3zEgqaksfMaiS8Z6e1", helper.DefaultAddressVersion)
		vs, _ := crypto.Base64Decode("EQ==")
		witness := tx2.Witness{
			InvocationScript: []byte{},
			VerificationScript: vs,
		}
		genesisHeader := &neo3.NeoBlockHeader{}
		genesisHeader.SetVersion(0)
		genesisHeader.SetPrevHash(prevHash)
		genesisHeader.SetMerkleRoot(merkleRoot)
		genesisHeader.SetTimeStamp(1468595301000)
		genesisHeader.SetIndex(0)
		genesisHeader.SetPrimaryIndex(0x00)
		genesisHeader.SetNextConsensus(nextConsensus)
		genesisHeader.SetWitnesses([]tx2.Witness{witness})

		param := new(hscom.SyncGenesisHeaderParam)
		param.ChainID = 4
		var err error
		sink := common.NewZeroCopySink(nil)
		err = genesisHeader.Serialization(sink)
		param.GenesisHeader = sink.Bytes()
		if err != nil {
			t.Errorf("NeoBlockHeaderToBytes error:%v", err)
		}

		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, nil)
		neoHandler := neo3.NewNeo3Handler()
		err = neoHandler.SyncGenesisHeader(native)
		assert.NoError(t, err)
	}

	{
		neoCrossChainMsgBs, _ := hex.DecodeString("004b0a01000f2adbe1d4c5022f34f7c4d1e5f033ea68102e716d66d70cdd194ae6dd7ed1b415f4c0ae08b37c51bd71a78b6aa5e79408f920a3b7b93e38d781985444b93b2b01c34059d48e180101adff905300e5e2074bfd12dee1d72e6a767cc00b384633fe64bfa2119811685c5062fc2ac947da15c08953dd603be791c41666f4ae759cb9ca8440b7201eb7d39052b22826a5445f4d7da896d337ffbec6dad0c08e413d5f1ec3f79caba079d77da11bc6ed800bccb9f689f34b4e85d7b4bb0dee14cbda75c5fedc40b7a409837de5e709f7e0b3e9f730f2475f6ddd8ba61d95849f16c97a78e9980879a740d8ab98bc012619b57170496fb7c6c515ae1d58eed50e9854667028acaf8b5321030f59a5482a4e42a2e5a848608dac4e84a698e567e2860e0ca5f23fc9e818d37c21032e78261370d4d62cf4c13584ca90f46c5565117b5b97544312f2e7b7c36b9eba21026e271722c21c482f0ac74dd932e61cdc2a2dd889633a2c5d8ecef43f2769f51e2103d55bfbcd493d06ab49c09cde0cea5d9ba890d81331a2fcd6f68d329932d0398f54ae")
		proofBs, _ := hex.DecodeString("25b0d4f20da68a6007d4fb7eac374b5566a5b0e229010202040000000000000000000000000c0bfd12010020f1bd34e03cf87844472bd97a245f6b4647f3da831101e8d35887b128447610cd0000205bb1e7f44a1a49702f37ec5b3c3ef254d9a55e84e30e91292e40b1a77b587c3520136df7bf0b604a15046281d14ff737f289fee06dcc21d2f5ebc3c879fe0b8f8100207f4fa0a7ca13b5712eae6df66b34c837134b1b3c206e35c8b08f58fd60eb95d9207dc275d32b07c66e4a7a85ca95812fe2bea2bc54205304db207789be9cdd9b440020ce58021310ac771901181d08fe4ea849214f5f8e8838325f88193589d06b5a6300209fc6b8cf16051f2cb0f16797064f3766ab90a0c7afe00a5ed205165a8ceb44da000020b235652da7f7892fd833b3f09656f3e74de8783896f85aa25a6315542d214d3700004a0127000d040f02000d0a06080a060000070d040f0b070e0a0c0307040b050506060a050b000e020209206fbaa1762862f38fcd56430a143632403b149491372122aed7a76252965008f85200206768e94d8cf974b35d8b20f36d4fd96a2a9ebdcb256c6cbccd67c741d8f3680500000020b532982b4fa5b52549439b9d6ab21dcdf1bdd08148573ce2cd8dbcfdc0647cd800000000000000000000000023010020584d7456e1a12be332ebe31e69970dfa18fa2bb7657c392fc63a44e01ed7f0dd92000020811d835143c81eb6d89295d33376ad7f2051fe9c12e97609373ab40983e3509d2075766bf9049fd5e20d14ffed1647f318518c125203de2dfc6473a03952cfc3e22068ea98f2f5f7d6e2c602885e57888adadd3e08e6c05043de7f1075fbfeb07f8f20c744e54af45aff995db84a5de396d73b916beed7cb54dcfadbde06e3845ada250000000000000000000000002401010020eff43eb246bba849c55908a9eca858194e43e89191555331dc2b6924621224875200002064022d7bf83e18daa3f0c521d74a430d4e9fca91edf56c484f249590378570bc20749a992c424e0da9d810b9954787b060190d631a83712f50134ac45bae43891f0000000000000000000000000000260103000200200d0c0166a089785a1f5dee523d1db836b6d30b6c226c5bd1e00493040e3b244b920000206f8cd767796594997c184b946d11a0305dc70c43bdc4ec142ec45ed1c10fd0b620409e5d22bbb60d53fb8043f823478024e7ebc89d635cd32efe27f0c301ea457e20cdf35a6da5556e432caba8db1559441a4ce2e79869340a964fda9637babf94d8209f4169310ee7c01aeacc88d3310f65cd2521d69da42a03c0ba56e1fe455e5cea0000000000000000000000003d011a000000000000000000000000000000000000000000000000000c201714f5861026fc3968ef442517da5bc15744770e2a48d55271ef40ed62aebbb1cb03c900c620dcb71d056f3cf1063bd47cd5466a0d0f480493bb963a1729f5b6119bbb2fb1002064084b8f53a425764bffd9c8d42ccef4550a6f298e92a57b339d9cbb37190751146591ea90e4cc4490f8bdd1a714f5f5d36a23711e020000000000000014a4260d6f81c436b8cbc99673eaf81e632c4e9d7d06756e6c6f636b4a14f8e41a74d1a9053acfc052df3686370452bb83c5140b24abdd39185055311aaa27082f9deb294a7255100000000000000000000000000000000000000000000000000000000000000000")
		param := &scom.EntranceParam{
			SourceChainID:         4,
			Height:                7006,
			Proof:                 proofBs,
			RelayerAddress:        acct.Address[:],
			Extra:                 []byte{},
			HeaderOrCrossChainMsg: neoCrossChainMsgBs,
		}

		var err error
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, native.GetCacheDB())
		neoHandler := NewNeo3Handler()
		SetContract(native, "b0d4f20da68a6007d4fb7eac374b5566a5b0e229")
		_, err = neoHandler.MakeDepositProposal(native)
		assert.Nil(t, err)
	}
}

func TestNEOHandler_SetCcmcId(t *testing.T) {
	// test positive int
	idp := 12
	idpBytes := helper.IntToBytes(idp)
	ss := helper.BytesToHex(idpBytes)
	log.Println(ss)
	assert.Equal(t, "0c000000", ss)

	idp2 := int(helper.BytesToUInt32(idpBytes))
	assert.Equal(t, idp, idp2)

	// test negative int
	idn := -5
	idnBytes := helper.IntToBytes(idn)
	log.Println(helper.BytesToHex(idnBytes))

	idn2 := int(int32(helper.BytesToUInt32(idnBytes)))
	assert.Equal(t, idn, idn2)
}
