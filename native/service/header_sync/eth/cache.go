package eth

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
	"golang.org/x/crypto/sha3"
	"reflect"
	"unsafe"
)

type Caches struct {
	native *native.NativeService
	cap   int
	items map[uint64][]uint32
}

func NewCaches(size int, native *native.NativeService) *Caches {
	caches := &Caches{
		cap:   size,
		native: native,
		items : make(map[uint64][]uint32),

	}
	return caches
}

func (self *Caches) serialize(values []uint32) []byte {
	buf := make([]byte, len(values) * 4)
	for i, value := range values {
		binary.LittleEndian.PutUint32(buf[i*4:], value)
	}
	return buf
}

func (self *Caches) deserialize(buf []byte) []uint32 {
	values := make([]uint32, len(buf) / 4)
	for i := 0; i < len(values); i++ {
		values[i] = binary.LittleEndian.Uint32(buf[i*4:])
	}
	return values
}

func (self *Caches) tryCache(epoch uint64) ([]uint32, []uint32) {
	contract := utils.HeaderSyncContractAddress
	current := self.items[epoch]
	future := self.items[epoch + 1]
	if current == nil {
		currentStorge, err := self.native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(common.ETH_CACHE), utils.GetUint64Bytes(epoch)))
		if currentStorge == nil || err != nil {
			current = nil
		} else {
			current1, err := states.GetValueFromRawStorageItem(currentStorge)
			if err != nil {
				current = nil
			} else {
				current = self.deserialize(current1)
				self.items[epoch] = current
			}
		}
	}
	if future == nil {
		futureStorge, err := self.native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(common.ETH_CACHE), utils.GetUint64Bytes(epoch+1)))
		if futureStorge == nil || err != nil {
			future = nil
		} else {
			future1, err := states.GetValueFromRawStorageItem(futureStorge)
			if err != nil {
				future = nil
			} else {
				future = self.deserialize(future1)
				self.items[epoch + 1] = future
			}
		}
	}
	return current, future
}

func (self *Caches) addCache(epoch uint64, items []uint32) {
	contract := utils.HeaderSyncContractAddress
	self.native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(common.ETH_CACHE), utils.GetUint64Bytes(epoch)), states.GenRawStorageItem(self.serialize(items)))
}

func (self *Caches) getCache(block uint64) []uint32 {
	epoch := block / epochLength
	current, future := self.tryCache(epoch)
	if current != nil && future != nil {
		return current
	}
	if current == nil {
		size := cacheSize(epoch*epochLength + 1)
		seed := seedHash(epoch*epochLength + 1)
		// If we don't store anything on disk, generate and return.
		cache := make([]uint32, size/4)
		log.Infof("current generate cache...... epoch: %d--------------------------", epoch)
		self.generateCache(cache, seed)
		self.addCache(epoch, cache)
		return cache
	}
	if future == nil {
		self.addCache(epoch+1, []uint32{1})
		go func(newepoch uint64) {
			size := cacheSize(newepoch*epochLength + 1)
			seed := seedHash(newepoch*epochLength + 1)
			// If we don't store anything on disk, generate and return.
			cache := make([]uint32, size/4)
			log.Infof("future generate cache......epoch: %d--------------------------", newepoch)
			self.generateCache(cache, seed)
			self.addCache(newepoch, cache)
		}(epoch + 1)
	}
	return current
}

func (self *Caches) generateCache(dest []uint32, seed []byte) {
	// Convert our destination slice to a byte buffer
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&dest))
	header.Len *= 4
	header.Cap *= 4
	cache := *(*[]byte)(unsafe.Pointer(&header))

	// Calculate the number of theoretical rows (we'll store in one buffer nonetheless)
	size := uint64(len(cache))
	rows := int(size) / hashBytes

	// Create a hasher to reuse between invocations
	keccak512 := makeHasher(sha3.NewLegacyKeccak512())

	// Sequentially produce the initial dataset
	keccak512(cache, seed)
	for offset := uint64(hashBytes); offset < size; offset += hashBytes {
		keccak512(cache[offset:], cache[offset-hashBytes:offset])
	}
	// Use a low-round version of randmemohash
	temp := make([]byte, hashBytes)
	for i := 0; i < cacheRounds; i++ {
		for j := 0; j < rows; j++ {
			var (
				srcOff = ((j - 1 + rows) % rows) * hashBytes
				dstOff = j * hashBytes
				xorOff = (binary.LittleEndian.Uint32(cache[dstOff:]) % uint32(rows)) * hashBytes
			)
			bitutil.XORBytes(temp, cache[srcOff:srcOff+hashBytes], cache[xorOff:xorOff+hashBytes])
			keccak512(cache[dstOff:], temp)
		}
	}
	// Swap the byte order on big endian systems and return
	if !isLittleEndian() {
		swap(cache)
	}
}

func lookup(cache []uint32, index uint32) []uint32 {
	keccak512 := makeHasher(sha3.NewLegacyKeccak512())
	rawData := generateDatasetItem(cache, index, keccak512)
	data := make([]uint32, len(rawData)/4)
	for i := 0; i < len(data); i++ {
		data[i] = binary.LittleEndian.Uint32(rawData[i*4:])
	}
	return data
}

func generateDatasetItem(cache []uint32, index uint32, keccak512 hasher) []byte {
	// Calculate the number of theoretical rows (we use one buffer nonetheless)
	rows := uint32(len(cache) / hashWords)
	// Initialize the mix
	mix := make([]byte, hashBytes)
	binary.LittleEndian.PutUint32(mix, cache[(index%rows)*hashWords]^index)
	for i := 1; i < hashWords; i++ {
		binary.LittleEndian.PutUint32(mix[i*4:], cache[(index%rows)*hashWords+uint32(i)])
	}
	keccak512(mix, mix)
	// Convert the mix to uint32s to avoid constant bit shifting
	intMix := make([]uint32, hashWords)
	for i := 0; i < len(intMix); i++ {
		intMix[i] = binary.LittleEndian.Uint32(mix[i*4:])
	}
	// fnv it with a lot of random cache nodes based on index
	for i := uint32(0); i < datasetParents; i++ {
		parent := fnv(index^i, intMix[i%16]) % rows
		fnvHash(intMix, cache[parent*hashWords:])
	}
	// Flatten the uint32 mix into a binary one and return
	for i, val := range intMix {
		binary.LittleEndian.PutUint32(mix[i*4:], val)
	}
	keccak512(mix, mix)
	return mix
}
