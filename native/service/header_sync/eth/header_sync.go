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
package eth

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"golang.org/x/crypto/sha3"
	"hash"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	cty "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

var (
	BIG_1             = big.NewInt(1)
	BIG_2             = big.NewInt(2)
	BIG_9             = big.NewInt(9)
	BIG_MINUS_99      = big.NewInt(-99)
	BLOCK_DIFF_FACTOR = big.NewInt(2048)
	DIFF_PERIOD       = big.NewInt(100000)
	BOMB_DELAY        = big.NewInt(8999999)
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	header, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader: %s", err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)))
	if err != nil {
		return fmt.Errorf("ETHHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return fmt.Errorf("ETHHandler GetHeaderByHeight, genesis header had been initialized")
	}

	//block header storage
	err = putGenesisBlockHeader(native, header, params.ChainID)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	return nil
}

func (this *ETHHandler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	caches := NewCaches(3, native)
	for _, v := range headerParams.Headers {
		var header cty.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}
		headerHash := header.Hash()
		exist, err := IsHeaderExist(native, headerHash.Bytes(), headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, check header exist err: %v", err)
		}
		if exist == true {
			log.Warnf("SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}
		// get pre header
		parentHeader, parentDifficultySum, err := GetHeaderByHash(native, header.ParentHash.Bytes(), headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, get the parent block failed. Error:%s, header: %s", err, string(v))
		}
		parentHeaderHash := parentHeader.Hash()
		/**
		this code source refer to https://github.com/ethereum/go-ethereum/blob/master/consensus/ethash/consensus.go
		verify header need to verify:
		1. parent hash
		2. extra size
		3. current time
		*/
		//verify whether parent hash validity
		if !bytes.Equal(parentHeaderHash.Bytes(), header.ParentHash.Bytes()) {
			return fmt.Errorf("SyncBlockHeader, parent header is not right. Header: %s", string(v))
		}
		//verify whether extra size validity
		if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
			return fmt.Errorf("SyncBlockHeader, SyncBlockHeader extra-data too long: %d > %d, header: %s", len(header.Extra), params.MaximumExtraDataSize, string(v))
		}
		//verify current time validity
		if header.Time > uint64(time.Now().Add(allowedFutureBlockTime).Unix()) {
			return fmt.Errorf("SyncBlockHeader,  verify header time error:%s, checktime: %d, header: %s", consensus.ErrFutureBlock, time.Now().Add(allowedFutureBlockTime).Unix(), string(v))
		}
		//verify whether current header time and prevent header time validity
		if header.Time <= parentHeader.Time {
			return fmt.Errorf("SyncBlockHeader, verify header time fail. Header: %s", string(v))
		}
		// Verify that the gas limit is <= 2^63-1
		cap := uint64(0x7fffffffffffffff)
		if header.GasLimit > cap {
			return fmt.Errorf("SyncBlockHeader, invalid gasLimit: have %v, max %v, header: %s", header.GasLimit, cap, string(v))
		}
		// Verify that the gasUsed is <= gasLimit
		if header.GasUsed > header.GasLimit {
			return fmt.Errorf("SyncBlockHeader, invalid gasUsed: have %d, gasLimit %d, header: %s", header.GasUsed, header.GasLimit, string(v))
		}
		// GasLimit adjustment range 0.0976%（=1/1024 ）
		diff := int64(parentHeader.GasLimit) - int64(header.GasLimit)
		if diff < 0 {
			diff *= -1
		}
		limit := parentHeader.GasLimit / params.GasLimitBoundDivisor
		if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
			return fmt.Errorf("SyncBlockHeader, invalid gas limit: have %d, want %d += %d, header: %s", header.GasLimit, parentHeader.GasLimit, limit, string(v))
		}
		//verify difficulty
		expected := difficultyCalculator(new(big.Int).SetUint64(header.Time), parentHeader)
		if expected.Cmp(header.Difficulty) != 0 {
			return fmt.Errorf("SyncBlockHeader, invalid difficulty: have %v, want %v, header: %s", header.Difficulty, expected, string(v))
		}
		// verfify header
		err = this.verifyHeader(&header, caches)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, verify header error: %v, header: %s", err, string(v))
		}
		//block header storage
		hederDifficultySum := new(big.Int).Add(header.Difficulty, parentDifficultySum)
		err = putBlockHeader(native, header, hederDifficultySum, headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v, header: %s", err, string(v))
		}
		// get current header of main
		currentHeader, currentDifficultySum, err := GetCurrentHeader(native, headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, get the current block failed. error:%s", err)
		}
		if bytes.Equal(currentHeader.Hash().Bytes(), header.ParentHash.Bytes()) {
			appendHeader2Main(native, header.Number.Uint64(), headerHash, headerParams.ChainID)
		} else {
			//
			if hederDifficultySum.Cmp(currentDifficultySum) > 0 {
				RestructChain(native, currentHeader, &header, headerParams.ChainID)
			}
		}
	}
	caches.deleteCaches()
	return nil
}

func (this *ETHHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func getGenesisHeader(input []byte) (cty.Header, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return cty.Header{}, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var header cty.Header
	err := json.Unmarshal(params.GenesisHeader, &header)
	if err != nil {
		return cty.Header{}, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}
	return header, nil
}

func difficultyCalculator(time *big.Int, parent *types.Header) *big.Int {
	// https://github.com/ethereum/EIPs/issues/100.
	// algorithm:
	// diff = (parent_diff +
	//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
	//        ) + 2^(periodCount / 10000 - 2)
	// diff = parent_diff + diff_adjust + diff_bomb

	//difficulty adjustment
	//(parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
	x := new(big.Int).Sub(time, new(big.Int).SetUint64(parent.Time))
	x.Div(x, BIG_9)

	if parent.UncleHash == types.EmptyUncleHash {
		x.Sub(BIG_1, x)
	} else {
		x.Sub(BIG_2, x)
	}

	if x.Cmp(BIG_MINUS_99) < 0 {
		x.Set(BIG_MINUS_99)
	}

	y := new(big.Int).Div(parent.Difficulty, BLOCK_DIFF_FACTOR)
	x.Mul(y, x)
	x.Add(parent.Difficulty, x)

	if x.Cmp(params.MinimumDifficulty) < 0 {
		x.Set(params.MinimumDifficulty)
	}

	//difficulty bomb
	//https://eips.ethereum.org/EIPS/eip-1234
	fakeBlockNumber := new(big.Int)
	if parent.Number.Cmp(BOMB_DELAY) >= 0 {
		fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, BOMB_DELAY)
	}

	periodCount := fakeBlockNumber
	periodCount.Div(periodCount, DIFF_PERIOD)

	if periodCount.Cmp(BIG_1) > 0 {
		y.Sub(periodCount, BIG_2)
		y.Exp(BIG_2, y, nil)
		x.Add(x, y)
	}
	return x
}

func (this *ETHHandler) verifyHeader(header *cty.Header, caches *Caches) error {
	// try to verfify header
	number := header.Number.Uint64()
	size := datasetSize(number)
	headerHash := HashHeader(header).Bytes()
	nonce := header.Nonce.Uint64()
	// get seed and seed head
	seed := make([]byte, 40)
	copy(seed, headerHash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)
	seed = crypto.Keccak512(seed)
	// get mix
	mix := make([]uint32, mixBytes/4)
	for i := 0; i < len(mix); i++ {
		mix[i] = binary.LittleEndian.Uint32(seed[i%16*4:])
	}
	// get cache
	cache := caches.getCache(number)
	if len(cache) <= 1 {
		return fmt.Errorf("cache of proof-of-work is not generated!")
	}
	// get new mix with DAG data
	rows := uint32(size / mixBytes)
	temp := make([]uint32, len(mix))
	seedHead := binary.LittleEndian.Uint32(seed)
	for i := 0; i < loopAccesses; i++ {
		parent := fnv(uint32(i)^seedHead, mix[i%len(mix)]) % rows
		for j := uint32(0); j < mixBytes/hashBytes; j++ {
			xx := lookup(cache, 2*parent+j)
			copy(temp[j*hashWords:], xx)
		}
		fnvHash(mix, temp)
	}
	// get new mix by compress
	for i := 0; i < len(mix); i += 4 {
		mix[i/4] = fnv(fnv(fnv(mix[i], mix[i+1]), mix[i+2]), mix[i+3])
	}
	mix = mix[:len(mix)/4]
	// get digest by compressed mix
	digest := make([]byte, ethcommon.HashLength)
	for i, val := range mix {
		binary.LittleEndian.PutUint32(digest[i*4:], val)
	}
	// get header result hash
	result := crypto.Keccak256(append(seed, digest...))
	// Verify the calculated digest against the ones provided in the header
	if !bytes.Equal(header.MixDigest[:], digest) {
		return fmt.Errorf("invalid mix digest!")
	}
	// compare result hash with target hash
	target := new(big.Int).Div(two256, header.Difficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return fmt.Errorf("invalid proof-of-work!")
	}
	return nil
}

func HashHeader(header *types.Header) (hash ethcommon.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	rlp.Encode(hasher, []interface{}{
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
		header.Extra,
	})
	hasher.Sum(hash[:0])
	return hash
}

type hasher func(dest []byte, data []byte)

func makeHasher(h hash.Hash) hasher {
	// sha3.state supports Read to get the sum, use it to avoid the overhead of Sum.
	// Read alters the state but we reset the hash before every operation.
	type readerHash interface {
		hash.Hash
		Read([]byte) (int, error)
	}
	rh, ok := h.(readerHash)
	if !ok {
		panic("can't find Read method on hash")
	}
	outputLen := rh.Size()
	return func(dest []byte, data []byte) {
		rh.Reset()
		rh.Write(data)
		rh.Read(dest[:outputLen])
	}
}

func seedHash(block uint64) []byte {
	seed := make([]byte, 32)
	if block < epochLength {
		return seed
	}
	keccak256 := makeHasher(sha3.NewLegacyKeccak256())
	for i := 0; i < int(block/epochLength); i++ {
		keccak256(seed, seed)
	}
	return seed
}
