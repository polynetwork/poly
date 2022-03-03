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
package bytom

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"golang.org/x/crypto/sha3"
)

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// GenesisHeader ...
type GenesisHeader struct {
	Header         types.Header
	PrevValidators []HeightAndValidators
}

// SyncGenesisHeader synchronize the initial block header of bytom chain to poly relay chain
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	//fetch the initial block header binary code and chain ID from the poly native service data
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator of poly chain
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, checkWitness error: %v", err)
	}

	// look for genesis header of bytom chain from poly chain
	stored, err := isGenesisStored(native, params)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, isGenesisStored error: %v", err)
	}
	if stored {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, genesis had been initialized")
	}
	//this part varies from chain to chain
	//bytom genesis header that contains PrevValidators
	var genesis GenesisHeader
	//assign the parsed json-encoded header data to genesis GenesisHeader
	err = json.Unmarshal(params.GenesisHeader, &genesis)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, deserialize GenesisHeader err: %v", err)
	}
	//check the format validity of extra field
	signersBytes := len(genesis.Header.Extra) - extraVanity - extraSeal
	if signersBytes == 0 || signersBytes%ecommon.AddressLength != 0 {
		return fmt.Errorf("invalid signer list, signersBytes:%d", signersBytes)
	}
	if len(genesis.PrevValidators) != 1 {
		return fmt.Errorf("invalid PrevValidators")
	}
	//the block height of PrevValidators should be smaller than genesis header
	if genesis.Header.Number.Cmp(genesis.PrevValidators[0].Height) <= 0 {
		return fmt.Errorf("invalid height orders")
	}
	//parse the address of validators from the extra field
	validators, err := ParseValidators(genesis.Header.Extra[extraVanity : extraVanity+signersBytes])
	if err != nil {
		return
	}
	genesis.PrevValidators = append([]HeightAndValidators{
		{Height: genesis.Header.Number, Validators: validators},
	}, genesis.PrevValidators...)
	//store the information of genesis header to poly chain
	err = storeGenesis(native, params, &genesis)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncGenesisHeader, storeGenesis error: %v", err)
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

func getGenesis(native *native.NativeService, chainID uint64) (genesisHeader *GenesisHeader, err error) {

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
		genesisHeader = &GenesisHeader{}
		err = json.Unmarshal(genesisBytes, &genesisHeader)
		if err != nil {
			err = fmt.Errorf("getGenesis, json.Unmarshal err:%v", err)
			return
		}
	}

	return
}

//storeGenesis stores several data with corresponding keys for future use
//GENESIS_HEADER => the genesis header byte code
//HEADER_INDEX => the mapping of header hash and block header byte code, for querying block header by its hash
//CURRENT_HEADER_HEIGHT => current block height of side chain in poly relay chain
//MAIN_CHAIN => the mapping of block height and block header hash, for querying block header hash by its height
func storeGenesis(native *native.NativeService, params *scom.SyncGenesisHeaderParam, genesisHeader *GenesisHeader) (err error) {

	genesisBytes, err := json.Marshal(genesisHeader)
	if err != nil {
		return
	}
	//store the genesis header binary code with key GENESIS_HEADER
	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)),
		cstates.GenRawStorageItem(genesisBytes))

	headerWithSum := &HeaderWithDifficultySum{Header: &genesisHeader.Header, DifficultySum: genesisHeader.Header.Difficulty}

	//store the mapping between header hash and block header byte code with key HEADER_INDEX
	err = putHeaderWithSum(native, params.ChainID, headerWithSum)
	if err != nil {
		return
	}
	//store the genesisHeader height with key CURRENT_HEADER_HEIGHT
	putCanonicalHeight(native, params.ChainID, genesisHeader.Header.Number.Uint64())
	//store the mapping between genesisHeader height and header hash
	putCanonicalHash(native, params.ChainID, genesisHeader.Header.Number.Uint64(), genesisHeader.Header.Hash())

	scom.NotifyPutHeader(native, params.ChainID, genesisHeader.Header.Number.Uint64(), genesisHeader.Header.Hash().Hex())
	return
}

// ExtraInfo ...
type ExtraInfo struct {
	ChainID *big.Int // for bytom
}

// Context ...
type Context struct {
	ExtraInfo ExtraInfo
	ChainID   uint64
}

// HeaderWithChainID ...
type HeaderWithChainID struct {
	Header  *HeaderWithDifficultySum
	ChainID uint64
}

// HeaderWithDifficultySum ...
type HeaderWithDifficultySum struct {
	Header          *types.Header `json:"header"`
	DifficultySum   *big.Int      `json:"difficultySum"`
	EpochParentHash *ecommon.Hash `json:"epochParentHash"`
}

// SyncBlockHeader synchronize the consequent block header of bytom chain to poly relay chain
func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	//fetch the block header binary code and chain ID from the native transaction parameter
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("bytom Handler SyncBlockHeader, contract params deserialize error: %v", err)
	}
	//get the registered bytom chain information
	side, err := side_chain_manager.GetSideChain(native, headerParams.ChainID)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncBlockHeader, GetSideChain error: %v", err)
	}
	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("bytom Handler SyncBlockHeader, ExtraInfo Unmarshal error: %v", err)
	}

	ctx := &Context{ExtraInfo: extraInfo, ChainID: headerParams.ChainID}

	for _, v := range headerParams.Headers {
		var header types.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("bytom Handler SyncBlockHeader, deserialize header err: %v", err)
		}
		headerHash := header.Hash()
		//look for header hash from poly chain to make sure this header hasn't been synced
		exist, err := isHeaderExist(native, headerHash, ctx)
		if err != nil {
			return fmt.Errorf("bytom Handler SyncBlockHeader, isHeaderExist headerHash err: %v", err)
		}
		if exist {
			log.Warnf("bytom Handler SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}
		//make sure the parent block has already been synced
		parentExist, err := isHeaderExist(native, header.ParentHash, ctx)
		if err != nil {
			return fmt.Errorf("bytom Handler SyncBlockHeader, isHeaderExist ParentHash err: %v", err)
		}
		if !parentExist {
			log.Warnf("bytom Handler SyncBlockHeader, parent header not exist. Header: %s", string(v))
			continue
		}

		//Verify the legitimacy of the block header
		//This function refers to https://github.com/binance-chain/bytom/blob/master/consensus/parlia/parlia.go#L324-L374
		signer, err := verifySignature(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("bytom Handler SyncBlockHeader, verifySignature err: %v", err)
		}

		// get prev epochs, also checking recent limit
		phv, pphv, lastSeenHeight, err := getPrevHeightAndValidators(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("bytom Handler SyncBlockHeader, getPrevHeightAndValidators err: %v", err)
		}

		var (
			inTurnHV *HeightAndValidators
		)

		diffWithLastEpoch := big.NewInt(0).Sub(header.Number, phv.Height).Int64()
		if diffWithLastEpoch <= int64(len(pphv.Validators)/2) {
			// pphv is in effect
			inTurnHV = pphv

			if len(header.Extra) > extraVanity+extraSeal {
				return fmt.Errorf("bytom Handler SyncBlockHeader: can not change epoch continuously")
			}
		} else {
			// phv is in effect
			inTurnHV = phv
		}

		if lastSeenHeight > 0 {
			limit := int64(len(inTurnHV.Validators) / 2)
			if header.Number.Int64() <= lastSeenHeight+limit {
				return fmt.Errorf("bytom Handler SyncBlockHeader, RecentlySigned, lastSeenHeight:%d currentHeight:%d #V:%d", lastSeenHeight, header.Number.Int64(), len(inTurnHV.Validators))
			}
		}

		indexInTurn := int(header.Number.Uint64()) % len(inTurnHV.Validators)
		if indexInTurn < 0 {
			return fmt.Errorf("indexInTurn is negative:%d inTurnHV.Height:%d header.Number:%d", indexInTurn, inTurnHV.Height.Int64(), header.Number.Int64())
		}
		valid := false
		for idx, v := range inTurnHV.Validators {
			if v == signer {
				valid = true
				if indexInTurn == idx {
					if header.Difficulty.Cmp(diffInTurn) != 0 {
						return fmt.Errorf("invalid difficulty, got %v expect %v index:%v", header.Difficulty.Int64(), diffInTurn.Int64(), int(indexInTurn)%len(inTurnHV.Validators))
					}
				} else {
					if header.Difficulty.Cmp(diffNoTurn) != 0 {
						return fmt.Errorf("invalid difficulty, got %v expect %v index:%v", header.Difficulty.Int64(), diffNoTurn.Int64(), int(indexInTurn)%len(inTurnHV.Validators))
					}
				}
			}
		}
		if !valid {
			return fmt.Errorf("bytom Handler SyncBlockHeader, invalid signer")
		}
		//put verified bytom header into relay chain
		err = addHeader(native, &header, phv, ctx)
		if err != nil {
			return fmt.Errorf("bytom Handler SyncBlockHeader, addHeader err: %v", err)
		}

		scom.NotifyPutHeader(native, headerParams.ChainID, header.Number.Uint64(), header.Hash().Hex())
	}
	return nil
}

func isHeaderExist(native *native.NativeService, headerHash ecommon.Hash, ctx *Context) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(ctx.ChainID), headerHash.Bytes()))
	if err != nil {
		return false, fmt.Errorf("bytom Handler isHeaderExist error: %v", err)
	}

	return headerStore != nil, nil
}

func verifySignature(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {
	return verifyHeader(native, header, ctx)
}

// GetCanonicalHeight ...
func GetCanonicalHeight(native *native.NativeService, chainID uint64) (height uint64, err error) {
	heightStore, err := native.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		err = fmt.Errorf("bytom Handler GetCanonicalHeight err:%v", err)
		return
	}

	storeBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		err = fmt.Errorf("bytom Handler GetCanonicalHeight, GetValueFromRawStorageItem err:%v", err)
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
		err = fmt.Errorf("bytom Handler getCanonicalHash, GetValueFromRawStorageItem err:%v", err)
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

func addHeader(native *native.NativeService, header *types.Header, phv *HeightAndValidators, ctx *Context) (err error) {

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

	headerWithSum := &HeaderWithDifficultySum{Header: header, DifficultySum: externTd, EpochParentHash: phv.Hash}
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

// HeightAndValidators ...
type HeightAndValidators struct {
	Height     *big.Int
	Validators []ecommon.Address
	Hash       *ecommon.Hash
}

func getPrevHeightAndValidators(native *native.NativeService, header *types.Header, ctx *Context) (phv, pphv *HeightAndValidators, lastSeenHeight int64, err error) {

	genesis, err := getGenesis(native, ctx.ChainID)
	if err != nil {
		err = fmt.Errorf("bytom Handler getGenesis error: %v", err)
		return
	}

	if genesis == nil {
		err = fmt.Errorf("bytom Handler genesis not set")
		return
	}

	genesisHeaderHash := genesis.Header.Hash()
	if header.Hash() == genesisHeaderHash {
		err = fmt.Errorf("genesis header should not be synced again")
		return
	}

	lastSeenHeight = -1
	targetCoinbase := header.Coinbase
	if header.ParentHash == genesisHeaderHash {
		if genesis.Header.Coinbase == targetCoinbase {
			lastSeenHeight = genesis.Header.Number.Int64()
		}

		phv = &genesis.PrevValidators[0]
		phv.Hash = &genesisHeaderHash
		pphv = &genesis.PrevValidators[1]
		return
	}

	prevHeaderWithSum, err := getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		err = fmt.Errorf("bytom Handler getHeader error: %v", err)
		return
	}

	if prevHeaderWithSum.Header.Coinbase == targetCoinbase {
		lastSeenHeight = prevHeaderWithSum.Header.Number.Int64()
	} else {
		nextRecentParentHash := prevHeaderWithSum.Header.ParentHash
		defer func() {
			if err == nil {
				maxV := len(phv.Validators)
				if maxV < len(pphv.Validators) {
					maxV = len(pphv.Validators)
				}
				maxLimit := maxV / 2
				for i := 0; i < maxLimit-1; i++ {
					prevHeaderWithSum, err := getHeader(native, nextRecentParentHash, ctx.ChainID)
					if err != nil {
						err = fmt.Errorf("bytom Handler getHeader error: %v", err)
						return
					}
					if prevHeaderWithSum.Header.Coinbase == targetCoinbase {
						lastSeenHeight = prevHeaderWithSum.Header.Number.Int64()
						return
					}

					if nextRecentParentHash == genesisHeaderHash {
						return
					}
					nextRecentParentHash = prevHeaderWithSum.Header.ParentHash
				}
			}
		}()
	}

	var (
		validators     []ecommon.Address
		nextParentHash ecommon.Hash
	)

	currentPV := &phv

	for {

		if len(prevHeaderWithSum.Header.Extra) > extraVanity+extraSeal {
			validators, err = ParseValidators(prevHeaderWithSum.Header.Extra[extraVanity : len(prevHeaderWithSum.Header.Extra)-extraSeal])
			if err != nil {
				err = fmt.Errorf("bytom Handler ParseValidators error: %v", err)
				return
			}
			*currentPV = &HeightAndValidators{
				Height:     prevHeaderWithSum.Header.Number,
				Validators: validators,
			}
			switch *currentPV {
			case phv:
				hash := prevHeaderWithSum.Header.Hash()
				phv.Hash = &hash
				currentPV = &pphv
			case pphv:
				return
			default:
				err = fmt.Errorf("bug in bytom Handler")
				return
			}
		}

		nextParentHash = prevHeaderWithSum.Header.ParentHash
		if prevHeaderWithSum.EpochParentHash != nil {
			nextParentHash = *prevHeaderWithSum.EpochParentHash
		}

		if nextParentHash == genesisHeaderHash {
			switch *currentPV {
			case phv:
				phv = &genesis.PrevValidators[0]
				phv.Hash = &genesisHeaderHash
				pphv = &genesis.PrevValidators[1]
			case pphv:
				pphv = &genesis.PrevValidators[0]
			default:
				err = fmt.Errorf("bug in bytom Handler")
				return
			}
			return
		}

		prevHeaderWithSum, err = getHeader(native, nextParentHash, ctx.ChainID)
		if err != nil {
			err = fmt.Errorf("bytom Handler getHeader error: %v", err)
			return
		}

	}
}

func getHeader(native *native.NativeService, hash ecommon.Hash, chainID uint64) (headerWithSum *HeaderWithDifficultySum, err error) {

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("bytom Handler getHeader error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("bytom Handler getHeader, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("bytom Handler getHeader, deserialize headerBytes from raw storage item err:%v", err)
	}
	headerWithSum = &HeaderWithDifficultySum{}
	if err := json.Unmarshal(storeBytes, &headerWithSum); err != nil {
		return nil, fmt.Errorf("bytom Handler getHeader, deserialize header error: %v", err)
	}

	return
}

var (
	inMemoryHeaders = 400
	inMemoryGenesis = 40
	extraVanity     = 32                       // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal       = crypto.SignatureLength   // Fixed number of extra-data suffix bytes reserved for signer seal
	uncleHash       = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
	diffInTurn      = big.NewInt(2)            // Block difficulty for in-turn signatures
	diffNoTurn      = big.NewInt(1)            // Block difficulty for out-of-turn signatures

	GasLimitBoundDivisor uint64 = 256 // The bound divisor of the gas limit, used in update calculations.
)

//https://github.com/binance-chain/bytom/blob/master/consensus/parlia/parlia.go#L324-L374
func verifyHeader(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		err = errors.New("block in the future")
		return
	}

	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		err = errors.New("extra-data 32 byte vanity prefix missing")
		return
	}
	if len(header.Extra) < extraVanity+extraSeal {
		err = errors.New("extra-data 65 byte signature suffix missing")
		return
	}

	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal

	if signersBytes%ecommon.AddressLength != 0 {
		err = errors.New("invalid signer list")
		return
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
	if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
		err = errors.New("invalid difficulty")
		return
	}

	// All basic checks passed, verify cascading fields
	return verifyCascadingFields(native, header, ctx)
}

func verifyCascadingFields(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {

	number := header.Number.Uint64()

	parent, err := getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	if parent.Header.Number.Uint64() != number-1 {
		err = errors.New("unknown ancestor")
		return
	}

	// Verify that the gas limit is <= 2^63-1
	capacity := uint64(0x7fffffffffffffff)
	if header.GasLimit > capacity {
		err = fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, capacity)
		return
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		err = fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
		return
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.Header.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.Header.GasLimit / GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		err = fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.Header.GasLimit, limit)
		return
	}

	return verifySeal(native, header, ctx)
}

// for test
var mockSigner ecommon.Address

func verifySeal(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {
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
	signer, err = ecrecover(header, ctx.ExtraInfo.ChainID)
	if err != nil {
		return
	}

	if signer != header.Coinbase {
		err = errors.New("coinbase do not match with signature")
		return
	}

	return
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header, chainID *big.Int) (ecommon.Address, error) {
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return ecommon.Address{}, errors.New("extra-data 65 byte signature suffix missing")
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header, chainID).Bytes(), signature)
	if err != nil {
		return ecommon.Address{}, err
	}
	var signer ecommon.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	return signer, nil
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header, chainID *big.Int) (hash ecommon.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header, chainID)
	hasher.Sum(hash[:0])
	return hash
}

func encodeSigHeader(w io.Writer, header *types.Header, chainID *big.Int) {
	err := rlp.Encode(w, []interface{}{
		chainID,
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

// ParseValidators ...
func ParseValidators(validatorsBytes []byte) ([]ecommon.Address, error) {
	if len(validatorsBytes)%ecommon.AddressLength != 0 {
		return nil, errors.New("invalid validators bytes")
	}
	n := len(validatorsBytes) / ecommon.AddressLength
	result := make([]ecommon.Address, n)
	for i := 0; i < n; i++ {
		address := make([]byte, ecommon.AddressLength)
		copy(address, validatorsBytes[i*ecommon.AddressLength:(i+1)*ecommon.AddressLength])
		result[i] = ecommon.BytesToAddress(address)
	}
	return result, nil
}

// SyncCrossChainMsg ...
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
