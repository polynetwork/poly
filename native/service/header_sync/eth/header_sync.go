package eth

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"hash"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	cty "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

var (
	BIG_1             = big.NewInt(1)
	BIG_2             = big.NewInt(2)
	BIG_9             = big.NewInt(9)
	BIG_MINUS_99      = big.NewInt(-99)
	BLOCK_DIFF_FACTOR = big.NewInt(2048)
	DIFF_PERIOD       = big.NewInt(10000)
	BOMB_DELAY        = big.NewInt(5000001)
)

type ETHHandler struct {
	caches *Caches
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{
		caches: NewCaches(3),
	}
}

func (this *ETHHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	//// get operator from database
	//operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	//if err != nil {
	//	return err
	//}

	////check witness
	//err = utils.ValidateOwner(native, operatorAddress)
	//if err != nil {
	//	return fmt.Errorf("ETHHandler SyncGenesisHeader, checkWitness error: %v", err)
	//}

	header, headerByte, err := getGenesisHeader(native.GetInput())
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader: %s", err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(params.ChainID), header.Number.Bytes()))
	if err != nil {
		return fmt.Errorf("ETHHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return fmt.Errorf("ETHHandler GetHeaderByHeight, genesis header had been initialized")
	}

	//block header storage
	err = putBlockHeader(native, header, headerByte, params.ChainID)
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
	for _, v := range headerParams.Headers {
		var header cty.Header
		err := json.Unmarshal(v, &header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}
		prevHeader, err := GetHeaderByHeight(native, header.Number.Uint64()-1, headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, height:%d, error:%s", header.Number.Uint64()-1, err)
		}
		/**
		this code source refer to https://github.com/ethereum/go-ethereum/blob/master/consensus/ethash/consensus.go
		verify header need to verify:
		1. parent hash
		2. extra size
		3. current time
		*/
		//verify whether parent hash validity
		if !bytes.Equal(prevHeader.Hash().Bytes(), header.ParentHash.Bytes()) {
			return fmt.Errorf("SyncBlockHeader, current header height:%d, parentHash:%x, prevent header height:%d, hash:%x", header.Number, header.ParentHash, prevHeader.Number, prevHeader.Hash())
		}
		//verify whether extra size validity
		if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
			return fmt.Errorf("SyncBlockHeader, SyncBlockHeader extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
		}
		//verify current time validity
		if header.Time > uint64(time.Now().Add(allowedFutureBlockTime).Unix()) {
			return fmt.Errorf("SyncBlockHeader,  verify header time error:%s", consensus.ErrFutureBlock)
		}
		//verify whether current header time and prevent header time validity
		if header.Time <= prevHeader.Time {
			return fmt.Errorf("SyncBlockHeader, verify header time fail, current header time:%d, prevent header time:%d", header.Time, prevHeader.Time)
		}

		// Verify that the gas limit is <= 2^63-1
		cap := uint64(0x7fffffffffffffff)
		if header.GasLimit > cap {
			return fmt.Errorf("SyncBlockHeader, invalid gasLimit: have %v, max %v", header.GasLimit, cap)
		}
		// Verify that the gasUsed is <= gasLimit
		if header.GasUsed > header.GasLimit {
			return fmt.Errorf("SyncBlockHeader, invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
		}

		// GasLimit adjustment range 0.0976%（=1/1024 ）
		diff := int64(prevHeader.GasLimit) - int64(header.GasLimit)
		if diff < 0 {
			diff *= -1
		}
		limit := prevHeader.GasLimit / params.GasLimitBoundDivisor
		if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
			return fmt.Errorf("SyncBlockHeader, invalid gas limit: have %d, want %d += %d", header.GasLimit, prevHeader.GasLimit, limit)
		}

		//verify difficulty
		expected := difficultyCalculator(new(big.Int).SetUint64(header.Time), prevHeader)
		if expected.Cmp(header.Difficulty) != 0 {
			return fmt.Errorf("SyncBlockHeader, invalid difficulty: have %v, want %v", header.Difficulty, expected)
		}

		//
		err = this.verifyHeader(&header)
		if err != nil {
			return err
		}

		//block header storage
		err = putBlockHeader(native, header, v, headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
		}
	}
	return nil
}

func (this *ETHHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func getGenesisHeader(input []byte) (cty.Header, []byte, error) {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(input)); err != nil {
		return cty.Header{}, nil, fmt.Errorf("getGenesisHeader, contract params deserialize error: %v", err)
	}
	var header cty.Header
	err := json.Unmarshal(params.GenesisHeader, &header)
	if err != nil {
		return cty.Header{}, nil, fmt.Errorf("getGenesisHeader, deserialize header err: %v", err)
	}
	return header, params.GenesisHeader, nil
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
	x.Mul(x, y)

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

func (this *ETHHandler) verifyHeader(header *cty.Header) error {
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
	cache := this.caches.getCache(number)
	if len(cache) <= 0 {
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
