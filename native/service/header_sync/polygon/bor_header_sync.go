/*
 * Copyright (C) 2021 The poly network Authors
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
package polygon

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	polygonTypes "github.com/polynetwork/poly/native/service/header_sync/polygon/types"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/tendermint/tendermint/crypto/merkle"
	"golang.org/x/crypto/sha3"
)

// BorHandler ...
type BorHandler struct {
}

// NewHandler ...
func NewBorHandler() *BorHandler {
	return &BorHandler{}
}

// HeaderWithOptionalSnap ...
type HeaderWithOptionalSnap struct {
	Header   types.Header
	Snapshot *Snapshot
}

// SyncGenesisHeader ...
func (h *BorHandler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, checkWitness error: %v", err)
	}

	// can only store once
	stored, err := isGenesisStored(native, params)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, isGenesisStored error: %v", err)
	}
	if stored {
		return fmt.Errorf("bor Handler SyncGenesisHeader, genesis had been initialized")
	}

	var genesis HeaderWithOptionalSnap
	err = json.Unmarshal(params.GenesisHeader, &genesis)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, deserialize GenesisHeader err: %v", err)
	}

	if genesis.Snapshot == nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, genesis.Snapshot is nil")
	}
	side, err := side_chain_manager.GetSideChain(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, GetSideChain error: %v", err)
	}
	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, ExtraInfo Unmarshal error: %v", err)
	}

	err = storeGenesis(native, params, &genesis)
	if err != nil {
		return fmt.Errorf("bor Handler SyncGenesisHeader, storeGenesis error: %v", err)
	}

	return
}

func isGenesisStored(native *native.NativeService, params *scom.SyncGenesisHeaderParam) (stored bool, err error) {
	genesis, err := getGenesis(native, params.ChainID)
	if err != nil {
		return
	}

	stored = genesis != nil
	return
}

func getGenesis(native *native.NativeService, chainID uint64) (genesisHeader *HeaderWithOptionalSnap, err error) {

	genesisBytes, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)))
	if err != nil {
		err = fmt.Errorf("getGenesis, GetCacheDB err:%v", err)
		return
	}

	if genesisBytes == nil {
		return
	}

	genesisBytes, err = cstates.GetValueFromRawStorageItem(genesisBytes)
	if err != nil {
		err = fmt.Errorf("getGenesis, GetValueFromRawStorageItem err:%v", err)
		return
	}

	{
		genesisHeader = &HeaderWithOptionalSnap{}
		err = json.Unmarshal(genesisBytes, &genesisHeader)
		if err != nil {
			err = fmt.Errorf("getGenesis, json.Unmarshal err:%v", err)
			return
		}
	}

	return
}

func storeGenesis(native *native.NativeService, params *scom.SyncGenesisHeaderParam, genesisHeader *HeaderWithOptionalSnap) (err error) {

	genesisBytes, err := json.Marshal(genesisHeader)
	if err != nil {
		return
	}

	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)),
		cstates.GenRawStorageItem(genesisBytes))

	headerWithSum := &HeaderWithDifficultySum{HeaderWithOptionalSnap: genesisHeader, DifficultySum: genesisHeader.Header.Difficulty}

	err = putHeaderWithSum(native, params.ChainID, headerWithSum)
	if err != nil {
		return
	}

	putCanonicalHeight(native, params.ChainID, genesisHeader.Header.Number.Uint64())
	putCanonicalHash(native, params.ChainID, genesisHeader.Header.Number.Uint64(), genesisHeader.Header.Hash())

	scom.NotifyPutHeader(native, params.ChainID, genesisHeader.Header.Number.Uint64(), genesisHeader.Header.Hash().Hex())
	return
}

type ExtraInfo struct {
	Sprint              uint64
	Period              uint64
	ProducerDelay       uint64
	BackupMultiplier    uint64
	HeimdallPolyChainID uint64
}

type Context struct {
	ExtraInfo ExtraInfo
	ChainID   uint64
	Cdc       *codec.Codec
}

type HeaderWithDifficultySum struct {
	HeaderWithOptionalSnap *HeaderWithOptionalSnap `json:"headerWithOptionalSnap"`
	DifficultySum          *big.Int                `json:"difficultySum"`
	SnapParentHash         *ecommon.Hash           `json:"snapParentHash"`
}

type HeaderWithOptionalProof struct {
	Header types.Header
	Proof  []byte
}

// SyncBlockHeader ...
func (h *BorHandler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("bor Handler SyncBlockHeader, contract params deserialize error: %v", err)
	}

	side, err := side_chain_manager.GetSideChain(native, headerParams.ChainID)
	if err != nil {
		return fmt.Errorf("bor Handler SyncBlockHeader, GetSideChain error: %v", err)
	}
	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("bor Handler SyncBlockHeader, ExtraInfo Unmarshal error: %v", err)
	}

	ctx := &Context{ExtraInfo: extraInfo, ChainID: headerParams.ChainID, Cdc: polygonTypes.NewCDC()}

	for _, v := range headerParams.Headers {
		var headerWOP HeaderWithOptionalProof
		err := json.Unmarshal(v, &headerWOP)
		if err != nil {
			return fmt.Errorf("bor Handler SyncBlockHeader, deserialize header err: %v", err)
		}
		headerHash := headerWOP.Header.Hash()

		exist, err := isHeaderExist(native, headerHash, ctx)
		if err != nil {
			return fmt.Errorf("bor Handler SyncBlockHeader, isHeaderExist headerHash err: %v", err)
		}
		if exist {
			log.Warnf("bor Handler SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}

		parentExist, err := isHeaderExist(native, headerWOP.Header.ParentHash, ctx)
		if err != nil {
			return fmt.Errorf("bor Handler SyncBlockHeader, isHeaderExist ParentHash err: %v", err)
		}
		if !parentExist {
			log.Warnf("bor Handler SyncBlockHeader, parent header not exist. Header: %s", string(v))
			continue
		}

		var snap *Snapshot
		snap, err = verifyHeader(native, &headerWOP, ctx)
		if err != nil {
			return fmt.Errorf("bor Handler SyncBlockHeader, verifyHeader err: %v", err)
		}

		err = addHeader(native, &headerWOP.Header, snap, ctx)
		if err != nil {
			return fmt.Errorf("bor Handler SyncBlockHeader, addHeader err: %v", err)
		}

		scom.NotifyPutHeader(native, headerParams.ChainID, headerWOP.Header.Number.Uint64(), headerWOP.Header.Hash().Hex())
	}
	return nil
}

func isHeaderExist(native *native.NativeService, headerHash ecommon.Hash, ctx *Context) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(ctx.ChainID), headerHash.Bytes()))
	if err != nil {
		return false, fmt.Errorf("bor Handler isHeaderExist error: %v", err)
	}

	return headerStore != nil, nil
}

// GetCanonicalHeight ...
func GetCanonicalHeight(native *native.NativeService, chainID uint64) (height uint64, err error) {
	heightStore, err := native.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		err = fmt.Errorf("bor Handler GetCanonicalHeight err:%v", err)
		return
	}

	storeBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		err = fmt.Errorf("bor Handler GetCanonicalHeight, GetValueFromRawStorageItem err:%v", err)
		return
	}

	height = utils.GetBytesUint64(storeBytes)
	return
}

// GetCanonicalHeader ...
func GetCanonicalHeader(native *native.NativeService, chainID uint64, height uint64) (headerWithSum *HeaderWithDifficultySum, err error) {
	hash, err := getCanonicalHash(native, chainID, height)
	if err != nil {
		return
	}

	if hash == (ecommon.Hash{}) {
		return
	}

	headerWithSum, err = getHeader(native, hash, chainID)
	return
}

func deleteCanonicalHash(native *native.NativeService, chainID uint64, height uint64) {
	native.GetCacheDB().Delete(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
}

func getCanonicalHash(native *native.NativeService, chainID uint64, height uint64) (hash ecommon.Hash, err error) {
	hashBytesStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return
	}

	if hashBytesStore == nil {
		return
	}

	hashBytes, err := cstates.GetValueFromRawStorageItem(hashBytesStore)
	if err != nil {
		err = fmt.Errorf("bor Handler getCanonicalHash, GetValueFromRawStorageItem err:%v", err)
		return
	}

	hash = ecommon.BytesToHash(hashBytes)
	return
}

func putCanonicalHash(native *native.NativeService, chainID uint64, height uint64, hash ecommon.Hash) {
	native.GetCacheDB().Put(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)),
		cstates.GenRawStorageItem(hash.Bytes()))
}

func putHeaderWithSum(native *native.NativeService, chainID uint64, headerWithSum *HeaderWithDifficultySum) (err error) {

	headerBytes, err := json.Marshal(headerWithSum)
	if err != nil {
		return
	}

	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), headerWithSum.HeaderWithOptionalSnap.Header.Hash().Bytes()),
		cstates.GenRawStorageItem(headerBytes))
	return
}

func putCanonicalHeight(native *native.NativeService, chainID uint64, height uint64) {
	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(utils.GetUint64Bytes(uint64(height))))
}

func addHeader(native *native.NativeService, header *types.Header, snap *Snapshot, ctx *Context) (err error) {

	parentHeader, err := getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	cheight, err := GetCanonicalHeight(native, ctx.ChainID)
	if err != nil {
		return
	}
	cheader, err := GetCanonicalHeader(native, ctx.ChainID, cheight)
	if err != nil {
		return
	}
	if cheader == nil {
		err = fmt.Errorf("getCanonicalHeader returns nil")
		return
	}

	localTd := cheader.DifficultySum
	externTd := new(big.Int).Add(header.Difficulty, parentHeader.DifficultySum)

	headerWithSum := &HeaderWithDifficultySum{HeaderWithOptionalSnap: &HeaderWithOptionalSnap{Header: *header}, DifficultySum: externTd}
	if snap.Hash == header.Hash() {
		headerWithSum.HeaderWithOptionalSnap.Snapshot = snap
	} else {
		headerWithSum.SnapParentHash = &snap.Hash
	}
	err = putHeaderWithSum(native, ctx.ChainID, headerWithSum)
	if err != nil {
		return
	}

	if externTd.Cmp(localTd) > 0 {
		// Delete any canonical number assignments above the new head
		var headerWithSum *HeaderWithDifficultySum
		for i := header.Number.Uint64() + 1; ; i++ {
			headerWithSum, err = GetCanonicalHeader(native, ctx.ChainID, i)
			if err != nil {
				return
			}
			if headerWithSum == nil {
				break
			}

			deleteCanonicalHash(native, ctx.ChainID, i)
		}

		// Overwrite any stale canonical number assignments
		var (
			hash       ecommon.Hash
			headHeader *HeaderWithDifficultySum
		)
		cheight := header.Number.Uint64() - 1
		headHash := header.ParentHash

		for {
			hash, err = getCanonicalHash(native, ctx.ChainID, cheight)
			if err != nil {
				return
			}
			if hash == headHash {
				break
			}

			putCanonicalHash(native, ctx.ChainID, cheight, headHash)
			headHeader, err = getHeader(native, headHash, ctx.ChainID)
			if err != nil {
				return
			}
			headHash = headHeader.HeaderWithOptionalSnap.Header.ParentHash
			cheight--
		}

		// Extend the canonical chain with the new header
		putCanonicalHash(native, ctx.ChainID, header.Number.Uint64(), header.Hash())
		putCanonicalHeight(native, ctx.ChainID, header.Number.Uint64())
	}

	return nil
}

func getHeader(native *native.NativeService, hash ecommon.Hash, chainID uint64) (headerWithSum *HeaderWithDifficultySum, err error) {

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("bor Handler getHeader error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("bor Handler getHeader, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("bor Handler getHeader, deserialize headerBytes from raw storage item err:%v", err)
	}
	headerWithSum = &HeaderWithDifficultySum{}
	if err := json.Unmarshal(storeBytes, &headerWithSum); err != nil {
		return nil, fmt.Errorf("bor Handler getHeader, deserialize header error: %v", err)
	}

	return
}

var (
	extraVanity = 32                       // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = crypto.SignatureLength   // Fixed number of extra-data suffix bytes reserved for signer seal
	uncleHash   = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	validatorHeaderBytesLength = ecommon.AddressLength + 20 // address + power

)

func verifyHeader(native *native.NativeService, headerWOP *HeaderWithOptionalProof, ctx *Context) (snap *Snapshot, err error) {
	header := &headerWOP.Header
	if header.Number == nil {
		err = fmt.Errorf("errUnknownBlock")
		return
	}
	number := header.Number.Uint64()
	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		err = errors.New("block in the future")
		return
	}

	// check extr adata
	isSprintEnd := (number+1)%ctx.ExtraInfo.Sprint == 0

	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	if !isSprintEnd && signersBytes != 0 {
		err = errors.New("errExtraValidators")
		return
	}
	if isSprintEnd && signersBytes%validatorHeaderBytesLength != 0 {
		err = errors.New("errInvalidSpanValidators")
		return
	}
	if isSprintEnd {
		if err = validateHeaderExtraField(native, headerWOP, ctx); err != nil {
			return
		}
	} else {
		if headerWOP.Proof != nil {
			err = fmt.Errorf("Proof should be nil for non sprint end")
			return
		}
	}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (ecommon.Hash{}) {
		err = errors.New("non-zero mix digest")
		return
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.UncleHash != uncleHash {
		err = errors.New("non empty uncle hash")
		return
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if number > 0 {
		if header.Difficulty == nil {
			err = errors.New("errInvalidDifficulty")
			return
		}
	}

	// All basic checks passed, verify cascading fields
	return verifyCascadingFields(native, header, ctx)
}

type CosmosProof struct {
	Value  CosmosProofValue
	Proof  merkle.Proof
	Header CosmosHeader
}

type CosmosProofValue struct {
	Kp    string
	Value []byte
}

func putSpan(native *native.NativeService, ctx *Context, span *Span) (err error) {

	spanBytes, err := ctx.Cdc.MarshalBinaryBare(span)
	if err != nil {
		err = fmt.Errorf("putSpan MarshalBinaryBare failed:%v", err)
		return
	}
	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.POLYGON_SPAN), utils.GetUint64Bytes(ctx.ChainID)),
		cstates.GenRawStorageItem(spanBytes))
	return
}

func getSpan(native *native.NativeService, ctx *Context) (span *Span, err error) {
	spanBytes, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.POLYGON_SPAN), utils.GetUint64Bytes(ctx.ChainID)))
	if err != nil {
		err = fmt.Errorf("getSpan failed:%v", err)
		return
	}

	if spanBytes == nil {
		err = fmt.Errorf("getSpan:no span")
		return
	}

	spanBytes, err = cstates.GetValueFromRawStorageItem(spanBytes)
	if err != nil {
		err = fmt.Errorf("getSpan, GetValueFromRawStorageItem err:%v", err)
		return
	}

	span = &Span{}
	err = ctx.Cdc.UnmarshalBinaryBare(spanBytes, span)
	if err != nil {
		err = fmt.Errorf("getSpan, UnmarshalBinaryBare err:%v", err)
		return
	}
	return
}

func validateHeaderExtraFieldWithSpan(native *native.NativeService, headerWOP *HeaderWithOptionalProof, ctx *Context, span *Span) (err error) {
	if span == nil {
		err = fmt.Errorf("empty span")
		return
	}

	height := headerWOP.Header.Number.Uint64()
	if !(span.StartBlock <= (height+1) && span.EndBlock >= (height+1)) {
		err = fmt.Errorf("span not correct, span.StartBlock:%d, span.EndBlock:%d, height:%d", span.StartBlock, span.EndBlock, height)
		return
	}
	var newValidators []*Validator
	for i := range span.SelectedProducers {
		newValidators = append(newValidators, &span.SelectedProducers[i])
	}
	sort.Sort(ValidatorsByAddress(newValidators))
	var extra []byte
	for _, val := range newValidators {
		extra = append(extra, val.HeaderBytes()...)
	}

	if !bytes.Equal(extra, headerWOP.Header.Extra[extraVanity:len(headerWOP.Header.Extra)-extraSeal]) {
		return fmt.Errorf("invalid validators for sprint end, expect:%s got:%s", hex.EncodeToString(extra), hex.EncodeToString(headerWOP.Header.Extra[extraVanity:len(headerWOP.Header.Extra)-extraSeal]))
	}
	return nil
}

func validateHeaderExtraField(native *native.NativeService, headerWOP *HeaderWithOptionalProof, ctx *Context) (err error) {
	if headerWOP.Proof == nil {
		var span *Span
		span, err = getSpan(native, ctx)
		if err != nil {
			return
		}

		return validateHeaderExtraFieldWithSpan(native, headerWOP, ctx, span)
	}
	cdc := polygonTypes.NewCDC()
	var proof CosmosProof
	if err = cdc.UnmarshalBinaryBare(headerWOP.Proof, &proof); err != nil {
		return fmt.Errorf("validateHeaderExtraField, unmarshal CosmosProof err: %v", err)
	}

	span, err := VerifySpan(native, ctx.ExtraInfo.HeimdallPolyChainID, &proof)
	if err != nil {
		return fmt.Errorf("VerifySpan err: %v", err)
	}

	err = validateHeaderExtraFieldWithSpan(native, headerWOP, ctx, span)
	if err == nil {
		err = putSpan(native, ctx, span)
	}
	return
}

func verifyCascadingFields(native *native.NativeService, header *types.Header, ctx *Context) (snap *Snapshot, err error) {

	number := header.Number.Uint64()

	parent, err := getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	if parent.HeaderWithOptionalSnap.Header.Number.Uint64() != number-1 {
		err = errors.New("unknown ancestor")
		return
	}

	if parent.HeaderWithOptionalSnap.Header.Time+ctx.ExtraInfo.Period > header.Time {
		err = errors.New("ErrInvalidTimestamp")
		return
	}

	snap, err = getSnapshot(native, parent, ctx)
	if err != nil {
		err = fmt.Errorf("getSnapshot failed:%v", err)
		return
	}

	if isSprintStart(number, ctx.ExtraInfo.Sprint) {
		parentHeader := parent.HeaderWithOptionalSnap.Header
		parentValidatorBytes := parentHeader.Extra[extraVanity : len(parentHeader.Extra)-extraSeal]
		var newVals []*Validator
		newVals, err = ParseValidators(parentValidatorBytes)
		if err != nil {
			err = fmt.Errorf("ParseValidators failed:%v", err)
			return
		}
		v := getUpdatedValidatorSet(snap.ValidatorSet.Copy(), newVals)
		v.IncrementProposerPriority(1)
		snap.ValidatorSet = v
		snap.Hash = header.Hash()
	}

	_, err = verifySeal(native, header, ctx, parent, snap)

	return
}

func getUpdatedValidatorSet(oldValidatorSet *ValidatorSet, newVals []*Validator) *ValidatorSet {
	v := oldValidatorSet
	oldVals := v.Validators

	var changes []*Validator
	for _, ov := range oldVals {
		if f, ok := validatorContains(newVals, ov); ok {
			ov.VotingPower = f.VotingPower
		} else {
			ov.VotingPower = 0
		}

		changes = append(changes, ov)
	}

	for _, nv := range newVals {
		if _, ok := validatorContains(changes, nv); !ok {
			changes = append(changes, nv)
		}
	}

	v.UpdateWithChangeSet(changes)
	return v
}

func validatorContains(a []*Validator, x *Validator) (*Validator, bool) {
	for _, n := range a {
		if bytes.Compare(n.Address.Bytes(), x.Address.Bytes()) == 0 {
			return n, true
		}
	}
	return nil, false
}

func getSnapshot(native *native.NativeService, parent *HeaderWithDifficultySum, ctx *Context) (s *Snapshot, err error) {
	if parent.HeaderWithOptionalSnap.Snapshot != nil {
		s = parent.HeaderWithOptionalSnap.Snapshot
		return
	}

	if parent.SnapParentHash == nil {
		err = fmt.Errorf("both Snapshot and SnapParentHash nil")
		return
	}
	snapHeader, err := getHeader(native, *parent.SnapParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	if snapHeader.HeaderWithOptionalSnap.Snapshot == nil {
		err = fmt.Errorf("snapHeader has no Snapshot")
		return
	}

	s = snapHeader.HeaderWithOptionalSnap.Snapshot
	return
}

func isSprintStart(number, sprint uint64) bool {
	return number%sprint == 0
}

// for test
var mockSigner ecommon.Address

func verifySeal(native *native.NativeService, header *types.Header, ctx *Context, parent *HeaderWithDifficultySum, snap *Snapshot) (signer ecommon.Address, err error) {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		err = errors.New("unknown block")
		return
	}

	if mockSigner != (ecommon.Address{}) {
		return mockSigner, nil
	}
	// Resolve the authorization key and check against validators
	signer, err = ecrecover(header)
	if err != nil {
		return
	}
	if !snap.ValidatorSet.HasAddress(signer.Bytes()) {
		err = fmt.Errorf("UnauthorizedSignerError:%d signer:%s", number, signer.Hex())
		return
	}
	succession, err := snap.GetSignerSuccessionNumber(signer)
	if err != nil {
		return
	}
	if header.Time < parent.HeaderWithOptionalSnap.Header.Time+CalcProducerDelay(number, succession, ctx) {
		err = fmt.Errorf("BlockTooSoonError, n:%d,succession:%d", number, succession)
		return
	}

	difficulty := snap.Difficulty(signer)
	if header.Difficulty.Uint64() != difficulty {
		err = fmt.Errorf("WrongDifficultyError, n:%d, expected:%d, actual:%d", number, difficulty, header.Difficulty.Uint64())
		return
	}

	return
}

// CalcProducerDelay is the block delay algorithm based on block time, period, producerDelay and turn-ness of a signer
func CalcProducerDelay(number uint64, succession int, ctx *Context) uint64 {
	// When the block is the first block of the sprint, it is expected to be delayed by `producerDelay`.
	// That is to allow time for block propagation in the last sprint
	delay := ctx.ExtraInfo.Period
	if number%ctx.ExtraInfo.Sprint == 0 {
		delay = ctx.ExtraInfo.ProducerDelay
	}
	if succession > 0 {
		delay += uint64(succession) * ctx.ExtraInfo.BackupMultiplier
	}
	return delay
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header) (ecommon.Address, error) {
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return ecommon.Address{}, errors.New("extra-data 65 byte signature suffix missing")
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return ecommon.Address{}, err
	}
	var signer ecommon.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	return signer, nil
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash ecommon.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.Sum(hash[:0])
	return hash
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	err := rlp.Encode(w, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-65], // this will panic if extra is too short, should check before calling encodeSigHeader
		header.MixDigest,
		header.Nonce,
	})
	if err != nil {
		panic("can't encode: " + err.Error())
	}
}

// SyncCrossChainMsg ...
func (h *BorHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
