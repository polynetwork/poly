/*
 * Copyright (C) 2021 The poly network Authors
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

package consensus_vote

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"reflect"
	"testing"

	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/poly/native/storage"
	"gotest.tools/assert"
)

var (
	acct1     *account.Account   = account.NewAccount("")
	acct2     *account.Account   = account.NewAccount("")
	acct3     *account.Account   = account.NewAccount("")
	acct4     *account.Account   = account.NewAccount("")
	acct5     *account.Account   = account.NewAccount("")
	acct6     *account.Account   = account.NewAccount("")
	acct7     *account.Account   = account.NewAccount("")
	acct11    *account.Account   = account.NewAccount("")
	acct12    *account.Account   = account.NewAccount("")
	acct13    *account.Account   = account.NewAccount("")
	acct14    *account.Account   = account.NewAccount("")
	acct15    *account.Account   = account.NewAccount("")
	acct16    *account.Account   = account.NewAccount("")
	acct17    *account.Account   = account.NewAccount("")
	acctList1 []*account.Account = []*account.Account{acct1, acct2, acct3, acct4, acct5, acct6, acct7}
	acctList2 []*account.Account = []*account.Account{acct11, acct12, acct13, acct14, acct15, acct16, acct17}
)

func Init(db *storage.CacheDB) {
	contractAddr, _ := hex.DecodeString("bA6F835ECAE18f5Fc5eBc074e5A0B94422a13126")
	side := &side_chain_manager.SideChain{
		Name:         "eth",
		ChainId:      2,
		BlocksToWait: 2,
		Router:       1,
		CCMCAddress:  contractAddr,
	}
	sink := common.NewZeroCopySink(nil)
	_ = side.Serialization(sink)
	db.Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(2)), cstates.GenRawStorageItem(sink.Bytes()))

	//put governance view
	var view uint32 = 1
	governanceView := &node_manager.GovernanceView{
		View: view, Height: 1, TxHash: common.Uint256{0},
	}
	contract := utils.NodeManagerContractAddress
	sink = common.NewZeroCopySink(nil)
	governanceView.Serialization(sink)
	db.Put(utils.ConcatKey(contract, []byte(node_manager.GOVERNANCE_VIEW)), cstates.GenRawStorageItem(sink.Bytes()))

	//put peer pool map
	peerPoolMap := &node_manager.PeerPoolMap{
		PeerPoolMap: make(map[string]*node_manager.PeerPoolItem),
	}
	for i, acct := range acctList1 {
		peerPoolMap.PeerPoolMap[hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey))] =
			&node_manager.PeerPoolItem{
				Index:      uint32(i + 1),
				PeerPubkey: hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey)),
				Address:    acct.Address,
				Status:     node_manager.ConsensusStatus,
			}
	}
	contract = utils.NodeManagerContractAddress
	viewBytes := utils.GetUint32Bytes(view)
	sink = common.NewZeroCopySink(nil)
	peerPoolMap.Serialization(sink)
	db.Put(utils.ConcatKey(contract, []byte(node_manager.PEER_POOL), viewBytes), cstates.GenRawStorageItem(sink.Bytes()))
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns, err := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)
	if err != nil {
		panic(fmt.Sprintf("NewNativeService error: %+v", err))
	}
	return ns
}

func TestNormalConsensusVote(t *testing.T) {
	store, _ := leveldbstore.NewMemLevelDBStore()
	db := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	Init(db)

	voteHandler := NewVoteHandler()
	{
		param := new(scom.EntranceParam)
		param.SourceChainID = 10
		param.Height = 20000
		makeTxParam := &scom.MakeTxParam{
			TxHash:              []byte{0x01, 0x02},
			CrossChainID:        []byte{0x01, 0x02},
			FromContractAddress: []byte{0x01, 0x02},
			ToChainID:           2,
			ToContractAddress:   []byte{0x01, 0x02},
			Method:              "lock",
			Args:                []byte{0x01, 0x02},
		}
		makeTxParamSink := common.NewZeroCopySink(nil)
		makeTxParam.Serialization(makeTxParamSink)

		for i := 0; i <= 3; i++ {
			param.RelayerAddress = acctList1[i].Address[:]
			param.Extra = makeTxParamSink.Bytes()
			sink := common.NewZeroCopySink(nil)
			param.Serialization(sink)

			tx := &types.Transaction{
				ChainID:    0,
				SignedAddr: []common.Address{acctList1[i].Address},
			}
			ns := NewNative(sink.Bytes(), tx, db)
			v, err := voteHandler.MakeDepositProposal(ns)
			assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
			assert.NilError(t, err, "test error")
		}
		//already signed sign
		param.RelayerAddress = acctList1[3].Address[:]
		param.Extra = makeTxParamSink.Bytes()
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acctList1[3].Address},
		}
		ns := NewNative(sink.Bytes(), tx, db)
		v, err := voteHandler.MakeDepositProposal(ns)
		assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
		assert.NilError(t, err, "test error")

		//quorum sign
		param.RelayerAddress = acctList1[4].Address[:]
		param.Extra = makeTxParamSink.Bytes()
		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx = &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acctList1[4].Address},
		}
		ns = NewNative(sink.Bytes(), tx, db)
		v, err = voteHandler.MakeDepositProposal(ns)
		assert.Equal(t, true, reflect.DeepEqual(makeTxParam, v), "makeTxParam is not correct")
		assert.NilError(t, err, "test error")

		//redundant sign
		param.RelayerAddress = acctList1[5].Address[:]
		param.Extra = makeTxParamSink.Bytes()
		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx = &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acctList1[5].Address},
		}
		ns = NewNative(sink.Bytes(), tx, db)
		v, err = voteHandler.MakeDepositProposal(ns)
		assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
		assert.NilError(t, err, "test error")
	}
}

func TestNodeChangeConsensusVote(t *testing.T) {
	store, _ := leveldbstore.NewMemLevelDBStore()
	db := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	Init(db)

	voteHandler := NewVoteHandler()
	{
		param := new(scom.EntranceParam)
		param.SourceChainID = 10
		param.Height = 20000
		makeTxParam := &scom.MakeTxParam{
			TxHash:              []byte{0x01, 0x02},
			CrossChainID:        []byte{0x01, 0x02},
			FromContractAddress: []byte{0x01, 0x02},
			ToChainID:           2,
			ToContractAddress:   []byte{0x01, 0x02},
			Method:              "lock",
			Args:                []byte{0x01, 0x02},
		}
		makeTxParamSink := common.NewZeroCopySink(nil)
		makeTxParam.Serialization(makeTxParamSink)

		for i := 0; i <= 3; i++ {
			param.RelayerAddress = acctList1[i].Address[:]
			param.Extra = makeTxParamSink.Bytes()
			sink := common.NewZeroCopySink(nil)
			param.Serialization(sink)

			tx := &types.Transaction{
				ChainID:    0,
				SignedAddr: []common.Address{acctList1[i].Address},
			}
			ns := NewNative(sink.Bytes(), tx, db)
			v, err := voteHandler.MakeDepositProposal(ns)
			assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
			assert.NilError(t, err, "test error")
		}

		//change consensus node
		var view uint32 = 1
		peerPoolMap := &node_manager.PeerPoolMap{
			PeerPoolMap: make(map[string]*node_manager.PeerPoolItem),
		}
		for i, acct := range acctList2 {
			peerPoolMap.PeerPoolMap[hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey))] =
				&node_manager.PeerPoolItem{
					Index:      uint32(i + 11),
					PeerPubkey: hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey)),
					Address:    acct.Address,
					Status:     node_manager.ConsensusStatus,
				}
		}
		contract := utils.NodeManagerContractAddress
		viewBytes := utils.GetUint32Bytes(view)
		sink := common.NewZeroCopySink(nil)
		peerPoolMap.Serialization(sink)
		db.Put(utils.ConcatKey(contract, []byte(node_manager.PEER_POOL), viewBytes), cstates.GenRawStorageItem(sink.Bytes()))

		//quorum sign
		param.RelayerAddress = acctList1[4].Address[:]
		param.Extra = makeTxParamSink.Bytes()
		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acctList1[4].Address},
		}
		ns := NewNative(sink.Bytes(), tx, db)
		v, err := voteHandler.MakeDepositProposal(ns)
		assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
		assert.NilError(t, err, "test error")

		for i := 0; i <= 3; i++ {
			param.RelayerAddress = acctList2[i].Address[:]
			param.Extra = makeTxParamSink.Bytes()
			sink := common.NewZeroCopySink(nil)
			param.Serialization(sink)

			tx := &types.Transaction{
				ChainID:    0,
				SignedAddr: []common.Address{acctList2[i].Address},
			}
			ns := NewNative(sink.Bytes(), tx, db)
			v, err := voteHandler.MakeDepositProposal(ns)
			assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
			assert.NilError(t, err, "test error")
		}
		//quorum sign
		param.RelayerAddress = acctList2[4].Address[:]
		param.Extra = makeTxParamSink.Bytes()
		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx = &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acctList2[4].Address},
		}
		ns = NewNative(sink.Bytes(), tx, db)
		v, err = voteHandler.MakeDepositProposal(ns)
		assert.Equal(t, true, reflect.DeepEqual(makeTxParam, v), "makeTxParam is not correct")
		assert.NilError(t, err, "test error")

		//redundant sign
		param.RelayerAddress = acctList2[5].Address[:]
		param.Extra = makeTxParamSink.Bytes()
		sink = common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx = &types.Transaction{
			ChainID:    0,
			SignedAddr: []common.Address{acctList2[5].Address},
		}
		ns = NewNative(sink.Bytes(), tx, db)
		v, err = voteHandler.MakeDepositProposal(ns)
		assert.Equal(t, (*scom.MakeTxParam)(nil), v, "makeTxParam not nil error")
		assert.NilError(t, err, "test error")
	}
}
