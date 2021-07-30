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
	"fmt"
	"sync"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/core/ledger"
	"github.com/polynetwork/poly/core/store"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/events/message"
)

type PendingBlock struct {
	block        *Block
	execResult   *store.ExecuteResult
	hasSubmitted bool
}
type ChainStore struct {
	mu *sync.Mutex

	db              *ledger.Ledger
	chainedBlockNum uint32
	pendingBlocks   map[uint32]*PendingBlock
	pid             *actor.PID
	needSubmitBlock bool
}

func OpenBlockStore(db *ledger.Ledger, serverPid *actor.PID) (*ChainStore, error) {
	chainstore := &ChainStore{
		db:              db,
		chainedBlockNum: db.GetCurrentBlockHeight(),
		pendingBlocks:   make(map[uint32]*PendingBlock),
		pid:             serverPid,
		needSubmitBlock: false,
		mu:              new(sync.Mutex),
	}
	merkleRoot, err := db.GetStateMerkleRoot(chainstore.chainedBlockNum)
	if err != nil {
		log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", chainstore.chainedBlockNum, err)
		return nil, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", chainstore.chainedBlockNum, err)
	}
	crossStatesRoot, err := db.GetCrossStateRoot(chainstore.chainedBlockNum)
	if err != nil {
		return nil, fmt.Errorf("GetCrossStatesRoot blockNum:%d, error :%s", chainstore.chainedBlockNum, err)
	}
	writeSet := overlaydb.NewMemDB(1, 1)
	block, err := chainstore.getBlock(chainstore.chainedBlockNum)
	if err != nil {
		return nil, err
	}
	chainstore.pendingBlocks[chainstore.chainedBlockNum] = &PendingBlock{block: block, execResult: &store.ExecuteResult{WriteSet: writeSet, MerkleRoot: merkleRoot, CrossStatesRoot: crossStatesRoot}}
	return chainstore, nil
}

func (self *ChainStore) close() {
	// TODO: any action on ledger actor??
}

func (self *ChainStore) GetChainedBlockNum() uint32 {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.chainedBlockNum
}

func (self *ChainStore) getExecMerkleRoot(blkNum uint32) (common.Uint256, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.MerkleRoot, nil
	}
	merkleRoot, err := self.db.GetStateMerkleRoot(blkNum)
	if err != nil {
		log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
		return common.Uint256{}, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return merkleRoot, nil
	}

}

func (self *ChainStore) getCrossStateRoot(blkNum uint32) (common.Uint256, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.CrossStatesRoot, nil
	}
	crossStateRoot, err := self.db.GetCrossStateRoot(blkNum)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("GetCrossStateRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return crossStateRoot, nil
	}
}

func (self *ChainStore) getExecWriteSet(blkNum uint32) *overlaydb.MemDB {
	self.mu.Lock()
	defer self.mu.Unlock()

	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.WriteSet
	}
	return nil
}

func (self *ChainStore) ReloadFromLedger() {
	self.mu.Lock()
	defer self.mu.Unlock()

	height := self.db.GetCurrentBlockHeight()
	if height > self.chainedBlockNum {
		// update chainstore height
		self.chainedBlockNum = height
		// remove persisted pending blocks
		newPending := make(map[uint32]*PendingBlock)
		for blkNum, blk := range self.pendingBlocks {
			if blkNum > height {
				newPending[blkNum] = blk
			}
		}
		// update pending blocks
		self.pendingBlocks = newPending
	}
}

func (self *ChainStore) AddBlock(block *Block) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if block == nil {
		return fmt.Errorf("try add nil block")
	}

	if block.getBlockNum() <= self.GetChainedBlockNum() {
		log.Warnf("chain store adding chained block(%d, %d)", block.getBlockNum(), self.GetChainedBlockNum())
		return nil
	}

	if block.Block.Header == nil {
		panic("nil block header")
	}
	blkNum := self.GetChainedBlockNum() + 1
	err := self.submitBlock(blkNum - 1)
	if err != nil {
		log.Errorf("chainstore blkNum:%d, SubmitBlock: %s", blkNum-1, err)
	}
	execResult, err := self.db.ExecuteBlock(block.Block)
	if err != nil {
		log.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
		return fmt.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
	}
	self.pendingBlocks[blkNum] = &PendingBlock{block: block, execResult: &execResult, hasSubmitted: false}
	if self.pid != nil {
		self.pid.Tell(
			&message.BlockConsensusComplete{
				Block: block.Block,
			})
	}
	self.chainedBlockNum = blkNum
	return nil
}

func (self *ChainStore) submitBlock(blkNum uint32) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if blkNum == 0 {
		return nil
	}
	if submitBlk, present := self.pendingBlocks[blkNum]; submitBlk != nil && submitBlk.hasSubmitted == false && present {
		err := self.db.SubmitBlock(submitBlk.block.Block, *submitBlk.execResult)
		if err != nil && blkNum > self.GetChainedBlockNum() {
			return fmt.Errorf("ledger add submitBlk (%d, %d) failed: %s", blkNum, self.GetChainedBlockNum(), err)
		}
		if _, present := self.pendingBlocks[blkNum-1]; present {
			delete(self.pendingBlocks, blkNum-1)
		}
		submitBlk.hasSubmitted = true
	}
	return nil
}

func (self *ChainStore) getBlock(blockNum uint32) (*Block, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk.block, nil
	}
	block, err := self.db.GetBlockByHeight(blockNum)
	if err != nil {
		return nil, err
	}
	return initVbftBlock(block)
}
