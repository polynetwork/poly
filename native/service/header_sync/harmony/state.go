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

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/harmony-one/harmony/block"
	"github.com/harmony-one/harmony/consensus/quorum"
	"github.com/harmony-one/harmony/crypto/bls"
	"github.com/harmony-one/harmony/shard"

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

func keyForHeaderHeight(chainID uint64) []byte {
	return utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID))
}


// Harmony config context
type Context struct {
	NetworkID quorum.NetworkID
	schedule quorum.Schedule
	networkType quorum.NetworkType
	chainConfig *quorum.ChainConfig
}

func (ctx *Context) Init() (err error){
	ctx.schedule, ctx.networkType, ctx.chainConfig, err = quorum.GetNetworkConfigAndShardSchedule(ctx.NetworkID)
	return
}

// Get last block in epoch
func (ctx *Context) EpochLastBlock(epoch uint64) uint64 {
	return ctx.schedule.EpochLastBlock(epoch)
}

// Check if block is the last one
func (ctx *Context) IsLastBlock(height uint64) bool {
	return ctx.schedule.IsLastBlock(height)
}

// Check if epoch is staking-enabled
func(ctx *Context) IsStaking(epoch *big.Int) bool {
	return ctx.chainConfig.IsStaking(epoch)
}

// Verify if an epoch is valid
func (ctx *Context) VerifyEpoch(epoch *Epoch ) (err error) {
	epochID := big.NewInt(int64(epoch.EpochID))
	// Check if epoch is skipped
	if ctx.schedule.IsSkippedEpoch(epoch.ShardID, epochID) {
		return fmt.Errorf("skipped epoch %d for shard %d", epoch.EpochID, epoch.ShardID)
	}
	return
}

// Verify new epoch has desired keys in slots
func (ctx *Context) VerifyNextEpoch(shardID int, epoch *Epoch) (err error) {
	epochID := big.NewInt(int64(epoch.EpochID))
	instance := ctx.schedule.InstanceForEpoch(epochID)
	num, err := quorum.CheckHarmonyAccountsInSlots(instance, shardID, epoch.Committee.Slots)
	if err != nil { return }
	maxStakingSlots := shard.ExternalSlotsAvailableForEpoch(epochID)
	if len(epoch.Committee.Slots) - num > maxStakingSlots {
		err = fmt.Errorf("stake nodes size(%d) bigger than available %d, current harmony nodes %d",
			len(epoch.Committee.Slots) - num, maxStakingSlots, num)
	}
	return
}

// Build harmony quorum verifier
func (ctx *Context) NewVerifier(committee *shard.Committee, epoch *big.Int, isStaking bool) (quorum.Verifier, error) {
	return quorum.NewVerifierWithConfig(ctx.networkType, ctx.schedule, committee, epoch, isStaking)
}

// Harmony Epoch
type Epoch struct {
	EpochID uint64
	ShardID uint32
	Committee *shard.Committee
	StartHeight uint64
}

// Verfiy next epoch
func (this *Epoch) ValidateNextEpoch(ctx *Context, header *HeaderWithSig) (err error){
	err = this.VerifyHeader(ctx, header.Header)
	if err != nil { return }
	if header.Header.Number().Uint64() != ctx.EpochLastBlock(this.EpochID) {
		err = fmt.Errorf("block(%s) to sync should be the latest one in current epoch %v, desired height %v",
			header.Header.Number(), this.EpochID, ctx.EpochLastBlock(this.EpochID))
		return
	}
	return
}

// Verify header with current epoch
func (this *Epoch) VerifyHeader(ctx *Context, header *block.Header) (err error) {
	if this.EpochID != header.Epoch().Uint64() {
		err = fmt.Errorf("epoch does not match, current %v, got %v", this.EpochID, header.Epoch().Uint64())
		return
	}
	if this.ShardID != header.ShardID() {
		err = fmt.Errorf("shard ID does not match, current %v, got %v", this.ShardID, header.ShardID())
	}
	height := header.Number().Uint64()
	if height < this.StartHeight || height > ctx.EpochLastBlock(this.EpochID) {
		err = fmt.Errorf("header height(%v) is not in range: %v to %v", height, this.StartHeight, ctx.EpochLastBlock(this.EpochID))
		return
	}
	return
}

// Verify harmony header signature
func (this *Epoch) VerifyHeaderSig (ctx *Context, header *HeaderWithSig) (err error) {
	isStaking := ctx.IsStaking(big.NewInt(int64(this.EpochID)))

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

	qrVerfier, err := ctx.NewVerifier(this.Committee, big.NewInt(int64(this.EpochID)), isStaking)
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
	epoch = &Epoch{}
	err = rlp.DecodeBytes(data, epoch)
	if err != nil {
		err = fmt.Errorf("%w, failed to rlp decode harmony epoch", err)
		epoch = nil
	}
	return
}

type HeaderPayload struct {
	HeaderRLP []byte
	Sig []byte
	Bitmap []byte
}

// Harmony Header with Signature
type HeaderWithSig struct {
	Header *block.Header
	HeaderRLP []byte
	Sig []byte
	Bitmap []byte
}

// Serialize header with sig
func EncodeHeaderWithSig(hs *HeaderWithSig) (data []byte, err error) {
	data, err = rlp.EncodeToBytes(&HeaderPayload{hs.HeaderRLP, hs.Sig, hs.Bitmap})
	if err != nil {
		err = fmt.Errorf("failed to encode harmony HeaderWithSig, err: %v", err)
	}
	return
}

// Deserialize
func DecodeHeaderWithSig(data []byte) (hs *HeaderWithSig, err error) {
	payload := new(HeaderPayload)
	err = rlp.DecodeBytes(data, payload)
	if err == nil {
		hs = &HeaderWithSig{HeaderRLP: payload.HeaderRLP, Sig: payload.Sig, Bitmap: payload.Bitmap}
		hs.Header = new(block.Header)
		err = rlp.DecodeBytes(hs.HeaderRLP, hs.Header)
	}
	if err != nil {
		hs = nil
		err = fmt.Errorf("failed to decode harmony HeaderWithSig, err: %v", err)
	}
	return
}

// Extract shard state for epoch info
func (hs *HeaderWithSig) ExtractEpoch() (epoch *Epoch, err error) {
	shardStateBytes := hs.Header.ShardState()
	if len(shardStateBytes) == 0 {
		err = fmt.Errorf("unexpected empty shard state in header")
		return
	}

	shardState, err := shard.DecodeWrapper(shardStateBytes)
	if err != nil {
		err = fmt.Errorf("%w, failed to decode header shard state", err)
		return
	}

	committee, err := shardState.FindCommitteeByID(hs.Header.ShardID())
	if err != nil {
		err = fmt.Errorf("%w, failed to find committee by shard id %v", err, hs.Header.ShardID())
		return
	}

	epoch = &Epoch{
		EpochID: hs.Header.Epoch().Uint64() + 1,
		ShardID: hs.Header.ShardID(),
		Committee: committee,
		StartHeight: hs.Header.Number().Uint64() + 1,
	}
	return
}

// Decode harmony context
func DecodeHarmonyContext(data []byte) (ctx *Context, err error) {
	ctx = new(Context)
	err = json.Unmarshal(data, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode harmony context, err: %v", err)
	}
	return
}

// Save epoch info into storage
func storeEpoch(native *native.NativeService, chainID uint64, epoch *Epoch) (err error) {
	bytes, err := EncodeEpoch(epoch)
	if err != nil { return }
	native.GetCacheDB().Put(keyForConsensus(chainID), cstates.GenRawStorageItem(bytes))
	return
}

// Get current harmony epoch
func GetEpoch(native *native.NativeService, chainID uint64) (epoch *Epoch, err error) {
	epochBytes, err := native.GetCacheDB().Get(keyForConsensus(chainID))
	if err != nil {
		err = fmt.Errorf("failed to get epoch info, err: %v", err)
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

// Save genesis header into storage
func storeGenesisHeader(native *native.NativeService, chainID uint64, header *HeaderWithSig) (err error) {
	headerBytes, err := EncodeHeaderWithSig(header)
	if err != nil {
		return fmt.Errorf("%w, failed to marshal header", err)
	}
	native.GetCacheDB().Put(keyForGenesisHeader(chainID), cstates.GenRawStorageItem(headerBytes))
	return
}

// Get genesis header from storage
func getGenesisHeader(native *native.NativeService, chainID uint64) (header *HeaderWithSig, err error) {
	headerBytes, err := native.GetCacheDB().Get(keyForGenesisHeader(chainID))
	if err != nil {
		err = fmt.Errorf("%w, failed to get genesis header", err)
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

	header, err = DecodeHeaderWithSig(bytes)
	if err != nil {
		err = fmt.Errorf("%w, failed to deserialize harmony header: %x", err, headerBytes)
	}
	return
}

// Update state, emit sync header event
func updateWithHeader(native *native.NativeService, chainID uint64, header *block.Header) (err error){
	height := header.Number().Uint64()
	native.GetCacheDB().Put(keyForHeaderHeight(chainID), cstates.GenRawStorageItem(utils.GetUint64Bytes(height)))

	scom.NotifyPutHeader(native, chainID, height, header.Hash().Hex())
	return
}
