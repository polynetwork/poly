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
package msc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// SyncGenesisHeader ...
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	side, err := side_chain_manager.GetSideChain(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, GetSideChain error: %v", err)
	}
	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, ExtraInfo Unmarshal error: %v", err)
	}
	if extraInfo.Epoch == 0 {
		return fmt.Errorf("msc Handler SyncGenesisHeader, invalid epoch")
	}
	if extraInfo.Period == 0 {
		return fmt.Errorf("msc Handler SyncGenesisHeader, invalid period")
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, checkWitness error: %v", err)
	}

	// can only store once
	stored, err := isGenesisStored(native, params)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, isGenesisStored error: %v", err)
	}
	if stored {
		return fmt.Errorf("msc Handler SyncGenesisHeader, genesis had been initialized")
	}

	var genesis types.Header
	err = json.Unmarshal(params.GenesisHeader, &genesis)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, deserialize GenesisHeader err: %v", err)
	}

	if genesis.Number.Uint64()%extraInfo.Epoch != 0 {
		return fmt.Errorf("invalid genesis height:%d", genesis.Number.Uint64())
	}
	signersBytes := len(genesis.Extra) - extraVanity - extraSeal
	if signersBytes == 0 || signersBytes%ecommon.AddressLength != 0 {
		return fmt.Errorf("invalid signer list, signersBytes:%d", signersBytes)
	}

	err = storeGenesis(native, params, &genesis)
	if err != nil {
		return fmt.Errorf("msc Handler SyncGenesisHeader, storeGenesis error: %v", err)
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

func getGenesis(native *native.NativeService, chainID uint64) (genesisHeader *types.Header, err error) {

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
		genesisHeader = &types.Header{}
		err = json.Unmarshal(genesisBytes, &genesisHeader)
		if err != nil {
			err = fmt.Errorf("getGenesis, json.Unmarshal err:%v", err)
			return
		}
	}

	return
}

func storeGenesis(native *native.NativeService, params *scom.SyncGenesisHeaderParam, genesisHeader *types.Header) (err error) {

	genesisBytes, err := json.Marshal(genesisHeader)
	if err != nil {
		return
	}

	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)),
		cstates.GenRawStorageItem(genesisBytes))

	headerWithSum := &HeaderWithDifficultySum{Header: genesisHeader, DifficultySum: genesisHeader.Difficulty}

	err = putHeaderWithSum(native, params.ChainID, headerWithSum)
	if err != nil {
		return
	}

	putCanonicalHeight(native, params.ChainID, genesisHeader.Number.Uint64())
	putCanonicalHash(native, params.ChainID, genesisHeader.Number.Uint64(), genesisHeader.Hash())

	scom.NotifyPutHeader(native, params.ChainID, genesisHeader.Number.Uint64(), genesisHeader.Hash().Hex())
	return
}

// ExtraInfo ...
type ExtraInfo struct {
	ChainID *big.Int // for msc
	Period  uint64
	Epoch   uint64
}

// Context ...
type Context struct {
	ExtraInfo ExtraInfo
	ChainID   uint64
}

// HeaderWithDifficultySum ...
type HeaderWithDifficultySum struct {
	Header        *types.Header `json:"header"`
	DifficultySum *big.Int      `json:"difficultySum"`
	// 1. empty for epoch header
	// 2. for non-epoch headers, either points to epoch header or a vote header
	LastVoteParentOrEpoch *ecommon.Hash `json:"lastVoteParentOrEpoch"`
}

// SyncBlockHeader ...
func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("msc Handler SyncBlockHeader, contract params deserialize error: %v", err)
	}

	side, err := side_chain_manager.GetSideChain(native, headerParams.ChainID)
	if err != nil {
		return fmt.Errorf("msc Handler SyncBlockHeader, GetSideChain error: %v", err)
	}
	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("msc Handler SyncBlockHeader, ExtraInfo Unmarshal error: %v", err)
	}

	ctx := &Context{ExtraInfo: extraInfo, ChainID: headerParams.ChainID}

	for _, v := range headerParams.Headers {
		var header types.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("msc Handler SyncBlockHeader, deserialize header err: %v", err)
		}
		headerHash := header.Hash()

		exist, err := isHeaderExist(native, headerHash, ctx)
		if err != nil {
			return fmt.Errorf("msc Handler SyncBlockHeader, isHeaderExist headerHash err: %v", err)
		}
		if exist {
			log.Warnf("msc Handler SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}

		parentExist, err := isHeaderExist(native, header.ParentHash, ctx)
		if err != nil {
			return fmt.Errorf("msc Handler SyncBlockHeader, isHeaderExist ParentHash err: %v", err)
		}
		if !parentExist {
			log.Warnf("msc Handler SyncBlockHeader, parent header not exist. Header: %s", string(v))
			continue
		}

		err = verifyHeader(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("msc Handler SyncBlockHeader, verifyHeader err: %v", err)
		}

		err = addHeader(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("msc Handler SyncBlockHeader, addHeader err: %v", err)
		}

		scom.NotifyPutHeader(native, headerParams.ChainID, header.Number.Uint64(), header.Hash().Hex())
	}
	return nil
}

func isHeaderExist(native *native.NativeService, headerHash ecommon.Hash, ctx *Context) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(ctx.ChainID), headerHash.Bytes()))
	if err != nil {
		return false, fmt.Errorf("msc Handler isHeaderExist error: %v", err)
	}

	return headerStore != nil, nil
}

// GetCanonicalHeight ...
func GetCanonicalHeight(native *native.NativeService, chainID uint64) (height uint64, err error) {
	heightStore, err := native.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		err = fmt.Errorf("msc Handler GetCanonicalHeight err:%v", err)
		return
	}

	storeBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		err = fmt.Errorf("msc Handler GetCanonicalHeight, GetValueFromRawStorageItem err:%v", err)
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
		err = fmt.Errorf("msc Handler getCanonicalHash, GetValueFromRawStorageItem err:%v", err)
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
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), headerWithSum.Header.Hash().Bytes()),
		cstates.GenRawStorageItem(headerBytes))
	return
}

func putCanonicalHeight(native *native.NativeService, chainID uint64, height uint64) {
	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(utils.GetUint64Bytes(uint64(height))))
}

func addHeader(native *native.NativeService, header *types.Header, ctx *Context) (err error) {

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

	headerWithSum := &HeaderWithDifficultySum{Header: header, DifficultySum: externTd}
	if header.Number.Uint64()%ctx.ExtraInfo.Epoch != 0 {
		if parentHeader.Header.Number.Uint64()%ctx.ExtraInfo.Epoch == 0 {
			lastVoteParentOrEpoch := parentHeader.Header.Hash()
			headerWithSum.LastVoteParentOrEpoch = &lastVoteParentOrEpoch
		} else {
			if parentHeader.Header.Coinbase != (ecommon.Address{}) {
				lastVoteParentOrEpoch := parentHeader.Header.Hash()
				headerWithSum.LastVoteParentOrEpoch = &lastVoteParentOrEpoch
			} else {
				headerWithSum.LastVoteParentOrEpoch = parentHeader.LastVoteParentOrEpoch
			}
		}
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
			headHash = headHeader.Header.ParentHash
			cheight--
		}

		// Extend the canonical chain with the new header
		putCanonicalHash(native, ctx.ChainID, header.Number.Uint64(), header.Hash())
		putCanonicalHeight(native, ctx.ChainID, header.Number.Uint64())
	}

	return nil
}

func snapshot(native *native.NativeService, number uint64, hash ecommon.Hash, targetSigner ecommon.Address, ctx *Context) (snap *Snapshot, lastSeenHeight uint64, err error) {
	var (
		headerWSs []*HeaderWithDifficultySum
		headerWS  *HeaderWithDifficultySum
	)

	startHash := hash
	genesis, err := getGenesis(native, ctx.ChainID)
	if err != nil {
		err = fmt.Errorf("msc Handler snapshot getGenesis error: %v", err)
		return
	}

	if genesis == nil {
		err = fmt.Errorf("msc Handler snapshot genesis not set")
		return
	}

	if number < genesis.Number.Uint64() {
		err = fmt.Errorf("msc Handler snapshot header before genesis is not allowed")
		return
	}

	var signer ecommon.Address

	for snap == nil {

		headerWS, err = getHeader(native, hash, ctx.ChainID)
		if err != nil {
			err = fmt.Errorf("msc Handler snapshot getHeader error: %v", err)
			return
		}

		if headerWS.LastVoteParentOrEpoch == nil {
			signers := make([]ecommon.Address, (len(headerWS.Header.Extra)-extraVanity-extraSeal)/ecommon.AddressLength)
			for i := 0; i < len(signers); i++ {
				copy(signers[i][:], headerWS.Header.Extra[extraVanity+i*ecommon.AddressLength:])
			}

			signer, err = ecrecover(headerWS.Header)
			if err != nil {
				err = fmt.Errorf("msc Handler snapshot ecrecover error: %v", err)
				return
			}
			if targetSigner == signer {
				lastSeenHeight = headerWS.Header.Number.Uint64()
			}
			snap = newSnapshot(headerWS.Header.Number.Uint64(), hash, signers, ctx)
			break
		}

		if headerWS.Header.Coinbase != (ecommon.Address{}) {
			headerWSs = append(headerWSs, headerWS)
		}

		// LastVoteParentOrEpoch must be non nil for non-epoch headers
		hash = *headerWS.LastVoteParentOrEpoch
	}

	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headerWSs)/2; i++ {
		headerWSs[i], headerWSs[len(headerWSs)-1-i] = headerWSs[len(headerWSs)-1-i], headerWSs[i]
	}

	err = snap.apply(headerWSs, targetSigner, &lastSeenHeight)
	if err != nil {
		err = fmt.Errorf("msc Handler snapshot apply error: %v", err)
		return
	}

	if lastSeenHeight > 0 {
		return
	}

	// try to search enough recent
	toSearch := len(snap.Signers) / 2
	for i := 0; i < toSearch; i++ {
		headerWS, err = getHeader(native, startHash, ctx.ChainID)
		if err != nil {
			err = fmt.Errorf("msc Handler snapshot getHeader error: %v", err)
			return
		}

		if number != headerWS.Header.Number.Uint64() {
			err = fmt.Errorf("bug happened in msc")
			return
		}
		signer, err = ecrecover(headerWS.Header)
		if err != nil {
			err = fmt.Errorf("msc Handler snapshot ecrecover error: %v", err)
			return
		}
		if targetSigner == signer {
			lastSeenHeight = headerWS.Header.Number.Uint64()
			break
		}
		number, startHash = number-1, headerWS.Header.ParentHash
		if number < genesis.Number.Uint64() {
			break
		}
	}
	return
}

func getHeader(native *native.NativeService, hash ecommon.Hash, chainID uint64) (headerWithSum *HeaderWithDifficultySum, err error) {

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("msc Handler getHeader error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("msc Handler getHeader, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("msc Handler getHeader, deserialize headerBytes from raw storage item err:%v", err)
	}
	headerWithSum = &HeaderWithDifficultySum{}
	if err := json.Unmarshal(storeBytes, &headerWithSum); err != nil {
		return nil, fmt.Errorf("msc Handler getHeader, deserialize header error: %v", err)
	}

	return
}

var (
	extraVanity   = 32                                       // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal     = crypto.SignatureLength                   // Fixed number of extra-data suffix bytes reserved for signer seal
	uncleHash     = types.CalcUncleHash(nil)                 // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
	diffInTurn    = big.NewInt(2)                            // Block difficulty for in-turn signatures
	diffNoTurn    = big.NewInt(1)                            // Block difficulty for out-of-turn signatures
	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

)

var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errInvalidCheckpointVote is returned if a checkpoint/epoch transition block
	// has a vote nonce set to non-zeroes.
	errInvalidCheckpointVote = errors.New("vote nonce in checkpoint block non-zero")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidCheckpointSigners is returned if a checkpoint block contains an
	// invalid list of signers (i.e. non divisible by 20 bytes).
	errInvalidCheckpointSigners = errors.New("invalid signer list on checkpoint block")

	// errMismatchingCheckpointSigners is returned if a checkpoint block contains a
	// list of signers different than the one the local node calculated.
	errMismatchingCheckpointSigners = errors.New("mismatching signer list on checkpoint block")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash = errors.New("non empty uncle hash")

	// errInvalidDifficulty is returned if the difficulty of a block neither 1 or 2.
	errInvalidDifficulty = errors.New("invalid difficulty")

	// errInvalidCheckpointBeneficiary is returned if a checkpoint/epoch transition
	// block has a beneficiary set to non-zeroes.
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")

	// errFutureBlock is returned when a block's timestamp is in the future according
	// to the current node.
	errFutureBlock = errors.New("block in the future")

	// errUnknownAncestor is returned when validating a block requires an ancestor
	// that is unknown.
	errUnknownAncestor = errors.New("unknown ancestor")

	// errInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")
)

func verifyHeader(native *native.NativeService, header *types.Header, ctx *Context) (err error) {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		err = errFutureBlock
		return
	}

	// Checkpoint blocks need to enforce zero beneficiary
	checkpoint := (number % ctx.ExtraInfo.Epoch) == 0
	if checkpoint && header.Coinbase != (ecommon.Address{}) {
		return errInvalidCheckpointBeneficiary
	}

	// Nonces must be 0x00..0 or 0xff..f, zeroes enforced on checkpoints
	if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidVote
	}

	if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidCheckpointVote
	}

	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		err = errMissingVanity
		return
	}
	if len(header.Extra) < extraVanity+extraSeal {
		err = errMissingSignature
		return
	}

	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	if !checkpoint && signersBytes != 0 {
		return errExtraSigners
	}

	if checkpoint && (signersBytes == 0 || signersBytes%ecommon.AddressLength != 0) {
		return errInvalidCheckpointSigners
	}

	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (ecommon.Hash{}) {
		err = errInvalidMixDigest
		return
	}

	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.UncleHash != uncleHash {
		err = errInvalidUncleHash
		return
	}

	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
		err = errInvalidDifficulty
		return
	}

	// All basic checks passed, verify cascading fields
	return verifyCascadingFields(native, header, ctx)
}

func verifyCascadingFields(native *native.NativeService, header *types.Header, ctx *Context) (err error) {

	number := header.Number.Uint64()

	parent, err := getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	if parent.Header.Number.Uint64() != number-1 {
		err = errUnknownAncestor
		return
	}

	if parent.Header.Time+ctx.ExtraInfo.Period > header.Time {
		return errInvalidTimestamp
	}

	return verifySeal(native, header, ctx)
}

// for test
var mockSigner ecommon.Address

func verifySeal(native *native.NativeService, header *types.Header, ctx *Context) (err error) {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		err = errUnknownBlock
		return
	}

	signer := mockSigner
	if signer == (ecommon.Address{}) {
		// Resolve the authorization key and check against validators
		signer, err = ecrecover(header)
		if err != nil {
			return
		}
	}

	snap, lastSeenHeight, err := snapshot(native, header.Number.Uint64()-1, header.ParentHash, signer, ctx)
	if err != nil {
		return fmt.Errorf("msc Handler SyncBlockHeader, snapshot err: %v", err)
	}

	if number%ctx.ExtraInfo.Epoch == 0 {
		signers := make([]byte, len(snap.Signers)*ecommon.AddressLength)
		for i, signer := range snap.signers() {
			copy(signers[i*ecommon.AddressLength:], signer[:])
		}
		extraSuffix := len(header.Extra) - extraSeal
		if !bytes.Equal(header.Extra[extraVanity:extraSuffix], signers) {
			return errMismatchingCheckpointSigners
		}
	}

	if lastSeenHeight > 0 {
		limit := uint64(len(snap.Signers)/2) + 1
		if header.Number.Uint64() < lastSeenHeight+limit {
			return fmt.Errorf("msc Handler SyncBlockHeader, RecentlySigned, lastSeenHeight:%d currentHeight:%d #V:%d", lastSeenHeight, header.Number.Int64(), len(snap.Signers))
		}
	}

	var offset int
	inturn := snap.inturn(header.Number.Uint64(), signer, &offset)
	if inturn {
		if header.Difficulty.Cmp(diffInTurn) != 0 {
			return fmt.Errorf("invalid difficulty, got %d expect %d offset:%d signers:%d", header.Difficulty.Int64(), diffInTurn.Int64(), offset, len(snap.Signers))
		}
	} else {
		if header.Difficulty.Cmp(diffNoTurn) != 0 {
			return fmt.Errorf("invalid difficulty, got %d expect %d offset:%d signers:%d", header.Difficulty.Int64(), diffNoTurn.Int64(), offset, len(snap.Signers))
		}
	}

	return
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header) (ecommon.Address, error) {
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return ecommon.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(clique.SealHash(header).Bytes(), signature)
	if err != nil {
		return ecommon.Address{}, err
	}
	var signer ecommon.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	return signer, nil
}

// SyncCrossChainMsg ...
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
