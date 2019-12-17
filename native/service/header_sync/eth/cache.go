package eth

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"golang.org/x/crypto/sha3"
	"math"
	"reflect"
	"unsafe"
)

type Caches struct {
	items map[uint64][]uint32
	cap int
}

func NewCaches(size int) *Caches {
	caches := &Caches {
		cap : size,
		items : make(map[uint64][]uint32),
	}
	return caches
}

func (self *Caches) tryCache (epoch uint64) ([]uint32, []uint32) {
	current := self.items[epoch]
	future := self.items[epoch + 1]
	return current, future
}

func (self *Caches) addCache(epoch uint64, item []uint32) {
	if len(self.items) < self.cap {
		self.items[epoch] = item
		return
	}
	var min uint64 = math.MaxUint64
	for key,_ := range self.items {
		if key < min {
			min = key
		}
	}
	delete(self.items, min)
	self.items[epoch] = item
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
		self.generateCache(cache, seed)
		self.addCache(epoch, cache)
		return cache
	}
	if future == nil {
		go func(newepoch uint64) {
			size := cacheSize(newepoch*epochLength + 1)
			seed := seedHash(newepoch*epochLength + 1)
			// If we don't store anything on disk, generate and return.
			cache := make([]uint32, size/4)
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
