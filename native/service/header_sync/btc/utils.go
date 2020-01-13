package btc

import (
	"encoding/hex"
	"time"
	"github.com/ontio/multi-chain/common/config"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ontio/multi-chain/common/log"
	"math/big"
	"github.com/ontio/multi-chain/common"
)

var OrphanHeaderError = errors.New("header does not extend any known headers")


const (

	targetTimespan      = time.Hour * 24 * 14
	targetSpacing       = time.Minute * 10
	epochLength         = int32(targetTimespan / targetSpacing) // 2016
	maxDiffAdjust       = 4
	minRetargetTimespan = int64(targetTimespan / maxDiffAdjust)
	maxRetargetTimespan = int64(targetTimespan * maxDiffAdjust)
)

var netParam = &chaincfg.TestNet3Params
//var netParam = &chaincfg.RegressionNetParams

func putGenesisBlockHeader(native *native.NativeService, chainID uint64, blockHeader StoredHeader) {
	contract := utils.HeaderSyncContractAddress
	blockHash := blockHeader.Header.BlockHash()
	blockHeight := uint64(blockHeader.Height)

	sink := new(common.ZeroCopySink)
	blockHeader.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(sink.Bytes()))

	putBlockHash(native, chainID, blockHeight, blockHash)

	putBlockHeader(native, chainID, blockHeader)

	putBestBlockHeader(native, chainID, blockHeader)

	notifyPutHeader(native, chainID, blockHeight, hex.EncodeToString(blockHash.CloneBytes()))
}

func notifyPutHeader(native *native.NativeService, chainID uint64, height uint64, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, blockHash, native.GetHeight()},
		})
}


func putBlockHash(native *native.NativeService, chainID uint64, height uint64, hash chainhash.Hash) {
	contract := utils.HeaderSyncContractAddress

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte("HeightToBlockHash"), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)),
		cstates.GenRawStorageItem(hash.CloneBytes()))
}

func GetBlockHash(native *native.NativeService, chainID uint64, height uint64) (*chainhash.Hash, error) {
	contract := utils.HeaderSyncContractAddress

	hashStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte("HeightToBlockHash"), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return nil, fmt.Errorf("GetBlockHash, get heightBlockHashStore error: %v", err)
	}
	if hashStore == nil {
		return nil, fmt.Errorf("GetBlockHash, can not find any index records")
	}
	hashBs, err := cstates.GetValueFromRawStorageItem(hashStore)
	if err != nil {
		return nil, fmt.Errorf("GetBlockHash, deserialize blockHashBytes from raw storage item err:%v", err)
	}

	hash := new(chainhash.Hash)
	err = hash.SetBytes(hashBs)
	if err != nil {
		return nil, fmt.Errorf("GetBlockHash at height = %d, error:%v", height, err)
	}
	return hash, nil
}


func putBlockHeader(native *native.NativeService, chainID uint64, sh StoredHeader) {
	contract := utils.HeaderSyncContractAddress

	blockHash := sh.Header.BlockHash()
	sink := new(common.ZeroCopySink)
	sh.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte("HashToStoredHeader"), utils.GetUint64Bytes(chainID), blockHash.CloneBytes()),
		cstates.GenRawStorageItem(sink.Bytes()))
}


func GetBlockHeader(native *native.NativeService, chainID uint64, hash chainhash.Hash) (*StoredHeader, error) {
	contract := utils.HeaderSyncContractAddress

	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte("HashToStoredHeader"), utils.GetUint64Bytes(chainID), hash.CloneBytes()))
	if err != nil {
		return nil, fmt.Errorf("GetBlockHeader, get hashBlockHeaderStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetBlockHeader, can not find any index records")
	}
	shBs, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetBlockHeader, deserialize blockHashBytes from raw storage item err:%v", err)
	}

	sh := new(StoredHeader)
	if err := sh.Deserialization(common.NewZeroCopySource(shBs)); err != nil {
		return nil, fmt.Errorf("GetStoredHeader, deserializeHeader error:%v", err)
	}

	return sh, nil
}



func putBestBlockHeader(native *native.NativeService, chainID uint64, bestHeader StoredHeader) {
	contract := utils.HeaderSyncContractAddress

	sink := new(common.ZeroCopySink)
	bestHeader.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte("BestBlockHeader"), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(sink.Bytes()))
}

func GetBestBlockHeader(native *native.NativeService, chainID uint64) (*StoredHeader, error) {
	contract := utils.HeaderSyncContractAddress

	bestBlockHeaderStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte("BestBlockHeader"), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return nil, fmt.Errorf("GetBestBlockHeader, get BestBlockHeader error: %v", err)
	}
	if bestBlockHeaderStore == nil {
		return nil, fmt.Errorf("GetBestBlockHeader, can not find any index records")
	}
	bestBlockHeaderBs, err := cstates.GetValueFromRawStorageItem(bestBlockHeaderStore)
	if err != nil {
		return nil, fmt.Errorf("GetBestBlockHeader, deserialize bestBlockHeaderBytes from raw storage item err:%v", err)
	}
	bestBlockHeader := new(StoredHeader)
	err = bestBlockHeader.Deserialization(common.NewZeroCopySource(bestBlockHeaderBs))
	if err != nil {
		return nil, fmt.Errorf("GetBestBlockHeader, deserialize storedHeader error:%v", err)
	}
	return bestBlockHeader, nil
}



func GetPreviousHeader(native *native.NativeService, chainID uint64, header wire.BlockHeader) (*StoredHeader, error) {
	hash := header.PrevBlock
	return GetBlockHeader(native, chainID, hash)
}


func CheckHeader(native *native.NativeService, chainID uint64, header wire.BlockHeader, prevHeader *StoredHeader) (bool, error) {
	// Get hash of n-1 header
	prevHash := prevHeader.Header.BlockHash()
	height := prevHeader.Height

	// Check if headers link together.  That whole 'blockchain' thing.
	if !prevHash.IsEqual(&header.PrevBlock) {
		return false, fmt.Errorf("Headers %d and %d don't link.\n", height, height+1)
	}

	// Check the header meets the difficulty requirement
	diffTarget, err := calcRequiredWork(native, chainID, header, int32(height+1), prevHeader)
	if err != nil {
		log.Errorf("calcRequiredWork error is %v", err)
		return false, fmt.Errorf("Error calclating difficulty")
	}
	if header.Bits != diffTarget {
		return false, fmt.Errorf("Block %d %s incorrect difficulty.  Read %d, expect %d\n",
			height+1, header.BlockHash().String(), header.Bits, diffTarget)
	}

	// Check if there's a valid proof of work.  That whole "Bitcoin" thing.
	if !checkProofOfWork(header, netParam) {
		log.Debugf("Block %d bad proof of work.\n", height+1)
		return false, nil
	}

	return true, nil// it must have worked if there's no errors and got to the end.
}


// Get the PoW target this block should meet. We may need to handle a difficulty adjustment
// or testnet difficulty rules.
func calcRequiredWork(native *native.NativeService, chainID uint64, header wire.BlockHeader, height int32, prevHeader *StoredHeader) (uint32, error) {
	// If this is not a difficulty adjustment period
	if height%epochLength != 0 {
		// If we are on testnet
		if netParam.ReduceMinDifficulty {
			// If it's been more than 20 minutes since the last header return the minimum difficulty
			if header.Timestamp.After(prevHeader.Header.Timestamp.Add(targetSpacing * 2)) {
				return netParam.PowLimitBits, nil
			} else { // Otherwise return the difficulty of the last block not using special difficulty rules
				for {
					var err error = nil
					for err == nil && int32(prevHeader.Height)%epochLength != 0 && prevHeader.Header.Bits == netParam.PowLimitBits{
						var sh *StoredHeader
						sh, err = GetPreviousHeader(native, chainID, prevHeader.Header)
						// Error should only be non-nil if prevHeader is the checkpoint.
						// In that case we should just return checkpoint bits
						if err == nil {
							prevHeader = sh
						}

					}
					return prevHeader.Header.Bits, nil
				}
			}
		}
		// Just return the bits from the last header
		return prevHeader.Header.Bits, nil
	}
	// We are on a difficulty adjustment period so we need to correctly calculate the new difficulty.
	epoch, err := GetEpoch(native, chainID)
	if err != nil {
		return 0, err
	}
	return calcDiffAdjust(*epoch, prevHeader.Header, netParam), nil
}

func GetEpoch(native *native.NativeService, chainID uint64) (*wire.BlockHeader, error) {
	sh, err := GetBestBlockHeader(native, chainID)
	if err != nil {
		return &sh.Header, err
	}
	for i := 0; i < 2015; i++ {
		sh, err = GetPreviousHeader(native, chainID, sh.Header)
		if err != nil {
			return &sh.Header, err
		}
	}
	log.Debug("Epoch", sh.Header.BlockHash().String())
	return &sh.Header, nil
}

func GetCommonAncestor(native *native.NativeService, chainID uint64, bestHeader, prevBestHeader *StoredHeader) (*StoredHeader, error) {
	var err error
	rollback := func(parent *StoredHeader, n int) (*StoredHeader, error) {
		for i := 0; i < n; i++ {
			parent, err = GetPreviousHeader(native, chainID, parent.Header)
			if err != nil {
				return parent, err
			}
		}
		return parent, nil
	}

	majority := bestHeader
	minority := prevBestHeader
	if bestHeader.Height > prevBestHeader.Height {
		majority, err = rollback(majority, int(bestHeader.Height-prevBestHeader.Height))
		if err != nil {
			return nil, err
		}
	} else if prevBestHeader.Height > bestHeader.Height {
		minority, err = rollback(minority, int(prevBestHeader.Height-bestHeader.Height))
		if err != nil {
			return nil, err
		}
	}

	for {
		majorityHash := majority.Header.BlockHash()
		minorityHash := minority.Header.BlockHash()
		if majorityHash.IsEqual(&minorityHash) {
			return majority, nil
		}
		majority, err = GetPreviousHeader(native, chainID, majority.Header)
		if err != nil {
			return nil, err
		}
		minority, err = GetPreviousHeader(native, chainID, minority.Header)
		if err != nil {
			return nil, err
		}
	}
}










func ReIndexHeaderHeight(native *native.NativeService, chainID uint64, commonAncestorHeight,  bestHeaderHeight uint64, newBlock *StoredHeader) error {
	contract := utils.HeaderSyncContractAddress

	for i := bestHeaderHeight; i > commonAncestorHeight; i-- {
		native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte("HeightToBlockHash"), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(i)))
	}

	best := newBlock
	for best, _ = GetPreviousHeader(native, chainID, best.Header); uint64(best.Height) > commonAncestorHeight; best, _ = GetPreviousHeader(native, chainID, best.Header) {
		value := best.Header.BlockHash()
		putBlockHash(native, chainID, uint64(best.Height), value)
	}

	return nil
}





// Verifies the header hashes into something lower than specified by the 4-byte bits field.
func checkProofOfWork(header wire.BlockHeader, p *chaincfg.Params) bool {
	target := CompactToBig(header.Bits)

	// The target must more than 0.  Why can you even encode negative...
	if target.Sign() <= 0 {
		log.Debugf("Block target %064x is neagtive(??)\n", target.Bytes())
		return false
	}
	// The target must be less than the maximum allowed (difficulty 1)
	if target.Cmp(p.PowLimit) > 0 {
		log.Debugf("Block target %064x is "+
			"higher than max of %064x", target, p.PowLimit.Bytes())
		return false
	}
	// The header hash must be less than the claimed target in the header.
	blockHash := header.BlockHash()
	hashNum := HashToBig(&blockHash)
	if hashNum.Cmp(target) > 0 {
		log.Debugf("Block hash %064x is higher than "+
			"required target of %064x", hashNum, target)
		return false
	}
	return true
}

// This function takes in a start and end block header and uses the timestamps in each
// to calculate how much of a difficulty adjustment is needed. It returns a new compact
// difficulty target.
func calcDiffAdjust(start, end wire.BlockHeader, p *chaincfg.Params) uint32 {
	duration := end.Timestamp.UnixNano() - start.Timestamp.UnixNano()
	fmt.Println("duration..............", duration)
	if duration < minRetargetTimespan {
		log.Debugf("Whoa there, block %s off-scale high 4X diff adjustment!",
			end.BlockHash().String())
		duration = minRetargetTimespan
	} else if duration > maxRetargetTimespan {
		log.Debugf("Uh-oh! block %s off-scale low 0.25X diff adjustment!\n",
			end.BlockHash().String())
		duration = maxRetargetTimespan
	}

	// calculation of new 32-byte difficulty target
	// first turn the previous target into a big int
	prevTarget := CompactToBig(end.Bits)
	// new target is old * duration...
	newTarget := new(big.Int).Mul(prevTarget, big.NewInt(duration))
	// divided by 2 weeks
	newTarget.Div(newTarget, big.NewInt(int64(targetTimespan)))

	// clip again if above minimum target (too easy)
	if newTarget.Cmp(p.PowLimit) > 0 {
		newTarget.Set(p.PowLimit)
	}

	return BigToCompact(newTarget)
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
//	* the most significant 8 bits represent the unsigned base 256 exponent
// 	* bit 23 (the 24th bit) represents the sign bit
//	* the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
// 	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used in bitcoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with bitcoind.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}



// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}


// HashToBig converts a chainhash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *chainhash.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *hash
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}



// CalcWork calculates a work value from difficulty bits.  Bitcoin increases
// the difficulty for generating a block by decreasing the value which the
// generated hash must be less than.  This difficulty target is stored in each
// block header using a compact representation as described in the documentation
// for CompactToBig.  The main chain is selected by choosing the chain that has
// the most proof of work (highest difficulty).  Since a lower target difficulty
// value equates to higher actual difficulty, the work value which will be
// accumulated must be the inverse of the difficulty.  Also, in order to avoid
// potential division by zero and really small floating point numbers, the
// result adds 1 to the denominator and multiplies the numerator by 2^256.
func CalcWork(bits uint32) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, BIG_1)
	return new(big.Int).Div(ONE_LSH_256, denominator)
}