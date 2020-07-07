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

package node_manager

import (
	"fmt"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/utils"
)

func executeCommitDpos(native *native.NativeService) error {
	governanceView, err := GetGovernanceView(native)
	if err != nil {
		return fmt.Errorf("executeCommitDpos, get GovernanceView error: %v", err)
	}
	if native.GetHeight() == governanceView.Height {
		return fmt.Errorf("executeCommitDpos, can not do commitDpos twice in one block")
	}
	//get current view
	view := governanceView.View
	newView := view + 1

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, view)
	if err != nil {
		return fmt.Errorf("executeCommitDpos, get peerPoolMap error: %v", err)
	}

	for k, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			peerPoolMap.PeerPoolMap[k].Status = ConsensusStatus
		}
	}

	putPeerPoolMap(native, peerPoolMap, newView)
	oldView := view - 1
	oldViewBytes := utils.GetUint32Bytes(oldView)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.NodeManagerContractAddress, []byte(PEER_POOL), oldViewBytes))

	//update view
	governanceView = &GovernanceView{
		View:   view + 1,
		Height: native.GetHeight(),
		TxHash: native.GetTx().Hash(),
	}
	putGovernanceView(native, governanceView)
	return nil
}
