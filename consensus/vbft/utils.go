/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package vbft

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/vrf"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/ledger"
	"github.com/polynetwork/poly/core/signature"
	"github.com/polynetwork/poly/core/states"
	scommon "github.com/polynetwork/poly/core/store/common"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	nutils "github.com/polynetwork/poly/native/service/utils"
)

func SignMsg(account *account.Account, msg ConsensusMsg) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg when signing: %s", err)
	}

	return signature.Sign(account, data)
}

func hashData(data []byte) common.Uint256 {
	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return common.Uint256(f)
}

func HashMsg(msg ConsensusMsg) (common.Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := SerializeVbftMsg(msg)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	return hashData(data), nil
}

type seedData struct {
	BlockNum          uint32         `json:"block_num"`
	PrevBlockProposer uint32         `json:"prev_block_proposer"` // TODO: change to NodeID
	BlockRoot         common.Uint256 `json:"block_root"`
	VrfValue          []byte         `json:"vrf_value"`
}

func getParticipantSelectionSeed(block *Block) vconfig.VRFValue {

	data, err := json.Marshal(&seedData{
		BlockNum:          block.getBlockNum() + 1,
		PrevBlockProposer: block.getProposer(),
		BlockRoot:         block.Block.Header.BlockRoot,
		VrfValue:          block.getVrfValue(),
	})
	if err != nil {
		return vconfig.VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return vconfig.VRFValue(f)
}

type vrfData struct {
	BlockNum uint32 `json:"block_num"`
	PrevVrf  []byte `json:"prev_vrf"`
}

func computeVrf(sk keypair.PrivateKey, blkNum uint32, prevVrf []byte) ([]byte, []byte, error) {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("computeVrf failed to marshal vrfData: %s", err)
	}

	return vrf.Vrf(sk, data)
}

func verifyVrf(pk keypair.PublicKey, blkNum uint32, prevVrf, newVrf, proof []byte) error {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return fmt.Errorf("verifyVrf failed to marshal vrfData: %s", err)
	}

	result, err := vrf.Verify(pk, data, newVrf, proof)
	if err != nil {
		return fmt.Errorf("verifyVrf failed: %s", err)
	}
	if !result {
		return fmt.Errorf("verifyVrf failed")
	}
	return nil
}
func GetVbftConfigInfo(memdb *overlaydb.MemDB) (*config.VBFTConfig, error) {
	data, err := GetStorageValue(memdb, ledger.DefLedger, nutils.NodeManagerContractAddress, []byte(node_manager.VBFT_CONFIG))
	if err != nil {
		return nil, err
	}
	cfg := new(node_manager.Configuration)
	err = cfg.Deserialization(common.NewZeroCopySource(data))
	if err != nil {
		return nil, err
	}
	chainconfig := &config.VBFTConfig{
		BlockMsgDelay:        uint32(cfg.BlockMsgDelay),
		HashMsgDelay:         uint32(cfg.HashMsgDelay),
		PeerHandshakeTimeout: uint32(cfg.PeerHandshakeTimeout),
		MaxBlockChangeView:   uint32(cfg.MaxBlockChangeView),
	}

	return chainconfig, nil
}

func GetPeersConfig(memdb *overlaydb.MemDB) ([]*config.VBFTPeerInfo, error) {
	goveranceview, err := GetGovernanceView(memdb)
	if err != nil {
		return nil, err
	}
	viewBytes := nutils.GetUint32Bytes(goveranceview.View)
	key := append([]byte(node_manager.PEER_POOL), viewBytes...)
	data, err := GetStorageValue(memdb, ledger.DefLedger, nutils.NodeManagerContractAddress, key)
	if err != nil {
		return nil, err
	}
	peerMap := &node_manager.PeerPoolMap{
		PeerPoolMap: make(map[string]*node_manager.PeerPoolItem),
	}
	err = peerMap.Deserialization(common.NewZeroCopySource(data))
	if err != nil {
		return nil, err
	}
	var peerstakes []*config.VBFTPeerInfo
	for _, id := range peerMap.PeerPoolMap {
		if id.Status == node_manager.CandidateStatus || id.Status == node_manager.ConsensusStatus {
			config := &config.VBFTPeerInfo{
				Index:      uint32(id.Index),
				PeerPubkey: id.PeerPubkey,
			}
			peerstakes = append(peerstakes, config)
		}
	}
	return peerstakes, nil
}

func isUpdate(memdb *overlaydb.MemDB, view uint32) (bool, error) {
	goveranceview, err := GetGovernanceView(memdb)
	if err != nil {
		return false, err
	}
	if goveranceview.View > view {
		return true, nil
	}
	return false, nil
}

func getRawStorageItemFromMemDb(memdb *overlaydb.MemDB, addr common.Address, key []byte) (value []byte, unkown bool) {
	rawKey := make([]byte, 0, 1+common.ADDR_LEN+len(key))
	rawKey = append(rawKey, byte(scommon.ST_STORAGE))
	rawKey = append(rawKey, addr[:]...)
	rawKey = append(rawKey, key...)
	return memdb.Get(rawKey)
}

func GetStorageValue(memdb *overlaydb.MemDB, backend *ledger.Ledger, addr common.Address, key []byte) (value []byte, err error) {
	if memdb == nil {
		return backend.GetStorageItem(addr, key)
	}
	rawValue, unknown := getRawStorageItemFromMemDb(memdb, addr, key)
	if unknown {
		return backend.GetStorageItem(addr, key)
	}
	if len(rawValue) == 0 {
		return nil, scommon.ErrNotFound
	}

	value, err = states.GetValueFromRawStorageItem(rawValue)
	return
}

func GetGovernanceView(memdb *overlaydb.MemDB) (*node_manager.GovernanceView, error) {
	value, err := GetStorageValue(memdb, ledger.DefLedger, nutils.NodeManagerContractAddress, []byte(node_manager.GOVERNANCE_VIEW))
	if err != nil {
		return nil, err
	}
	governanceView := new(node_manager.GovernanceView)
	err = governanceView.Deserialization(common.NewZeroCopySource(value))
	if err != nil {
		return nil, err
	}
	return governanceView, nil
}

func getChainConfig(memdb *overlaydb.MemDB, blkNum uint32) (*vconfig.ChainConfig, error) {
	config, err := GetVbftConfigInfo(memdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get chainconfig from leveldb: %s", err)
	}

	peersinfo, err := GetPeersConfig(memdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get peersinfo from leveldb: %s", err)
	}
	goverview, err := GetGovernanceView(memdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get governanceview failed:%s", err)
	}

	cfg, err := vconfig.GenesisChainConfig(config, peersinfo, blkNum)
	if err != nil {
		return nil, fmt.Errorf("GenesisChainConfig failed: %s", err)
	}
	cfg.View = goverview.View
	return cfg, err
}
