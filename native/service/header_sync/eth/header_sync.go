package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

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
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
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

		////verify difficulty
		//expected := difficultyCalculator(new(big.Int).SetUint64(header.Time), prevHeader)
		//
		//if expected.Cmp(header.Difficulty) != 0 {
		//	return fmt.Errorf("SyncBlockHeader, invalid difficulty: have %v, want %v", header.Difficulty, expected)
		//}

		//block header storage
		err = putBlockHeader(native, header, v, headerParams.ChainID)
		if err != nil {
			return fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
		}
	}
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
