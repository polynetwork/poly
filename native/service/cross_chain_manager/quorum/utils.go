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
package quorum

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	eth2 "github.com/polynetwork/poly/native/service/cross_chain_manager/eth"
	cmanager "github.com/polynetwork/poly/native/service/governance/side_chain_manager"
)

func verifyFromQuorumTx(proof, extra []byte, hdr *types.Header, sideChain *cmanager.SideChain) error {
	ethProof := new(eth2.ETHProof)
	if err := json.Unmarshal(proof, ethProof); err != nil {
		return fmt.Errorf("VerifyFromEthProof, unmarshal proof error:%s", err)
	}
	if len(ethProof.StorageProofs) != 1 {
		return fmt.Errorf("VerifyFromEthProof, incorrect proof format")
	}
	proofResult, err := eth2.VerifyMerkleProof(ethProof, hdr, sideChain.CCMCAddress)
	if err != nil {
		return fmt.Errorf("VerifyFromEthProof, verifyMerkleProof error:%v", err)
	}
	if proofResult == nil {
		return fmt.Errorf("VerifyFromEthProof, verifyMerkleProof failed!")
	}
	if !eth2.CheckProofResult(proofResult, extra) {
		return fmt.Errorf("VerifyFromEthProof, verify proof value hash failed, proof result:%x, extra:%x", proofResult, extra)
	}
	return nil
}