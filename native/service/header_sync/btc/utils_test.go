package btc

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/storage"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

var (
	// New chain starting from regtest genesis
	chain = []string{
		"0000002006226e46111a0b59caaf126043eb5bbf28c34f3a5e332a1fc7b2b73cf188910fc3ed4523bf94fc1fa184bee85af604c9ebeea6b39b498f62703fd3f03e7475534658d158ffff7f2001000000",
		"000000207c3d2d417ff34a46f4f11a972d8e32bc98b300112dd4d9a1dae9ff87468eae136b90f1757adfab2056d693160b417b8f87a65c2c0735a47e63768f26473905506059d158ffff7f2003000000",
		"000000200c6ea2eaf928b2d5d080c2f36dac1185865db1289c7339834b98e8034e4274073ed977491ebe6f9c0e01f5796e36ed66bf4e410bbbc2635129d6e0ecfc1897908459d158ffff7f2001000000",
		"000000202e1569563ff6463f65bb7669b35fb9dd95ba0b251e30251b9877d9578b8700680337ff38b71d9667190c99e8fae337ba8c9c40cbd2c4678ba71d81cf6d3a1aa2ac59d158ffff7f2001000000",
		"000000204525edcccf706e3769a54c8772934f291d6810315a26c177862c66feb9f3896e090c84be811cfdfed6da043cb337fccecff95fc73810ca82adb3d032b5d49140c759d158ffff7f2000000000",
		"00000020ada1a9efa81df10d7b430e2fd5f3b085180c91b0e9b0f6e9af2d9b733544015eab404ef503e538909a04a419499133af9bcee47fcfc84baaab5344f77ebd455dec59d158ffff7f2000000000",
		"000000204fdcb9ca4cc47ae7485bfc2f8adcbd515b1ee0cb724d343c91f02b6ec5a0ba507dddd2639fc1bd522489a2c2f2b681a60c6c7939490458dc1c008f3217cb47d6035ad158ffff7f2001000000",
		"0000002019dbc9a6cec93be207053e4dfbc63af20c3cedba68f890c5a90f27aeb2ecc73386692b64e16ea4b87fc877cb3762394d12b597a0ca8d5efb2ea2c6e163f9e4c8225ad158ffff7f2000000000",
		"000000203afc4a1c100fe3e21fa24ef92857613bb00890564e3529623780bc8d4a86d15cfd35aef39950dc53c348b5013f4ee3d94afc16745d6b3c8a9e6acfb8a2641c6f3e5ad158ffff7f2000000000",
		"000000200e1b58feab56f9fe5ed7484a8c7bfecdb270da528db7a805d18208891bde3726a5ccb0a073d0cc7402ac89f4bb4b64c39bc365bfee7ccd7ea3a24996ee684c775a5ad158ffff7f2000000000",
	}
	// Forks `chain` starting at block 6
	fork = []string{
		"00000020ada1a9efa81df10d7b430e2fd5f3b085180c91b0e9b0f6e9af2d9b733544015eead915a2f4521c58cb1c42a469aefede5a9d1dddfe8ccc408f8135fc2560f25a096dd158ffff7f20e9aace03",
		"0000002097e3603b40c0c7add951e3a7dba5088836d17e1123ef7cffdd60174e3dce0024cffe0c74189d854a778a3e57fee8510103e83d95b221b8bfe1159806b3bde27e236dd158ffff7f20794caff6",
		"0000002085a3bf0898ed1cad9e868120c8e044673425a13ecc7ab2daec204ca9190e643ca32434566054789e79214a7cb7c1b6e37084cbfce7564d4aabb10ef6fc1d655c3d6dd158ffff7f20c2e4cb6f",
		"000000209aa626e76fbcfc08bc1626a0a9bc7b82d8521de22a477e7b377d8f83be8d446a05aae352ffe9f09af1d79d24992dbee2785b3fe4eb4a0e21e7a3b26a90115dac536dd158ffff7f201d2f76eb",
		"000000208d6d636589b4056d1486fbcc0b46adefbb770b7e6a8d668fe65c3f58f5c2c70934008f98664ffec01f583870f843b617c869ec30f1b37723b3d0f0d4a3ba6a88686dd158ffff7f209d12ee06",
		"0000002067cf05afedc2b5956c10845006358fe480893e1199a0c0e2b70d5ecf2787af760385ca3d191d1800cd7b6a56d8b44853109f3e5983a94c7e10818541278ec6027b6dd158ffff7f2004e2c75c",
		"00000020b2227c6c858a36af167d9667dcf4f58df604ab7962a660d69d233a63e7269f06ecb669fff090b7f2f6952d52c96ca0c8abe1e266d9740f8548eeb10eea9e3536906dd158ffff7f20c0ac3d1e",
	}
)

func TestGetEpoch(t *testing.T) {
	cacheDB, err := syncGenesisHeader(&chaincfg.RegressionNetParams.GenesisBlock.Header)
	assert.Nil(t, err)
	syncAssumedBtcBlockChain(cacheDB)
	nativeService := getNativeFunc(nil, cacheDB)
	sh, _ := GetBestBlockHeader(nativeService, 0)
	epoch, err := GetEpoch(nativeService, 0, sh)
	assert.NoError(t, err)
	assert.Equal(t, "0f9188f13cb7b2c71f2a335e3a4fc328bf5beb436012afca590b1a11466e2206", epoch.BlockHash().String(),
		"Returned incorrect epoch")
}

func TestCalcRequiredWork(t *testing.T) {
	genesisHeader := chaincfg.TestNet3Params.GenesisBlock.Header
	cacheDB, err := syncGenesisHeader(&genesisHeader)
	assert.Nil(t, err)
	syncAssumedBtcBlockChain(cacheDB)

	nativeService := native.NewNativeService(cacheDB, nil, 0, 0, common.Uint256{}, 0, nil, false)
	bestHeader, err := GetBestBlockHeader(nativeService, 0)
	if err != nil {
		t.Error(err)
	}

	// Test during difficulty adjust period
	newHdr := wire.BlockHeader{}
	newHdr.PrevBlock = bestHeader.Header.BlockHash()
	work, err := calcRequiredWork(nativeService, 0, newHdr, 2016, bestHeader)
	if err != nil {
		t.Error(err)
	}
	if work < bestHeader.Header.Bits {
		t.Error("Returned in correct bits")
	}
	newHdr.Bits = work
	sh := StoredHeader{
		Header:    newHdr,
		Height:    2016,
		totalWork: blockchain.CompactToBig(work),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, sh.Height, sh.Header.BlockHash())

	// Test during normal adjustment
	netParam.ReduceMinDifficulty = false
	newHdr1 := wire.BlockHeader{}
	newHdr1.PrevBlock = newHdr.BlockHash()
	work1, err := calcRequiredWork(nativeService, 0, newHdr1, 2017, &sh)
	if err != nil {
		t.Error(err)
	}
	if work1 != work {
		t.Error("Returned in correct bits")
	}
	newHdr1.Bits = work1
	sh = StoredHeader{
		Header:    newHdr1,
		Height:    2017,
		totalWork: blockchain.CompactToBig(work1),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, sh.Height, sh.Header.BlockHash())

	// Test with reduced difficult flag
	netParam.ReduceMinDifficulty = true
	newHdr2 := wire.BlockHeader{}
	newHdr2.PrevBlock = newHdr1.BlockHash()
	work2, err := calcRequiredWork(nativeService, 0, newHdr2, 2018, &sh)
	if err != nil {
		t.Error(err)
	}
	if work2 != work1 {
		t.Error("Returned in correct bits")
	}
	newHdr2.Bits = work2
	sh = StoredHeader{
		Header:    newHdr2,
		Height:    2018,
		totalWork: blockchain.CompactToBig(work2),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, sh.Height, sh.Header.BlockHash())

	// Test testnet exemption
	newHdr3 := wire.BlockHeader{}
	newHdr3.PrevBlock = newHdr2.BlockHash()
	newHdr3.Timestamp = newHdr2.Timestamp.Add(time.Minute * 21)
	work3, err := calcRequiredWork(nativeService, 0, newHdr3, 2019, &sh)
	if err != nil {
		t.Error(err)
	}
	if work3 != netParam.PowLimitBits {
		t.Error("Returned in correct bits")
	}
	newHdr3.Bits = work3
	sh = StoredHeader{
		Header:    newHdr3,
		Height:    2019,
		totalWork: blockchain.CompactToBig(work3),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, sh.Height, sh.Header.BlockHash())

	// Test multiple special difficulty blocks in a row
	netParam.ReduceMinDifficulty = true
	newHdr4 := wire.BlockHeader{}
	newHdr4.PrevBlock = newHdr3.BlockHash()
	work4, err := calcRequiredWork(nativeService, 0, newHdr4, 2020, &sh)
	if err != nil {
		t.Error(err)
	}
	if work4 != work2 {
		t.Error("Returned in correct bits")
	}

}

func TestBlockchain_checkProofOfWork(t *testing.T) {
	// Test valid
	header0, err := hex.DecodeString(chain[0])
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	buf.Write(header0)
	hdr0 := wire.BlockHeader{}
	hdr0.Deserialize(&buf)
	if !checkProofOfWork(hdr0, &chaincfg.RegressionNetParams) {
		t.Error("checkProofOfWork failed")
	}

	// Test negative target
	neg := hdr0
	neg.Bits = 1000000000
	if checkProofOfWork(neg, &chaincfg.RegressionNetParams) {
		t.Error("checkProofOfWork failed to negative target")
	}

	// Test too high diff
	params := chaincfg.RegressionNetParams
	params.PowLimit = big.NewInt(0)
	if checkProofOfWork(hdr0, &params) {
		t.Error("checkProofOfWork failed to detect above max PoW")
	}

	// Test to low work
	badHeader := "1" + chain[0][1:]
	header0, err = hex.DecodeString(badHeader)
	if err != nil {
		t.Error(err)
	}
	badHdr := wire.BlockHeader{}
	buf.Write(header0)
	badHdr.Deserialize(&buf)
	if checkProofOfWork(badHdr, &chaincfg.RegressionNetParams) {
		t.Error("checkProofOfWork failed to detect insuffient work")
	}
}

func TestGetCommonAncestor(t *testing.T) {
	db, _ := syncGenesisHeader(&chaincfg.RegressionNetParams.GenesisBlock.Header)
	ns := getNativeFunc(nil, db)

	var hdr wire.BlockHeader
	for i, c := range chain {
		b, _ := hex.DecodeString(c)
		hdr.Deserialize(bytes.NewReader(b))
		sh := StoredHeader{
			Header:    hdr,
			Height:    uint32(i + 1),
			totalWork: big.NewInt(0),
		}
		putBlockHeader(ns, 0, sh)
		putBestBlockHeader(ns, 0, sh)
		putBlockHash(ns, 0, sh.Height, sh.Header.BlockHash())
	}
	prevBest := StoredHeader{Header: hdr, Height: 10}
	prevs := make([]string, len(fork)-1)
	for i := 0; i < len(fork)-1; i++ {
		b, _ := hex.DecodeString(fork[i])
		hdr.Deserialize(bytes.NewReader(b))
		prevs[i] = hdr.BlockHash().String()
		sh := StoredHeader{
			Header:    hdr,
			Height:    uint32(i + 1),
			totalWork: big.NewInt(0),
		}
		putBlockHeader(ns, 0, sh)
		putBestBlockHeader(ns, 0, sh)
		putBlockHash(ns, 0, sh.Height, sh.Header.BlockHash())
	}
	currentBest := StoredHeader{Header: hdr, Height: 11}

	last, hashes, err := GetCommonAncestor(ns, 0, &currentBest, &prevBest)
	assert.NoError(t, err)
	assert.Equal(t, len(prevs), len(hashes))
	for i, v := range prevs {
		assert.Equal(t, v, hashes[len(hashes)-i-1].String(), fmt.Sprintf("prevs not equal: no.%d shoud be %s, "+
			"not %s", i, v, hashes[len(hashes)-i-1].String()))
	}
	assert.Equal(t, last.Height, uint32(5))
}

func TestGetBestBlockHeader(t *testing.T) {
	ns := getNativeFunc(nil, nil)
	best, err := GetBestBlockHeader(ns, 0)
	assert.Error(t, err)

	db, _ := syncGenesisHeader(&netParam.GenesisBlock.Header)
	ns = getNativeFunc(nil, db)
	best, err = GetBestBlockHeader(ns, 0)
	assert.NoError(t, err)
	assert.Equal(t, best.Header.BlockHash().String(), netParam.GenesisBlock.Header.BlockHash().String())

	best, err = GetBestBlockHeader(ns, 1)
	assert.Error(t, err)
}

func TestGetHeaderByHash(t *testing.T) {
	ns := getNativeFunc(nil, nil)
	sh, err := GetHeaderByHash(ns, 0, *netParam.GenesisHash)
	assert.Error(t, err)
	if sh != nil {
		t.Fatal("wrong hash")
	}

	db, _ := syncGenesisHeader(&netParam.GenesisBlock.Header)
	ns = getNativeFunc(nil, db)
	sh, err = GetHeaderByHash(ns, 0, *netParam.GenesisHash)
	assert.NoError(t, err)
	assert.Equal(t, netParam.GenesisHash.String(), sh.Header.BlockHash().String())

	sh, err = GetHeaderByHash(ns, 1, *netParam.GenesisHash)
	assert.Error(t, err)
}

func TestGetBlockHashByHeight(t *testing.T) {
	ns := getNativeFunc(nil, nil)
	hash, err := GetBlockHashByHeight(ns, 0, 0)
	assert.Error(t, err)
	if hash != nil {
		t.Fatal("wrong hash")
	}

	db, _ := syncGenesisHeader(&netParam.GenesisBlock.Header)
	ns = getNativeFunc(nil, db)
	hash, err = GetBlockHashByHeight(ns, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, netParam.GenesisHash.String(), hash.String(), "wrong hash")

	hash, err = GetBlockHashByHeight(ns, 0, 1)
	assert.Error(t, err)
}

func TestGetPreviousHeader(t *testing.T) {
	netParam = &chaincfg.RegressionNetParams
	db, _ := syncGenesisHeader(&netParam.GenesisBlock.Header)
	ns := getNativeFunc(nil, db)

	_, err := GetPreviousHeader(ns, 0, netParam.GenesisBlock.Header)
	assert.Error(t, err)

	var hdr wire.BlockHeader
	b, _ := hex.DecodeString(chain[0])
	hdr.Deserialize(bytes.NewReader(b))

	_, _, _, err = commitHeader(ns, 0, hdr)
	assert.NoError(t, err)
	gsh, err := GetPreviousHeader(ns, 0, hdr)

	assert.NoError(t, err)
	assert.Equal(t, netParam.GenesisHash.String(), gsh.Header.BlockHash().String())
}

func syncGenesisHeader(genesisHeader *wire.BlockHeader) (*storage.CacheDB, error) {
	var buf bytes.Buffer
	_ = genesisHeader.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	btcHander := NewBTCHandler()

	hb := make([]byte, 4)
	binary.BigEndian.PutUint32(hb, 0)
	sink := new(common.ZeroCopySink)
	params := &scom.SyncGenesisHeaderParam{
		ChainID:       0,
		GenesisHeader: append(buf.Bytes(), hb...),
	}
	sink = new(common.ZeroCopySink)
	params.Serialization(sink)
	ns := getNativeFunc(sink.Bytes(), nil)
	_ = btcHander.SyncGenesisHeader(ns)

	return ns.GetCacheDB(), nil
}

func syncAssumedBtcBlockChain(cacheDB *storage.CacheDB) {
	nativeService := getNativeFunc(nil, cacheDB)
	bestHeader, _ := GetBestBlockHeader(nativeService, 0)
	x := bestHeader.Height
	last := bestHeader
	for i := 0; i < 2015; i++ {
		x++
		hdr := wire.BlockHeader{}
		hdr.PrevBlock = last.Header.BlockHash()
		hdr.Nonce = 0
		hdr.Timestamp = time.Now().Add(time.Minute * time.Duration(i))
		mr := make([]byte, 32)
		rand.Read(mr)
		ch, _ := chainhash.NewHash(mr)
		hdr.MerkleRoot = *ch
		hdr.Bits = bestHeader.Header.Bits
		hdr.Version = 3
		sh := StoredHeader{
			Header:    hdr,
			Height:    x,
			totalWork: big.NewInt(0),
		}
		putBlockHeader(nativeService, 0, sh)
		putBestBlockHeader(nativeService, 0, sh)
		putBlockHash(nativeService, 0, sh.Height, sh.Header.BlockHash())

		last = &sh
	}
}

func TestReIndexHeaderHeight(t *testing.T) {
	db, _ := syncGenesisHeader(&chaincfg.RegressionNetParams.GenesisBlock.Header)
	ns := getNativeFunc(nil, db)

	var hdr wire.BlockHeader
	for i, c := range chain {
		b, _ := hex.DecodeString(c)
		hdr.Deserialize(bytes.NewReader(b))
		sh := StoredHeader{
			Header:    hdr,
			Height:    uint32(i + 1),
			totalWork: big.NewInt(0),
		}
		putBlockHeader(ns, 0, sh)
		putBestBlockHeader(ns, 0, sh)
		putBlockHash(ns, 0, sh.Height, sh.Header.BlockHash())
	}

	var nb StoredHeader
	prevs := make([]chainhash.Hash, len(fork)-1)
	for i := 0; i < len(fork)-1; i++ {
		b, _ := hex.DecodeString(fork[i])
		hdr.Deserialize(bytes.NewReader(b))
		prevs[len(prevs)-i-1] = hdr.BlockHash()
		psh, _ := GetPreviousHeader(ns, 0, hdr)
		sh := StoredHeader{
			Header:    hdr,
			Height:    uint32(psh.Height + 1),
			totalWork: big.NewInt(0),
		}
		if i == len(fork)-2 {
			nb = sh
		}
		putBlockHeader(ns, 0, sh)
		putBestBlockHeader(ns, 0, sh)
		putBlockHash(ns, 0, sh.Height, sh.Header.BlockHash())
	}

	err := ReIndexHeaderHeight(ns, 0, 10, prevs, &nb)
	assert.NoError(t, err)
	for i := uint32(6); i <= 11; i++ {
		hash, _ := GetBlockHashByHeight(ns, 0, i)
		assert.Equal(t, true, hash.IsEqual(&prevs[11-i]))
	}
}
