package node_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
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

	err = putPeerPoolMap(native, peerPoolMap, newView)
	if err != nil {
		return fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	oldView := view - 1
	oldViewBytes := utils.GetUint32Bytes(oldView)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.NodeManagerContractAddress, []byte(PEER_POOL), oldViewBytes))

	//update view
	governanceView = &GovernanceView{
		View:   view + 1,
		Height: native.GetHeight(),
		TxHash: native.GetTx().Hash(),
	}
	err = putGovernanceView(native, governanceView)
	if err != nil {
		return fmt.Errorf("putGovernanceView, put governanceView error: %v", err)
	}
	return nil
}
