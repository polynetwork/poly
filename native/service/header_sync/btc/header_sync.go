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

package btc

import (
	"encoding/binary"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"math/big"

	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"

	"bytes"
	"github.com/btcsuite/btcd/wire"
	"github.com/polynetwork/poly/common/log"
)

type BTCHandler struct {
}

func NewBTCHandler() *BTCHandler {
	return &BTCHandler{}
}

func (this *BTCHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	header, height, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("BTCHandler SyncGenesisHeader: %s", err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)))
	if err != nil {
		return fmt.Errorf("BTCHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return fmt.Errorf("BTCHandler GetHeaderByHeight, genesis header had been initialized")
	}

	//block header storage
	storedHeader := StoredHeader{
		Header:    *header,
		Height:    height,
		totalWork: big.NewInt(0),
	}
	putGenesisBlockHeader(native, params.ChainID, storedHeader)

	return nil
}

func (this *BTCHandler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range headerParams.Headers {
		var blockHeader wire.BlockHeader
		err := blockHeader.Deserialize(bytes.NewBuffer(v))
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}

		_, err = GetHeaderByHash(native, headerParams.ChainID, blockHeader.BlockHash())
		if err == nil {
			continue
		}

		//isBestHeader, commonAncestor, heightOfHeader, err := commitHeader(native, headerParams.ChainID, blockHeader)
		_, _, _, err = commitHeader(native, headerParams.ChainID, blockHeader)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, commit header err: %v", err)
		}

	}
	return nil
}

func (this *BTCHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func getGenesisHeader(input []byte) (*wire.BlockHeader, uint32, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return nil, 0, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	l := len(params.GenesisHeader)
	if l != 84 {
		return nil, 0, fmt.Errorf("getGenesisHeader, wrong genesis header length %d", l)
	}
	header := new(wire.BlockHeader)
	if err := header.BtcDecode(bytes.NewBuffer(params.GenesisHeader[:l-4]),
		wire.ProtocolVersion, wire.LatestEncoding); err != nil {
		return nil, 0, fmt.Errorf("getGenesisHeader, deserialize wire.BlockHeader err: %v", err)
	}

	return header, binary.BigEndian.Uint32(params.GenesisHeader[l-4:]), nil
}

// the bool value indicates whether header is the newest
// the *StoredHeader indicates the common ancestor if header is not newest, and if there exists common ancestor
// height of the header, which we try to commit to the db
// error info
func commitHeader(native *native.NativeService, chainID uint64, header wire.BlockHeader) (bool, *StoredHeader, uint32, error) {
	newTip := false
	var commonAncestor *StoredHeader
	// Fetch our current best header from db
	bestHeader, err := GetBestBlockHeader(native, chainID)
	if err != nil {
		return false, nil, 0, err
	}
	tipHash := bestHeader.Header.BlockHash()
	parentHeader := new(StoredHeader)
	headerHash := header.BlockHash()

	// If the tip is also the parent of this header, then we can save a database read by skipping
	// the lookup of the parent header. Otherwise (ophan?) we need to fetch the parent.
	if header.PrevBlock.IsEqual(&tipHash) {
		parentHeader = bestHeader
	} else {
		parentHeader, err = GetPreviousHeader(native, chainID, header)
		if err != nil {
			return false, nil, 0, fmt.Errorf("commit header error: header %s is an orphan with details: %v",
				headerHash, err)
		}
	}
	valid, err := CheckHeader(native, chainID, header, parentHeader)
	if err != nil {
		return false, nil, 0, err
	}
	if !valid {
		return false, nil, 0, nil
	}
	// If this block is already the tip, return
	if tipHash.IsEqual(&headerHash) {
		return false, nil, 0, nil
	}

	// Add the work of this header to the total work stored at the previous header
	cumulativeWork := new(big.Int).Add(parentHeader.totalWork, blockchain.CalcWork(header.Bits))

	// If the cumulative work is greater than the total work of our best header
	// then we have a new best header. Update the chain tip and check for a reorg.
	var hdrsToUpdate []chainhash.Hash
	if cumulativeWork.Cmp(bestHeader.totalWork) == 1 {
		newTip = true
		prevHash := parentHeader.Header.BlockHash()
		// If this header is not extending the previous best header then we have a reorg.
		if !tipHash.IsEqual(&prevHash) {
			commonAncestor, hdrsToUpdate, err = GetCommonAncestor(native, chainID, &StoredHeader{
				Header: header,
				Height: parentHeader.Height + 1,
			}, bestHeader)
			if err != nil {
				return newTip, commonAncestor, 0, fmt.Errorf("Error calculating common ancestor: %s", err.Error())
			}
			log.Warnf("REORG! At block %d, Wiped out %d blocks", int(bestHeader.Height),
				int(bestHeader.Height-commonAncestor.Height))
		}
	}

	newHeight := parentHeader.Height + 1
	nb := StoredHeader{
		Header:    header,
		Height:    newHeight,
		totalWork: cumulativeWork,
	}
	// Put the header to the database
	//err = b.db.Put(nb, newTip)

	// whether newTip is false or true, update hash -> blockheader
	putBlockHeader(native, chainID, nb)

	if newTip {
		// update fixedkey -> bestblockheader
		putBestBlockHeader(native, chainID, nb)
		// update height -> blockhash
		putBlockHash(native, chainID, nb.Height, nb.Header.BlockHash())

	}
	if err != nil {
		return newTip, commonAncestor, 0, err
	}
	if commonAncestor != nil {
		err = ReIndexHeaderHeight(native, chainID, bestHeader.Height, hdrsToUpdate, &nb)
		if err != nil {
			return newTip, commonAncestor, 0, err
		}
	}

	return newTip, commonAncestor, newHeight, nil
}
