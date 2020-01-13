package btc

import (
	"testing"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/native/storage"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/common"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"math/big"
	"github.com/stretchr/testify/assert"
	"fmt"
	"time"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"crypto/rand"
	"github.com/ontio/eth_tools/log"
	"github.com/btcsuite/btcd/chaincfg"
)



var (
	getNativeFunc = func() *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		service := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, nil, false, nil)
		return service
	}
	getBtcHanderFunc = func() *BTCHandler {
		return NewBTCHandler()
	}

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

func TestHeaderSync_BTC_SyncGenesisHeader(t *testing.T) {

	genesisHeader := StoredHeader{
		Header: netParam.GenesisBlock.Header,
		Height: 0,
		totalWork: big.NewInt(0),
	}
	cacheDB, err := syncGenesisHeader(genesisHeader)
	assert.Nil(t, err)

	nativeService := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, nil, false, nil)

	genesisBlockHash, err := GetBlockHash(nativeService, 0, uint64(genesisHeader.Height))
	assert.Nil(t, err)
	assert.Equal(t, genesisBlockHash.String(), genesisHeader.Header.BlockHash().String())
	bestBlockHeader, err := GetBestBlockHeader(nativeService, 0)
	assert.Nil(t, err)

	s1 := new(common.ZeroCopySink)
	genesisHeader.Serialization(s1)

	s2 := common.NewZeroCopySink(nil)
	bestBlockHeader.Serialization(s2)
	assert.Equal(t, s1.Bytes(), s1.Bytes())
}


func Test_GetEpoch(t *testing.T) {
	genesisHeader := StoredHeader{
		Header: chaincfg.RegressionNetParams.GenesisBlock.Header,
		Height: 0,
		totalWork: big.NewInt(0),
	}
	cacheDB, err := syncGenesisHeader(genesisHeader)
	assert.Nil(t, err)
	err = syncAssumedBtcBlockChain(cacheDB)
	if err != nil {
		fmt.Println("syncAssumedBtcBlockChain error is ", err)
	}

	nativeService := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, nil, false, nil)
	epoch, err := GetEpoch(nativeService, 0)
	if err != nil {
		t.Error(err)
	}
	if epoch.BlockHash().String() != "0f9188f13cb7b2c71f2a335e3a4fc328bf5beb436012afca590b1a11466e2206" {
		t.Error("Returned incorrect epoch")
	}

}


func Test_CalcRequiredWork(t *testing.T) {
	testnet3Prev1, _ := chainhash.NewHashFromStr("000000000003e8e7755d9b8299b28c71d9f0e18909f25bc9f3eeec3464ece1dd")
	testnet3Merk1, _ := chainhash.NewHashFromStr("7b91fe22059063bcbb1cfac6fd376cf459f4387d1bc1989989252495b06b52be")
	genesisHeader := StoredHeader{
		Header: wire.BlockHeader{
			Version:    536870912,
			PrevBlock:  *testnet3Prev1,
			MerkleRoot: *testnet3Merk1,
			Timestamp:  time.Unix(1517822323, 0),
			Bits:       453210804,
			Nonce:      2456211891,
		},
		Height: 0,
		totalWork: big.NewInt(0),
	}
	cacheDB, err := syncGenesisHeader(genesisHeader)
	assert.Nil(t, err)
	err = syncAssumedBtcBlockChain(cacheDB)
	if err != nil {
		fmt.Println("syncAssumedBtcBlockChain error is ", err)
	}

	nativeService := native.NewNativeService(cacheDB, nil, 0, 0, common.Uint256{}, 0, nil, false, nil)
	bestHeader, err := GetBestBlockHeader(nativeService, 0)
	if err != nil {
		t.Error(err)
	}

	epoch, err := GetEpoch(nativeService, 0)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("getEpoch -- ", epoch.BlockHash().String())
	fmt.Println("genesis -- ", genesisHeader.Header.BlockHash().String())

	// Test during difficulty adjust period
	newHdr := wire.BlockHeader{}
	newHdr.PrevBlock = bestHeader.Header.BlockHash()
	work, err := calcRequiredWork(nativeService, 0, newHdr, 2016, bestHeader)
	if err != nil {
		t.Error(err)
	}
	if work <= bestHeader.Header.Bits {
		t.Error("Returned in correct bits")
	}
	newHdr.Bits = work
	sh := StoredHeader{
		Header:    newHdr,
		Height:    2016,
		totalWork: CompactToBig(work),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, uint64(sh.Height), sh.Header.BlockHash())


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
		totalWork: CompactToBig(work1),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, uint64(sh.Height), sh.Header.BlockHash())



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
		totalWork: CompactToBig(work2),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, uint64(sh.Height), sh.Header.BlockHash())


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
		totalWork: CompactToBig(work3),
	}
	putBlockHeader(nativeService, 0, sh)
	// update fixedkey -> bestblockheader
	putBestBlockHeader(nativeService, 0, sh)
	// update height -> blockhash
	putBlockHash(nativeService, 0, uint64(sh.Height), sh.Header.BlockHash())


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


func TestBlockchain_CommitHeader(t *testing.T) {

}



func syncGenesisHeader (genesisHeader StoredHeader) (*storage.CacheDB, error) {
	store, _ := leveldbstore.NewMemLevelDBStore()
	cacheDB := storage.NewCacheDB(overlaydb.NewOverlayDB(store))


	btcHander := getBtcHanderFunc()

	sink := new(common.ZeroCopySink)
	genesisHeader.Serialization(sink)
	params := &scom.SyncGenesisHeaderParam{
		ChainID: 0,
		GenesisHeader: sink.Bytes(),
	}
	sink = new(common.ZeroCopySink)
	params.Serialization(sink)
	paramsBs := sink.Bytes()

	nativeService := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, paramsBs, false, nil)

	btcHander.SyncGenesisHeader(nativeService)

	return cacheDB, nil
}

func syncAssumedBtcBlockChain(cacheDB *storage.CacheDB) (error) {
	nativeService := native.NewNativeService(cacheDB, nil, 0, 200, common.Uint256{}, 0, nil, false, nil)
	bestHeader, err := GetBestBlockHeader(nativeService, 0)
	if err != nil {
		log.Errorf("syncBtcBlockChain error %v", err)
		return err
	}
	x := bestHeader.Height
	last := bestHeader.Header
	for i:= 0; i < 2015; i++ {
		x++
		hdr := wire.BlockHeader{}
		hdr.PrevBlock = last.BlockHash()
		hdr.Nonce = 0
		hdr.Timestamp = time.Now().Add(time.Minute * time.Duration(i))
		mr := make([]byte, 32)
		rand.Read(mr)
		ch, err := chainhash.NewHash(mr)
		if err != nil {
			return err
		}
		hdr.MerkleRoot = *ch
		hdr.Bits = bestHeader.Header.Bits
		hdr.Version = 3
		sh := StoredHeader{
			Header:    hdr,
			Height:    x,
			totalWork: big.NewInt(0),
		}
		putBlockHeader(nativeService, 0, sh)
		// update fixedkey -> bestblockheader
		putBestBlockHeader(nativeService, 0, sh)
		// update height -> blockhash
		putBlockHash(nativeService, 0, uint64(sh.Height), sh.Header.BlockHash())
		last = hdr
	}
	return nil

}
