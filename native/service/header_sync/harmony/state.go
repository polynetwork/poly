/*
 * Copyright (C) 2022 The poly network Authors
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

package harmony

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/harmony-one/harmony/block"
	"github.com/harmony-one/harmony/shard"
	"github.com/harmony-one/harmony/consensus/quorum"
	"github.com/harmony-one/harmony/crypto/bls"

	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

// Storage Keys
func keyForGenesisHeader(chainID uint64) []byte {
	return utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID))
}

func keyForConsensus(chainID uint64) []byte {
	return utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CONSENSUS_PEER), utils.GetUint64Bytes(chainID))
}

// Check if harmony staking enabled
func IsStaking(epoch uint64) (staking bool) {
	return
}

// Check if the block is the last one in an epoch
func IsLastEpochBlock(*big.Int) bool {
	return false
}

// Harmony Epoch
type Epoch struct {
	EpochID uint64
	ShardID uint32
	Committee *shard.Committee
	StartHeight uint64
}

// Verfiy next epoch
func (this *Epoch) ValidateNextEpoch(header *HeaderWithSig) (err error){
	return
}

// Verify harmony header signature
func (this *Epoch) VerifyHeaderSig (header *HeaderWithSig) (err error) {
	isStaking := IsStaking(this.EpochID)

	pubKeys, err := this.Committee.BLSPublicKeys()
	if err != nil {
		return fmt.Errorf("failed to get bls public keys, err: %v", err)
	}

	sigBytes := bls.SerializedSignature{}
	copy(sigBytes[:], header.Sig)

	// Decode bls aggregated sigantures
	aggSig, mask, err := DecodeSigBitmap(sigBytes, []byte(header.Bitmap), pubKeys)
	if err != nil {
		return fmt.Errorf("failed to decode bls sig bitmap, err: %v", err)
	}

	qrVerfier, err := quorum.NewVerifier(this.Committee, big.NewInt(int64(this.EpochID)), isStaking)
	if err != nil {
		return fmt.Errorf("failed to create quorum verifier, err: %v", err)
	}

	// Verify consensus signature
	if !qrVerfier.IsQuorumAchievedByMask(mask) {
		return fmt.Errorf("failed to check consensus quorum")
	}

	payload := ConstructCommitPayload(
		isStaking, header.Header.Hash(), header.Header.Number().Uint64(),
		header.Header.ViewID().Uint64())

	// Verify header hash
	if !aggSig.VerifyHash(mask.AggregatePublic, payload) {
		return fmt.Errorf("failed to verify header hash with consensus signature")
	}

	return
}

func EncodeEpoch(epoch *Epoch) (data []byte, err error) {
	data, err = rlp.EncodeToBytes(epoch)
	if err != nil {
		err = fmt.Errorf("%w, rlp encode harmony epoch error", err)
	}
	return
}

func DecodeEpoch(data []byte) (epoch *Epoch, err error) {
	bytes, err := cstates.GetValueFromRawStorageItem(data)
	if err != nil {
		err = fmt.Errorf("%w, failed to get harmony epoch value from raw storage item", err)
		return
	}
	epoch = &Epoch{}
	err = rlp.DecodeBytes(bytes, epoch)
	if err != nil {
		err = fmt.Errorf("%w, failed to rlp decode harmony epoch", err)
		epoch = nil
	}
	return
}

// Harmony Header with Signature
type HeaderWithSig struct {
	Header *block.Header
	Sig hexutil.Bytes
	Bitmap hexutil.Bytes
}

// Extract shard state for epoch info
func (hs *HeaderWithSig) ExtractEpoch() (epoch *Epoch, err error) {
	shardStateBytes := hs.Header.ShardState()
	if len(shardStateBytes) == 0 {
		err = fmt.Errorf("%w, HarmonyHandler unexpected empty shard state in header", err)
		return
	}

	shardState, err := shard.DecodeWrapper(shardStateBytes)
	if err != nil {
		err = fmt.Errorf("%w, HarmonyHandler failed to decode header shard state", err)
		return
	}

	committee, err := shardState.FindCommitteeByID(hs.Header.ShardID())
	if err != nil {
		err = fmt.Errorf("%w, HarmonyHandler failed to find committee by shard id %v", err, hs.Header.ShardID())
		return
	}

	epoch = &Epoch{
		EpochID: hs.Header.Epoch().Uint64(),
		ShardID: hs.Header.ShardID(),
		Committee: committee,
		StartHeight: hs.Header.Number().Uint64() + 1,
	}
	return
}


func storeEpoch(native *native.NativeService, chainID uint64, epoch *Epoch) (err error) {
	bytes, err := EncodeEpoch(epoch)
	if err != nil { return }
	native.GetCacheDB().Put(keyForConsensus(chainID), cstates.GenRawStorageItem(bytes))
	return
}

func getEpoch(native *native.NativeService, chainID uint64) (epoch *Epoch, err error) {
	epochBytes, err := native.GetCacheDB().Get(keyForConsensus(chainID))
	if err != nil {
		err = fmt.Errorf("%w HarmonyHandler failed to get epoch info", err)
		return
	}
	if epochBytes == nil {
		return
	}

	bytes, err := cstates.GetValueFromRawStorageItem(epochBytes)
	if err != nil {
		err = fmt.Errorf("%w, failed to get harmony epoch value from raw storage item", err)
		return
	}

	return DecodeEpoch(bytes)
}

func storeGenesisHeader(native *native.NativeService, chainID uint64, header *HeaderWithSig) (err error) {
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("%w HarmonyHandler failed to marshal header", err)
	}
	native.GetCacheDB().Put(keyForGenesisHeader(chainID), cstates.GenRawStorageItem(headerBytes))
	return
}

func getGenesisHeader(native *native.NativeService, chainID uint64) (header *HeaderWithSig, err error) {
	headerBytes, err := native.GetCacheDB().Get(keyForGenesisHeader(chainID))
	if err != nil {
		err = fmt.Errorf("%w HarmonyHandler failed to get genesis header", err)
		return
	}
	if headerBytes == nil {
		return
	}

	bytes, err := cstates.GetValueFromRawStorageItem(headerBytes)
	if err != nil {
		err = fmt.Errorf("%w, failed to get harmony genesis value from raw storage item", err)
		return
	}

	header = &HeaderWithSig{}
	err = json.Unmarshal(bytes, header)
	if err != nil {
		err = fmt.Errorf("%w, HarmonyHandler failed to deserialize harmony header: %x", err, headerBytes)
	}
	return
}