package bsc

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
	lru "github.com/hashicorp/golang-lru"
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
	recentHeaders *lru.ARCCache
	genesis       *lru.ARCCache
}

// NewHandler ...
func NewHandler() *Handler {
	recentHeaders, err := lru.NewARC(inMemoryHeaders)
	if err != nil {
		panic(err)
	}
	genesis, err := lru.NewARC(inMemoryGenesis)

	return &Handler{recentHeaders: recentHeaders, genesis: genesis}
}

// GenesisHeader ...
type GenesisHeader struct {
	Header         types.Header
	PrevValidators []HeightAndValidators
}

// SyncGenesisHeader ...
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, checkWitness error: %v", err)
	}

	// can only store once
	stored, err := h.isGenesisStored(native, params)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, isGenesisStored error: %v", err)
	}
	if stored {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, genesis had been initialized")
	}

	var genesis GenesisHeader
	err = json.Unmarshal(params.GenesisHeader, &genesis)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, deserialize GenesisHeader err: %v", err)
	}

	signersBytes := len(genesis.Header.Extra) - extraVanity - extraSeal
	if signersBytes == 0 || signersBytes%ecommon.AddressLength != 0 {
		return fmt.Errorf("invalid signer list, signersBytes:%d", signersBytes)
	}

	if len(genesis.PrevValidators) != 2 {
		return fmt.Errorf("invalid PrevValidators")
	}
	validators, err := ParseValidators(genesis.Header.Extra[extraVanity : extraVanity+signersBytes])
	if err != nil {
		return
	}
	genesis.PrevValidators = append([]HeightAndValidators{
		{Height: genesis.Header.Number, Validators: validators},
	}, genesis.PrevValidators...)

	err = h.storeGenesis(native, params, &genesis)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncGenesisHeader, storeGenesis error: %v", err)
	}

	return
}

func (h *Handler) isGenesisStored(native *native.NativeService, params *scom.SyncGenesisHeaderParam) (stored bool, err error) {
	genesis, err := h.getGenesis(native, params.ChainID)
	if err != nil {
		return
	}

	stored = genesis != nil
	return
}

func (h *Handler) getGenesis(native *native.NativeService, chainID uint64) (genesis *GenesisHeader, err error) {
	cache, ok := h.genesis.Get(chainID)
	if ok {
		genesis = cache.(*GenesisHeader)
		if genesis != nil {
			return
		}
	}

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
		genesis = &GenesisHeader{}
		err = json.Unmarshal(genesisBytes, &genesis)
		if err != nil {
			err = fmt.Errorf("getGenesis, json.Unmarshal err:%v", err)
			return
		}
	}

	h.genesis.Add(chainID, genesis)
	return
}

func (h *Handler) storeGenesis(native *native.NativeService, params *scom.SyncGenesisHeaderParam, genesis *GenesisHeader) (err error) {

	genesisBytes, err := json.Marshal(genesis)
	if err != nil {
		return
	}

	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)),
		cstates.GenRawStorageItem(genesisBytes))

	headerWithSum := &HeaderWithDifficultySum{Header: genesis.Header, DifficultySum: genesis.Header.Difficulty}

	err = h.putHeaderWithSum(native, params.ChainID, headerWithSum)
	if err != nil {
		return
	}

	h.putCanonicalHeight(native, params.ChainID, genesis.Header.Number.Uint64())
	h.putCanonicalHash(native, params.ChainID, genesis.Header.Number.Uint64(), genesis.Header.Hash())

	h.genesis.Add(params.ChainID, genesis)
	return
}

// ExtraInfo ...
type ExtraInfo struct {
	ChainID *big.Int // for bsc
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
	Header        types.Header `json:"header"`
	DifficultySum *big.Int     `json:"difficultySum"`
}

// SyncBlockHeader ...
func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("bsc Handler SyncBlockHeader, contract params deserialize error: %v", err)
	}

	side, err := side_chain_manager.GetSideChain(native, headerParams.ChainID)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncBlockHeader, GetSideChain error: %v", err)
	}
	var extraInfo ExtraInfo
	err = json.Unmarshal(side.ExtraInfo, &extraInfo)
	if err != nil {
		return fmt.Errorf("bsc Handler SyncBlockHeader, ExtraInfo Unmarshal error: %v", err)
	}

	ctx := &Context{ExtraInfo: extraInfo, ChainID: headerParams.ChainID}
	for _, v := range headerParams.Headers {
		var header types.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("bsc Handler SyncBlockHeader, deserialize header err: %v", err)
		}
		headerHash := header.Hash()

		exist, err := h.isHeaderExist(native, headerHash, ctx)
		if err != nil {
			return fmt.Errorf("bsc Handler SyncBlockHeader, isHeaderExist headerHash err: %v", err)
		}
		if exist {
			log.Warnf("bsc Handler SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}

		parentExist, err := h.isHeaderExist(native, header.ParentHash, ctx)
		if err != nil {
			return fmt.Errorf("bsc Handler SyncBlockHeader, isHeaderExist ParentHash err: %v", err)
		}
		if !parentExist {
			log.Warnf("bsc Handler SyncBlockHeader, parent header not exist. Header: %s", string(v))
			continue
		}

		signer, err := h.verifySignature(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("bsc Handler SyncBlockHeader, verifySignature err: %v", err)
		}

		// get prev epochs, also checking recent limit
		phv, pphv, ppphv, err := h.getPrevHeightAndValidators(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("bsc Handler SyncBlockHeader, getPrevHeightAndValidators err: %v", err)
		}

		var (
			inTurnHV, prevHV *HeightAndValidators
		)

		diffWithLastEpoch := big.NewInt(0).Sub(header.Number, phv.Height).Int64()
		if diffWithLastEpoch <= int64(len(pphv.Validators)/2) {
			// pphv is in effect
			inTurnHV = pphv
			prevHV = ppphv
		} else {
			// phv is in effect
			inTurnHV = phv
			prevHV = pphv
		}

		inTurnEpochStartHeight := big.NewInt(0).Add(inTurnHV.Height, big.NewInt(int64(len(prevHV.Validators)/2)))
		indexInTurn := big.NewInt(0).Sub(header.Number, inTurnEpochStartHeight).Int64() - 1
		if indexInTurn < 0 {
			return fmt.Errorf("indexInTurn is negative:%d inTurnHV.Height:%d prevHV.Validators:%d inTurnEpochStartHeight:%d header.Number:%d", indexInTurn, inTurnHV.Height.Int64(), len(prevHV.Validators), inTurnEpochStartHeight, header.Number.Int64())
		}
		valid := false
		for idx, v := range inTurnHV.Validators {
			if v == signer {
				valid = true
				if int(indexInTurn)%len(inTurnHV.Validators) == idx {
					if header.Difficulty.Cmp(diffInTurn) != 0 {
						return fmt.Errorf("invalid difficulty, got %v expect %v", header.Difficulty.Int64(), diffInTurn.Int64())
					}
				} else {
					if header.Difficulty.Cmp(diffNoTurn) != 0 {
						return fmt.Errorf("invalid difficulty, got %v expect %v", header.Difficulty.Int64(), diffNoTurn.Int64())
					}
				}
			}
		}
		if !valid {
			return fmt.Errorf("bsc Handler SyncBlockHeader, invalid signer")
		}

		err = h.addHeader(native, &header, ctx)
		if err != nil {
			return fmt.Errorf("bsc Handler SyncBlockHeader, addHeader err: %v", err)
		}

	}
	return nil
}

func (h *Handler) isHeaderExist(native *native.NativeService, headerHash ecommon.Hash, ctx *Context) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(ctx.ChainID), headerHash.Bytes()))
	if err != nil {
		return false, fmt.Errorf("bsc Handler isHeaderExist error: %v", err)
	}

	return headerStore != nil, nil
}

func (h *Handler) verifySignature(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {
	return h.verifyHeader(native, header, ctx)
}

func (h *Handler) getCanonicalHeight(native *native.NativeService, chainID uint64) (height uint64, err error) {
	heightStore, err := native.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		err = fmt.Errorf("bsc Handler getCanonicalHeight err:%v", err)
		return
	}

	storeBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		err = fmt.Errorf("bsc Handler getCanonicalHeight, GetValueFromRawStorageItem err:%v", err)
		return
	}

	height = utils.GetBytesUint64(storeBytes)
	return
}

func (h *Handler) getCanonicalHeader(native *native.NativeService, chainID uint64, height uint64) (headerWithSum *HeaderWithDifficultySum, err error) {
	hash, err := h.getCanonicalHash(native, chainID, height)
	if err != nil {
		return
	}

	if hash == (ecommon.Hash{}) {
		return
	}

	headerWithSum, err = h.getHeader(native, hash, chainID)
	return
}

func (h *Handler) deleteCanonicalHash(native *native.NativeService, chainID uint64, height uint64) {
	native.GetCacheDB().Delete(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
}

func (h *Handler) getCanonicalHash(native *native.NativeService, chainID uint64, height uint64) (hash ecommon.Hash, err error) {
	hashBytesStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return
	}

	if hashBytesStore == nil {
		return
	}

	hashBytes, err := cstates.GetValueFromRawStorageItem(hashBytesStore)
	if err != nil {
		err = fmt.Errorf("bsc Handler getCanonicalHash, GetValueFromRawStorageItem err:%v", err)
		return
	}

	hash = ecommon.BytesToHash(hashBytes)
	return
}

func (h *Handler) putCanonicalHash(native *native.NativeService, chainID uint64, height uint64, hash ecommon.Hash) {
	native.GetCacheDB().Put(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)),
		cstates.GenRawStorageItem(hash.Bytes()))
}

func (h *Handler) putHeaderWithSum(native *native.NativeService, chainID uint64, headerWithSum *HeaderWithDifficultySum) (err error) {

	headerBytes, err := json.Marshal(headerWithSum)
	if err != nil {
		return
	}

	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), headerWithSum.Header.Hash().Bytes()),
		cstates.GenRawStorageItem(headerBytes))
	return
}

func (h *Handler) putCanonicalHeight(native *native.NativeService, chainID uint64, height uint64) {
	native.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(utils.GetUint64Bytes(uint64(height))))
}

func (h *Handler) addHeader(native *native.NativeService, header *types.Header, ctx *Context) (err error) {

	parentHeader, err := h.getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	cheight, err := h.getCanonicalHeight(native, ctx.ChainID)
	if err != nil {
		return
	}
	cheader, err := h.getCanonicalHeader(native, ctx.ChainID, cheight)
	if err != nil {
		return
	}
	if cheader == nil {
		err = fmt.Errorf("getCanonicalHeader returns nil")
		return
	}

	localTd := cheader.DifficultySum
	externTd := new(big.Int).Add(header.Difficulty, parentHeader.DifficultySum)

	headerWithSum := &HeaderWithDifficultySum{Header: *header, DifficultySum: externTd}
	err = h.putHeaderWithSum(native, ctx.ChainID, headerWithSum)
	if err != nil {
		return
	}

	if externTd.Cmp(localTd) > 0 {
		// Delete any canonical number assignments above the new head
		var headerWithSum *HeaderWithDifficultySum
		for i := header.Number.Uint64() + 1; ; i++ {
			headerWithSum, err = h.getCanonicalHeader(native, ctx.ChainID, i)
			if err != nil {
				return
			}
			if headerWithSum == nil {
				break
			}

			h.deleteCanonicalHash(native, ctx.ChainID, i)
		}

		// Overwrite any stale canonical number assignments
		var (
			hash       ecommon.Hash
			headHeader *HeaderWithDifficultySum
		)
		cheight := header.Number.Uint64() - 1
		headHash := header.ParentHash

		headHeader, err = h.getHeader(native, headHash, ctx.ChainID)
		if err != nil {
			return
		}
		for {
			hash, err = h.getCanonicalHash(native, ctx.ChainID, cheight)
			if err != nil {
				return
			}
			if hash == headHash {
				break
			}

			h.putCanonicalHash(native, ctx.ChainID, cheight, hash)
			headHash = headHeader.Header.ParentHash
			cheight--
			headHeader, err = h.getHeader(native, headHash, ctx.ChainID)
			if err != nil {
				return
			}
		}

		// Extend the canonical chain with the new header
		h.putCanonicalHash(native, ctx.ChainID, header.Number.Uint64(), header.Hash())
		h.putCanonicalHeight(native, ctx.ChainID, header.Number.Uint64())
	}

	h.recentHeaders.Add(header.Hash(), &HeaderWithChainID{Header: headerWithSum, ChainID: ctx.ChainID})

	return nil
}

// HeightAndValidators ...
type HeightAndValidators struct {
	Height     *big.Int
	Validators []ecommon.Address
}

func (h *Handler) getPrevHeightAndValidators(native *native.NativeService, header *types.Header, ctx *Context) (phv *HeightAndValidators, pphv *HeightAndValidators, ppphv *HeightAndValidators, err error) {

	genesis, err := h.getGenesis(native, ctx.ChainID)
	if err != nil {
		err = fmt.Errorf("bsc Handler getGenesis error: %v", err)
		return
	}

	if genesis == nil {
		err = fmt.Errorf("bsc Handler genesis not set")
		return
	}

	if header.Hash() == genesis.Header.Hash() {
		err = fmt.Errorf("genesis header should not be synced again")
		return
	}

	var (
		prevHeaderWithSum *HeaderWithDifficultySum
		validators        []ecommon.Address
	)
	currentPV := &phv
	for {
		if header.ParentHash == genesis.Header.Hash() {
			switch *currentPV {
			case phv:
				phv = &genesis.PrevValidators[0]
				pphv = &genesis.PrevValidators[1]
				ppphv = &genesis.PrevValidators[2]
			case pphv:
				pphv = &genesis.PrevValidators[0]
				ppphv = &genesis.PrevValidators[1]
			case ppphv:
				ppphv = &genesis.PrevValidators[0]
				return
			default:
				err = fmt.Errorf("bug in bsc Handler")
				return
			}
		}
		prevHeaderWithSum, err = h.getHeader(native, header.ParentHash, ctx.ChainID)
		if err != nil {
			err = fmt.Errorf("bsc Handler getHeader error: %v", err)
			return
		}

		if len(prevHeaderWithSum.Header.Extra) > extraVanity+extraSeal {
			validators, err = ParseValidators(prevHeaderWithSum.Header.Extra[extraVanity : len(prevHeaderWithSum.Header.Extra)-extraSeal])
			if err != nil {
				err = fmt.Errorf("bsc Handler ParseValidators error: %v", err)
				return
			}
			*currentPV = &HeightAndValidators{
				Height:     prevHeaderWithSum.Header.Number,
				Validators: validators,
			}
			switch *currentPV {
			case phv:
				currentPV = &pphv
			case pphv:
				currentPV = &ppphv
			case ppphv:
				return
			default:
				err = fmt.Errorf("bug in bsc Handler")
				return
			}
		}

		header = &prevHeaderWithSum.Header
	}
}

func (h *Handler) getHeader(native *native.NativeService, hash ecommon.Hash, chainID uint64) (headerWithSum *HeaderWithDifficultySum, err error) {
	cache, ok := h.recentHeaders.Get(hash)
	if ok {
		headerWithChainID := cache.(*HeaderWithChainID)
		if headerWithChainID != nil && headerWithChainID.ChainID == chainID {
			headerWithSum = headerWithChainID.Header
			return
		}
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("bsc Handler getHeader error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("bsc Handler getHeader, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("bsc Handler getHeader, deserialize headerBytes from raw storage item err:%v", err)
	}
	headerWithSum = &HeaderWithDifficultySum{}
	if err := json.Unmarshal(storeBytes, &headerWithSum); err != nil {
		return nil, fmt.Errorf("bsc Handler getHeader, deserialize header error: %v", err)
	}

	h.recentHeaders.Add(hash, &HeaderWithChainID{Header: headerWithSum, ChainID: chainID})
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
)

func (h *Handler) verifyHeader(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {

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
	return h.verifyCascadingFields(native, header, ctx)
}

func (h *Handler) verifyCascadingFields(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {

	number := header.Number.Uint64()

	parent, err := h.getHeader(native, header.ParentHash, ctx.ChainID)
	if err != nil {
		return
	}

	if parent.Header.Number.Uint64() != number-1 || parent.Header.Hash() != header.ParentHash {
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
	limit := parent.Header.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		err = fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.Header.GasLimit, limit)
		return
	}

	return h.verifySeal(native, header, ctx)
}

func (h *Handler) verifySeal(native *native.NativeService, header *types.Header, ctx *Context) (signer ecommon.Address, err error) {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		err = errors.New("unknown block")
		return
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

// SyncCrossChainMsg ...
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
