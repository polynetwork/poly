package node_manager

import (
	"fmt"
	"sort"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
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

	//update config
	preConfig, err := getPreConfig(native)
	if err != nil {
		return fmt.Errorf("executeCommitDpos, get preConfig error: %v", err)
	}
	if preConfig.SetView == view {
		err = putConfig(native, preConfig.Configuration)
		if err != nil {
			return fmt.Errorf("executeCommitDpos, put config error: %v", err)
		}
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, view)
	if err != nil {
		return fmt.Errorf("executeCommitDpos, get peerPoolMap error: %v", err)
	}

	var peers []*PeerStakeInfo
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			pos := peerPoolItem.Pos + peerPoolItem.LockPos
			address, err := common.AddressParseFromBytes(peerPoolItem.Address)
			if err != nil {
				return fmt.Errorf("executeCommitDpos, common.AddressParseFromBytes error: %v", err)
			}
			//ont transfer
			err = appCallTransferOnt(native, utils.NodeManagerContractAddress, address, pos)
			if err != nil {
				return fmt.Errorf("executeCommitDpos normal, ont transfer error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			pos := peerPoolItem.Pos + peerPoolItem.LockPos
			// get operator from database
			operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
			if err != nil {
				return fmt.Errorf("executeCommitDpos black, types.AddressFromBookkeepers error: %v", err)
			}
			//ont transfer
			err = appCallTransferOnt(native, utils.NodeManagerContractAddress, operatorAddress, pos)
			if err != nil {
				return fmt.Errorf("executeCommitDpos black, ont transfer error: %v", err)
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			address, err := common.AddressParseFromBytes(peerPoolItem.Address)
			if err != nil {
				return fmt.Errorf("executeCommitDpos, common.AddressParseFromBytes error: %v", err)
			}
			//ont transfer
			err = appCallTransferOnt(native, utils.NodeManagerContractAddress, address, peerPoolItem.LockPos)
			if err != nil {
				return fmt.Errorf("executeCommitDpos normal, ont transfer error: %v", err)
			}
			peerPoolItem.LockPos = 0
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      peerPoolItem.Pos,
			})
		}
	}

	// get config
	config, err := GetConfig(native)
	if err != nil {
		return fmt.Errorf("executeCommitDpos, get config error: %v", err)
	}
	if len(peers) < int(config.K) {
		return fmt.Errorf("executeCommitDpos, num of peers is less than K")
	}
	// sort peers by stake
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("commitDpos, peerPubkey is not in peerPoolMap")
		}

		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return fmt.Errorf("authorizeForPeer, peerPubkey is not in peerPoolMap")
		}

		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
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
