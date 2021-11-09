package starcoin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/params"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/starcoinorg/starcoin-go/types"
	"math/big"

	"github.com/pkg/errors"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	stc "github.com/starcoinorg/starcoin-go/client"
)

var NETURLMAP = make(map[uint64]string)

func init() {
	NETURLMAP[254] = "http://localhost:9850"
	NETURLMAP[251] = "https://barnard-seed.starcoin.org"
	NETURLMAP[252] = "https://proxima-seed.starcoin.org"
	NETURLMAP[253] = "https://halley-seed.starcoin.org"
	NETURLMAP[1] = "https://main-seed.starcoin.org/"
}

func findNetwork(chainId uint64) (string, error) {
	if url, found := NETURLMAP[chainId]; found {
		return url, nil
	} else {
		return "", fmt.Errorf("cant't found url by chainid %d", chainId)
	}
}

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
		return errors.Errorf("StarcoinHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return errors.Errorf("StarcoinHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return errors.Errorf("StarcoinHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	url, err := findNetwork(native.GetChainID())
	if err != nil {
		return errors.WithStack(err)
	}

	client := stc.NewStarcoinClient(url)
	block, err := client.GetBlockByNumber(context.Background(), 0)
	if err != nil {
		return errors.WithStack(err)
	}

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(params.ChainID)))
	if err != nil {
		return errors.Errorf("ETHHandler GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore != nil {
		return errors.Errorf("ETHHandler GetHeaderByHeight, genesis header had been initialized")
	}

	//block header storage
	blockHeader, err := block.GetHeader()
	if err != nil {
		return errors.WithStack(err)
	}
	err = putGenesisBlockHeader(native, blockHeader, params.ChainID, block.BlockHeader)
	if err != nil {
		return errors.Errorf("ETHHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	return nil
}

func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	headerParams := new(scom.SyncBlockHeaderParam)
	if err := headerParams.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return errors.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}

	//caches := NewCaches(3, native)
	for _, v := range headerParams.Headers {
		var header types.BlockHeader
		err := json.Unmarshal(v, &header)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, deserialize header err: %v", err)
		}

		headerHash, err := header.GetHash()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get header hash err: %v", err)
		}

		exist, err := IsHeaderExist(native, *headerHash, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, check header exist err: %v", err)
		}
		if exist == true {
			log.Warnf("SyncBlockHeader, header has exist. Header: %s", string(v))
			continue
		}
		// get pre header
		parentHeader, err := GetHeaderByHash(native, header.ParentHash, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the parent block failed. Error:%s, header: %s", err, string(v))
		}
		parentHeaderHash, err := parentHeader.GetHash()
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the parent block header hash failed. Error:%s, header: %s", err, string(v))
		}
		/**
		this code source refer to https://github.com/ethereum/go-ethereum/blob/master/consensus/ethash/consensus.go
		verify header need to verify:
		1. parent hash
		2. extra size
		3. current time
		*/
		//verify whether parent hash validity
		if !bytes.Equal(*parentHeaderHash, header.ParentHash) {
			return errors.Errorf("SyncBlockHeader, parent header is not right. Header: %s", string(v))
		}
		//verify whether extra size validity
		if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
			return errors.Errorf("SyncBlockHeader, SyncBlockHeader extra-data too long: %d > %d, header: %s", len(header.Extra), params.MaximumExtraDataSize, string(v))
		}
		//verify current time validity
		//if header.Timestamp > uint64(time.Now().Add(allowedFutureBlockTime).Unix()) {
		//	return errors.Errorf("SyncBlockHeader,  verify header time error:%s, checktime: %d, header: %s", consensus.ErrFutureBlock, time.Now().Add(allowedFutureBlockTime).Unix(), string(v))
		//}
		//verify whether current header time and prevent header time validity
		if header.Timestamp >= parentHeader.Timestamp {
			return errors.Errorf("SyncBlockHeader, verify header time fail. Header: %s", string(v))
		}
		// Verify that the gas limit is <= 2^63-1
		cap := uint64(0x7fffffffffffffff)
		if header.GasUsed > cap {
			return errors.Errorf("SyncBlockHeader, invalid gasuseed: have %v, max %v, header: %s", header.GasUsed, cap, string(v))
		}

		//verify difficulty
		var expected *big.Int
		expected = difficultyCalculator(new(big.Int).SetUint64(header.Timestamp), parentHeader)
		if expected.Cmp(header.GetDiffculty()) != 0 {
			return errors.Errorf("SyncBlockHeader, invalid difficulty: have %v, want %v, header: %s", header.Difficulty, expected, string(v))
		}
		// verfify header
		err = h.verifyHeader(&header)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, verify header error: %v, header: %s", err, string(v))
		}
		//block header storage
		hederDifficultySum := new(big.Int).Add(header.GetDiffculty(), parentHeader.GetDiffculty())
		err = putBlockHeader(native, header, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncGenesisHeader, put blockHeader error: %v, header: %s", err, string(v))
		}
		// get current header of main
		currentHeader, err := GetCurrentHeader(native, headerParams.ChainID)
		if err != nil {
			return errors.Errorf("SyncBlockHeader, get the current block failed. error:%s", err)
		}
		currentHeaderHash, err := currentHeader.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}
		if bytes.Equal(*currentHeaderHash, header.ParentHash) {
			appendHeader2Main(native, header.Number, *headerHash, headerParams.ChainID)
		} else {
			//
			if hederDifficultySum.Cmp(currentHeader.GetDiffculty()) > 0 {
				RestructChain(native, currentHeader, &header, headerParams.ChainID)
			}
		}
	}
	//caches.deleteCaches()
	return nil
}

func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func (h *Handler) verifyHeader(header *types.BlockHeader) error {
	return nil
}

func GetCurrentHeader(native *native.NativeService, chainID uint64) (*types.BlockHeader, error) {
	height, err := GetCurrentHeaderHeight(native, chainID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	header, err := GetHeaderByHeight(native, height, chainID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return header, nil
}

func GetCurrentHeaderHeight(native *native.NativeService, chainID uint64) (uint64, error) {
	heightStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return 0, errors.Errorf("getPrevHeaderHeight error: %v", err)
	}
	if heightStore == nil {
		return 0, errors.Errorf("getPrevHeaderHeight, heightStore is nil")
	}
	heightBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		return 0, errors.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return utils.GetBytesUint64(heightBytes), err
}

func GetHeaderByHeight(native *native.NativeService, height, chainID uint64) (*types.BlockHeader, error) {
	latestHeight, err := GetCurrentHeaderHeight(native, chainID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if height > latestHeight {
		return nil, errors.Errorf("GetHeaderByHeight, height is too big")
	}
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, errors.Errorf("GetHeaderByHeight, can not find any header records")
	}
	hashBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return GetHeaderByHash(native, hashBytes, chainID)
}

func putBlockHeader(native *native.NativeService, blockHeader types.BlockHeader, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	storeBytes, _ := json.Marshal(&blockHeader)
	headerHash, err := blockHeader.GetHash()
	if err != nil {
		return errors.WithStack(err)
	}

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), *headerHash),
		cstates.GenRawStorageItem(storeBytes))
	scom.NotifyPutHeader(native, chainID, blockHeader.Number, stc.BytesToHexString(*headerHash))
	return nil
}

func putGenesisBlockHeader(native *native.NativeService, blockHeader *types.BlockHeader, chainID uint64, jsonBlockHeader stc.BlockHeader) error {
	contract := utils.HeaderSyncContractAddress

	headerHash, err := blockHeader.GetHash()
	if err != nil {
		return errors.WithStack(err)
	}

	storeBytes, _ := json.Marshal(&jsonBlockHeader)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), *headerHash),
		cstates.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(blockHeader.Number)),
		cstates.GenRawStorageItem(*headerHash))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(blockHeader.Number)))
	scom.NotifyPutHeader(native, chainID, blockHeader.Number, stc.BytesToHexString(*headerHash))
	return nil
}

func difficultyCalculator(time *big.Int, parent *types.BlockHeader) *big.Int {
	return nil
}

func IsHeaderExist(native *native.NativeService, hash []byte, chainID uint64) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return false, errors.Errorf("IsHeaderExist, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func GetHeaderByHash(native *native.NativeService, hash []byte, chainID uint64) (*types.BlockHeader, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHash, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, errors.Errorf("GetHeaderByHash, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHash, deserialize headerBytes from raw storage item err:%v", err)
	}
	var blockHeader types.BlockHeader
	if err := json.Unmarshal(storeBytes, &blockHeader); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return &blockHeader, nil
}

func appendHeader2Main(native *native.NativeService, height uint64, txhash types.HashValue, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)),
		cstates.GenRawStorageItem(txhash))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(height)))
	scom.NotifyPutHeader(native, chainID, height, stc.BytesToHexString(txhash))
	return nil
}

func RestructChain(native *native.NativeService, current, new *types.BlockHeader, chainID uint64) error {
	si, ti := current.Number, new.Number
	var err error
	if si > ti {
		current, err = GetHeaderByHeight(native, ti, chainID)
		if err != nil {
			return errors.Errorf("RestructChain GetHeaderByHeight height:%d error:%s", ti, err)
		}
		si = ti
	}
	newHashs := make([]types.HashValue, 0)
	for ti > si {
		newHash, err := new.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}
		newHashs = append(newHashs, *newHash)
		new, err = GetHeaderByHash(native, new.ParentHash, chainID)
		if err != nil {
			return errors.Errorf("RestructChain GetHeaderByHash hash:%x error:%s", new.ParentHash, err)
		}
		ti--
	}
	for !bytes.Equal(current.ParentHash, new.ParentHash) {
		newHash, err := new.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}

		newHashs = append(newHashs, *newHash)
		new, err = GetHeaderByHash(native, new.ParentHash, chainID)
		if err != nil {
			return errors.Errorf("RestructChain GetHeaderByHash hash:%x  error:%s", new.ParentHash, err)
		}
		ti--
		si--
		current, err = GetHeaderByHeight(native, si, chainID)
		if err != nil {
			return errors.Errorf("RestructChain GetHeaderByHeight height:%d error:%s", ti, err)
		}
	}
	newHash, err := new.GetHash()
	if err != nil {
		return errors.WithStack(err)
	}
	newHashs = append(newHashs, *newHash)
	for i := len(newHashs) - 1; i >= 0; i-- {
		appendHeader2Main(native, ti, newHashs[i], chainID)
		ti++
	}
	return nil
}
