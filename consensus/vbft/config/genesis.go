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

package vconfig

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"time"

	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/common/log"
)

const (
	SCALE       = 15
	DEFAULT_POS = 100000
)

func shuffle_hash(height uint32, id string, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		Height uint32 `json:"height"`
		NodeID string `json:"node_id"`
		Index  int    `json:"index"`
	}{height, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}

func deepCopy(peersInfo []*config.VBFTPeerInfo) ([]*config.VBFTPeerInfo, error) {
	var peers []*config.VBFTPeerInfo
	buf, err := json.Marshal(peersInfo)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func genConsensusPayload(cfg *config.VBFTConfig, height uint32) ([]byte, error) {
	// deep copy to avoid modify global config
	peers, err := deepCopy(cfg.Peers)
	if err != nil {
		return nil, err
	}

	chainConfig, err := GenesisChainConfig(cfg, peers, height)
	if err != nil {
		return nil, err
	}

	// save VRF in genesis config file, to genesis block
	vrfValue, err := hex.DecodeString(cfg.VrfValue)
	if err != nil {
		return nil, fmt.Errorf("invalid config, vrf_value: %s", err)
	}

	vrfProof, err := hex.DecodeString(cfg.VrfProof)
	if err != nil {
		return nil, fmt.Errorf("invalid config, vrf_proof: %s", err)
	}

	// Notice:
	// take genesis msg as random source,
	// don't need verify (genesisProposer, vrfValue, vrfProof)

	vbftBlockInfo := &VbftBlockInfo{
		Proposer:           math.MaxUint32,
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: math.MaxUint32,
		NewChainConfig:     chainConfig,
	}
	return json.Marshal(vbftBlockInfo)
}

//GenesisChainConfig return chainconfig
func GenesisChainConfig(conf *config.VBFTConfig, peers []*config.VBFTPeerInfo, height uint32) (*ChainConfig, error) {
	log.Debugf("sorted peers: %v", peers)
	k := uint32(len(peers))
	sum := uint64(len(peers)) * DEFAULT_POS

	// calculate peer ranks
	scale := SCALE
	peerRanks := make([]uint64, 0)
	for i := 0; i < int(k); i++ {
		s := uint64(math.Ceil(float64(DEFAULT_POS) * float64(scale) * float64(k) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}

	log.Debugf("peers rank table: %v", peerRanks)

	// calculate pos table
	chainPeers := make(map[uint32]*PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(k); i++ {
		nodeId := peers[i].PeerPubkey
		chainPeers[peers[i].Index] = &PeerConfig{
			Index: peers[i].Index,
			ID:    nodeId,
		}
		for j := uint64(0); j < peerRanks[i]; j++ {
			posTable = append(posTable, peers[i].Index)
		}
	}
	// shuffle
	for i := len(posTable) - 1; i > 0; i-- {
		h, err := shuffle_hash(height, chainPeers[posTable[i]].ID, i)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash value: %s", err)
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}
	log.Debugf("init pos table: %v", posTable)

	// generate chain conf, and save to ChainConfigFile
	peerCfgs := make([]*PeerConfig, 0)
	for i := 0; i < int(k); i++ {
		peerCfgs = append(peerCfgs, chainPeers[peers[i].Index])
	}

	chainConfig := &ChainConfig{
		Version:              1,
		View:                 1,
		N:                    k,
		C:                    k / 3,
		BlockMsgDelay:        time.Duration(conf.BlockMsgDelay) * time.Millisecond,
		HashMsgDelay:         time.Duration(conf.HashMsgDelay) * time.Millisecond,
		PeerHandshakeTimeout: time.Duration(conf.PeerHandshakeTimeout) * time.Second,
		Peers:                peerCfgs,
		PosTable:             posTable,
		MaxBlockChangeView:   conf.MaxBlockChangeView,
	}
	return chainConfig, nil
}

func GenesisConsensusPayload(height uint32) ([]byte, error) {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)

	switch consensusType {
	case "vbft":
		return genConsensusPayload(config.DefConfig.Genesis.VBFT, height)
	}
	return nil, nil
}
