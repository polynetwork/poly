package eth

import (
	"fmt"
	"testing"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/hex"
)

func TestGetEthBlockByNumber(t *testing.T) {
	num := uint32(6097203)
	blockData, err := GetEthBlockByNumber(num)
	if err != nil {
		fmt.Printf("err:%v", err)
	}
	fmt.Printf("blockData:%v\n", blockData)
}

func updateString(trie *trie.Trie, k, v string) {
	trie.Update([]byte(k), []byte(v))
}

func TestMissingKeyProof(t *testing.T) {
	mtrie := new(trie.Trie)

	key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")
	val := common.Hex2Bytes("162e")

	updateString(mtrie, string(key), string(val))
	updateString(mtrie, "otherkey", "otherval")
	updateString(mtrie, "otherkey2", "otherval2")

	for i, key := range []string{string(key)} {
		//proof := memorydb.New()
		proof :=  light.NewNodeSet()
		mtrie.Prove([]byte(key), 0, proof)
		bs,_:= rlp.EncodeToBytes(proof)
		fmt.Printf("============bs:%v\n",bs)
		fmt.Printf("proof:%v\n",proof)

		//it := proof.NewIterator()
		//for it.Next() {
		//	fmt.Printf("key :%v, val :%s\n",it.Key(), it.Value())
		//	fmt.Printf("key :%v, val :%v\n",it.Key(), it.Value())
		//}

		fmt.Printf("hash:%v\n",mtrie.Hash().Bytes())
		tmpp := memorydb.New()
		rlp.DecodeBytes(bs,tmpp)

		//if proof.Len() != 1 {
		//	t.Errorf("test %d: proof should have one element", i)
		//}

		fmt.Println("%+v:", proof)

		val, _, err := trie.VerifyProof(mtrie.Hash(), []byte(key), proof)
		//val, _, err := trie.VerifyProof(mtrie.Hash(), []byte(key), tmpp)
		if err != nil {
			t.Fatalf("test %d: failed to verify proof: %v\nraw proof: %x", i, err, proof)
		}
		//if val != nil {
		//	t.Fatalf("test %d: verified value mismatch: have %x, want nil", i, val)
		//}
		fmt.Printf("%v\n",val)
	}
}


func TestProof(t *testing.T) {
	byteslist := make([][]byte, 3)
	//byteslist[0] = common.Hex2Bytes("f8518080a036bb5f2fd6f99b186600638644e2f0396989955e201672f7e81e8c8f466ed5b9808080a0533b860a9589c48890ae7f18ee2a3f65eb9123ef1e780a8695c092a7fcebe26c80808080808080808080")
	//byteslist[1] = common.Hex2Bytes("f85180808080808080a02be56bb5d291730e9de9ca6fa80d60f17ed17267992bbd2e249d94918fc7116f8080808080a0d521a3267eeddff95e711f89799bd3ecf5866a01456fc2f930b367248b183b14808080")
	//byteslist[2] = common.Hex2Bytes("e5a020302b501e50f675161fd6ec96cb939560d754e19bc3353e35a1db177ed5cfdc8382162e")


	byteslist[0],_ = hex.DecodeString("f8518080a036bb5f2fd6f99b186600638644e2f0396989955e201672f7e81e8c8f466ed5b9808080a043444909a619758931a7b2476a75741bb838d304fe657ccfcf4d1f0aef21299680808080808080808080")
	byteslist[1] = common.Hex2Bytes("f87180808080808080a02be56bb5d291730e9de9ca6fa80d60f17ed17267992bbd2e249d94918fc7116f8080808080a0d521a3267eeddff95e711f89799bd3ecf5866a01456fc2f930b367248b183b1480a0b9594e73a5545f57e2430ae055187a5d1fc3b03ea4307f616e07719ecd2fee1080")
	byteslist[2] = common.Hex2Bytes("e5a020302b501e50f675161fd6ec96cb939560d754e19bc3353e35a1db177ed5cfdc8382162e")

	fmt.Printf("b[0]:%v\n",byteslist[0])
	fmt.Printf("b[1]:%v\n",byteslist[1])
	fmt.Printf("b[2]:%v\n",byteslist[2])


	bs,_:= rlp.EncodeToBytes(byteslist)
	fmt.Printf("bs:%v\n",crypto.Keccak256(bs))

	nl := new(light.NodeList)
	nl.Put(nil,byteslist[0])
	nl.Put(nil,byteslist[1])
	nl.Put(nil,byteslist[2])
	fmt.Printf("nl:%v\n",nl)

	acctroot := "f90211a05b21d1f1051be9fb67ff88a36df78e9b1c326c29e485397962dcff3eb7e61e23a0ab68109b7b4f2085a800ab7575e7b1ab684e4aa480301f42967fa5536657125aa0654b077fc666742555a1bcbc995547164006d152f92a9ef2e32147f6c9c92a50a0db47b38efa5fbdab84f28fbb74bb786b93a4406a3f1a4de783d97c252fb5879da01d2701d3eb708750c3be767d09e4fdcb23e3640d9825bbea687bb35af018b29ea0cca6d38758f96b7487e882adc790ed68b4297fd4fa47cc1968cddadc4db91c27a08b7d55a61e7b3f946673dc87f5deafe43bd3dbf54e555bbd2344d15179d09963a05b0f2e027ec07003084a10fa85ef87e8968bb4bf9ccbf53ea15d461eaadd787ba0cf6f749ac553066bd7d8aa9817ee64ca77e8dc3a1a27c10303ce4ef4e187583ba07507a94ff248c915e3254122b3244eb294e2b4c11e9e807e061573a6a301d84da0556fb84eb16c3af4ccbb057c555093e5f9d3248f5aeb0d28aa92638df3fe4a22a0f51b5515e2de5976c554234626fc85aa6914f058501addd9f0f8d12f07e7f662a0f4a5c69dd7ae4b9493b79df2222681b6923f1ecb6d7c97357d50e51adf5bc867a07ff02fa14e0af071ace651ed1ef21fe4476e5646f0a08aa17a0253e5726b16e7a0d3e28f565dd687acee52e2f20bd1ff36177e1d30ae923f03b9d0b88de3104a55a0d0b03880ffdab312806c20997e7e94c74717ccb65039b2cf82312067101b42ec80"
	accthash := common.Hex2Bytes(acctroot)
	fmt.Printf("accthash:%v\n",crypto.Keccak256(accthash))
	fmt.Printf("accthash:%v\n",common.HexToHash(acctroot).Hex())

	//db := memorydb.New()

	orignkey:=common.HexToHash("fa98bb293724fa6b012da0f39d4e185f0fe4a749")
	fmt.Printf("orignkey:%s\n",orignkey.Hex())

	key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")


	storagehash := common.HexToHash("4653bf8a1185cd66e7e42cc25fcc73425b974457b7f90d7c73a425338880f62a")
	//db.Put(stateroothash[:],bs)
	fmt.Printf("stateroothashbs:%v\n",crypto.Keccak256(storagehash[:]))


	stateroot:=common.HexToHash("a9b7cc1629394efd7181e20559b42638a5d9a8caed9c648c73333ac5b01c4168")
	fmt.Printf("stateroot:%v\n",stateroot[:])
	fmt.Printf("staterootbs:%v\n",crypto.Keccak256(stateroot[:]))

	ns := nl.NodeSet()
	fmt.Printf("nodeset:%v\n",ns)
	fmt.Printf("key count:%d\n",ns.KeyCount())

	_, v := ns.Get(storagehash[:])
	fmt.Printf("v:%v\n",v)
	fmt.Printf("stateroothash:%v\n",storagehash[:])


	val, _, err := trie.VerifyProof(storagehash, key, ns)
	if err != nil{
		fmt.Printf("err:%s\n",err.Error())
	}
	fmt.Printf("val is :%v\n",val)

}

func TestProof2(t *testing.T){
	acctroot:="f90211a033bd6c44d340bf007ca209f89a561e414f6d7a5faa25cd14b377c0d1569b5fada018e08d9a915f3e1db16a58e20a5b4a8568967ceb4cbb8bf81d7e796b3317bbdaa094c861577eb8c4c38fbfb6fd5bedf6e0b1ae42cbd6542d48f11ca01127232b45a012bd91b6dfbaf57f78dfbc51a478e0589b437344613ecd1b1cac09ee3df038eca0454ce1bac3adfdac8b1437afdf507634b42e38a38abb3412d193740ea3100e3ca005858ed613dfef7f753beb5547acc1c617624b69a1e87ac8c030763fd77b707ca054567ad5d9580dd19e779d31d176ba31e8a2b305753aee4ffca1fc33a7bcac79a0f4146915ceda37b1f05d9cdeb2561f3ea4860aca3baa3f5fcc7edc36c2bd5111a0896732d56a602fc37d72eb085e90eac4e1326542aceb83ced5b90be8958b0743a0abc1511f84ace4f9b981ed64d2918af6d49a4e91b78988944142fe2bde32f1b6a0552fa6e6dadccf24618606c06c7b09b81106216bf4be8c03b324609ac23cb8bba0b4c6f52e614c01828f3dc8d3b4147f85ce46694cae6d495c47dfe35ded4750d4a06aeb5b377cccbcc75e6c0f2e126d7776abf3190c3cd077a1582f5e9e6504189ca0beb74c4a98b925369874b31f65108bff3e8b22508e3013c6a483a66df89577bca085a073adef04ae42f613bf3a4ab6796e07c47ba7cfbd3da34138cc57d5a09fffa02bdc62c83d4ac46fe265eef863eabd30b2c62fc038ec7d842c84263c5bf6fcac80"
	hs := common.HexToHash(acctroot).Hex()
	fmt.Printf("hs :%s\n",hs)

	fmt.Printf("hs bytes:%v\n",crypto.Keccak256(common.Hex2Bytes(acctroot)))
	stateroot := "f761fb671353f53ec2b2bf050d6e105731ebd729e69e023a982871f460272fac"
	fmt.Printf("stateroot:%v\n",common.Hex2Bytes(stateroot))
	fmt.Printf("stateroot:%v\n",common.HexToHash(stateroot))

	acct1 := "f90211a033bd6c44d340bf007ca209f89a561e414f6d7a5faa25cd14b377c0d1569b5fada018e08d9a915f3e1db16a58e20a5b4a8568967ceb4cbb8bf81d7e796b3317bbdaa094c861577eb8c4c38fbfb6fd5bedf6e0b1ae42cbd6542d48f11ca01127232b45a012bd91b6dfbaf57f78dfbc51a478e0589b437344613ecd1b1cac09ee3df038eca0454ce1bac3adfdac8b1437afdf507634b42e38a38abb3412d193740ea3100e3ca005858ed613dfef7f753beb5547acc1c617624b69a1e87ac8c030763fd77b707ca054567ad5d9580dd19e779d31d176ba31e8a2b305753aee4ffca1fc33a7bcac79a0f4146915ceda37b1f05d9cdeb2561f3ea4860aca3baa3f5fcc7edc36c2bd5111a0896732d56a602fc37d72eb085e90eac4e1326542aceb83ced5b90be8958b0743a0abc1511f84ace4f9b981ed64d2918af6d49a4e91b78988944142fe2bde32f1b6a0552fa6e6dadccf24618606c06c7b09b81106216bf4be8c03b324609ac23cb8bba0b4c6f52e614c01828f3dc8d3b4147f85ce46694cae6d495c47dfe35ded4750d4a06aeb5b377cccbcc75e6c0f2e126d7776abf3190c3cd077a1582f5e9e6504189ca0beb74c4a98b925369874b31f65108bff3e8b22508e3013c6a483a66df89577bca085a073adef04ae42f613bf3a4ab6796e07c47ba7cfbd3da34138cc57d5a09fffa02bdc62c83d4ac46fe265eef863eabd30b2c62fc038ec7d842c84263c5bf6fcac80"
	acct2 := "f90211a031c8e9bf2c2221d7a29ec763cf7131252d6e02e5b21fd9e74498af965dcec515a033f1ab51b051c82085635660587c1b3dc59146a5d71206aa3f224f66e9b75a50a05fa048d7d39a9c2fb9d8249f12317b36b5c6445f59f48afca1e3c2ef6fdefa73a06df33806a323c4d79873f1901091fb3f8da48353e0af3896a5994c61d75ef192a0c623174d8897edc8ccb102d304f8efe92cbb74128f6ca69b083044f522f37237a023d95280d6ce1330607b8b770587dd31787153d28135333b97e813c6b07a9fbea0aa0d27f07c1f275deae36992b6f82ad11b5951b167e26b8cff2ef85847e0da2fa08b1539ac14ab6d25663254db882f8f424cfd3203cdc65d5026844afd847fd4e5a0025a94800290160c372d94aa78a08b59512cbc788f89843b8da7295522509e92a0818c768d2db2f2f0fbf167de2c489c40c20548aa8403c46eb4b0c80272106bcfa02cfa2a910f3961388e5b8666dce3e6cb696694e3ffff47d7626f8d6eedb744d8a0500df99a6dac1969ce2c917edbf1eea93d3c8fff8859db41af5049c41d9fa5f7a04b792f06db12a4e719bd951cfc85918a10d051ffeafb1e489c199516f5497018a0bd101ad42efad1685a9929178765a8ed840dcc144fa1e75868d12727a9b0d60ba016383336e648e31d7d55d425184b10f3d93bcab912ddaa931d2a347b1e4f6d3ba00e5cd92dcb22cc7fd0b5af2b5b79568768be838a67a7954e62abe442a22f334380"
	acct3 := "f90211a0c0899494c201e3f11b6d8b67e54c53838aa2543a65bee2ee8594ff592a42d248a0901ef92fc43998b501f5b721afb5e5f7b7b45c9ba3b1e343feb5c85fb802a06ba0b83614b67d4c49edb5858278a2c39ad0f5f079ca9659c011d4c6ea6dffa3259ca021cb4b2fd88e15fcb85cfbcba5c58fce4e52defef4c98f46e3536090647d4a75a0136894b03a375eb9dd27c7db5e2329c038d88f81771d0931a8a1d62009b79d57a0be9e1c7fc9ab54865c646eeedb5df59f470f79a5830b66a0eea50c105ff49f84a00d3e10b5535e80a9818dcaf55c5a88911377eaba8d4adde170207fec2fee49d0a030c093578ab4d82ffa3f21d50f9125e58b703ebb8d565f056dc673021ef84556a0a522814d49017249ac1af4399e2a4b3c3b6ed3e09b2e3ee6dae950c4023ddfc8a029af978323854378af0d48e7264770c79582a157b8ffc50e7e61ea050fd1e45ca02c557d6f9294ed8abbbda94cdf6adcb5b647492423ed5cc1461cedffcc808763a0f626a2f17abc4a3e840a54da4f93c18e0b0787ccb4876ea131b9472605e8a6aaa0e01987836544a68a2399b6a3a6d33db5ba6a067659fef824d03e7fac8b5e09aea07a1f2be82ae20a44c1909db1f4189930fbf48d1291da275e2560344709750f67a02102eb733ffdf0e79efb2884606df6b7108fe8edc59cde3c6ef0e06168f2ffaea02e9715262f5b371de7809dc422d06b08f1df81056708771a8a04c4bd69223df680"
	acct4 := "f90211a08cb4a83418785eedffc5dbfbf5caa3de8eb40b72c770450ef0d1de36f07ab525a0b91337386786d0679bb853506c3204876c44ab6014a9419ac98bd6646ac4ffcfa0ab9107881c992928a601c40f1234c585d1d9c6a75c0f16a7f97e089e4a899f29a0c1d4df4d396f9485f22f0f0e05749f936663c0bee52e91d3d6be8aaa2f6394d5a033ad0a9ab3d3b0776e67a54671165a40b3bb9a80f16d60df0ce0e39d050c76a5a0ebb08a9bf8f26ca106014282dcd3e00122697e5bb6895de1696fb341042d1626a05c7c6840174407b0f8f7d9c310b38e39d8c0bd109a03cb691da283c5aaef49aea0919eb9b125a7b09c8c337262dfc7253189ef88c5b5189577921fc85c8014a087a0e1589f44397bcd36efd3137d1fa2deed02c536ce28bb82b804791d84164097dba080d39b87abf8f42c5ac95c28cb9187b8ee421c6b1c3ecc33229602059321bbb1a02e497cfe4e30bbe2396caa242d9de50bcb19cc06d82320e2b1d536c2036704e9a0c1c8fce60cf2ab2320c828df3be9223002df7382b5f24fb2565566f550800d67a0e5740a04a21c678564343a90ee0a10571289640c7c4d39a5729747b321b672a5a0f356503d9a396ee2c87193662b473afdc58ce3176c7feb5448c040f6e145020ea0800f4efdb908f3f69091e01dbe7b9e5aa5cd1837a17efbeb3ea450bae330c4c4a0016de1857a12eb41dbcfa27d1e9682792e27f0edd63cfc1b9efc3fe74a66caf080"
	acct5 := "f90211a0312527d8277fb405cc0f2c0b2d890d98fb3aa34233766f4c986d8147e62faf04a0409fb0d5fcb5931f1cef5b6ce2532ee2e120c0b565bb69056c38a6165dd257d8a03b60cb027dfa4671464dee778058893a76b4df60ac64f2b352973a4bdd756680a05579d60fc9d2709e02e026742ebaf1edebcca72ba2ea21e430da180df551ffa3a080fe56a478c85cc02689e1e8420863c213d6728d6eb89911f193d1a98dd9dae7a02da5b3607504a4be58e5c43b49c955d468fe581ee442b13d1d0941cc6dc7f801a019f42fcd96a3f26cc1e7f2e99db92b11f9b0b9abd9e3d5b93a983f7bda364b37a0c74d2a823a83dcc21404a70fcbfa82987d62e62adccca645634cc86d8da60c25a0701b1b1ba5cbae80d123315825f06a9cdb8daeeb5a48c331900cb29722a169aca0a92dc30e6dc61862d16fd360e6511c6a22f95708a0b7ca8a404e0f6d9652b69ba039f262ff9d22b2cdc91046ca9a44cdc62842e325c041561e5d401230e78c15a9a0d155ae62e6aad845e179408eee718bd5faf4412cd83c49ed1f6fe7fb6bf1d394a02db22011d0135ce9af5556f55b2ae34d8c17a583d6a90bbfb34c06f0882d6a99a00380a27f7843bbaa8644ccf23dd7da051f585f531e12484de8adad888637af8aa0a9cd7b05edd20e1d20bcf74cba8ede21f18c71a04575fc3b8d9d41bcc3da872ba08c3e453438fd4a1dc3ef09e33ec207a8afee313410d2dce9b145b4fa68577b3580"
	acct6 := "f901d1a0df5893b8a8854eb18a36c98c52faa73bbdb333b02e2fc87d6ccb7c879623180e80a070c3184d2cce1451406f7849d5985578db73548f56dc272f92ea47571f03f29ba035232b4a4a43f890151223de6167bd4cf17e8ad751be6068b73a9009de9e1276a0fea614b64bf8e2670135914937b2562f9de6f867111522c1caff5ad5566ef1e2a04fa9c4a46444b1b22a2217e8abd7a7dca17c216a9e36d83bbfe5d9aa7d5fe3ff80a0d74bd7aa610b0083a8806f19748ab644f1a098207560fa58bc21f39b046f6e44a0cd59f1272e11f6e64f5b062d795253ac49be4eb882452ced68ab12b5788e2dc7a012205f811540de8fda82ad09eeef23cadb631a79c4df567d795386d262d4e21aa0b94fb1dc6f341fdb9b3dff958f0d248e688ce0002f9305d351175179677d7b17a062523c69cd949248d8e45cdc24b23e921a9dea7ba33d13465d46600541f9a46aa0bdf1716728836fd2a02b0a0f34c15b1f862932bb5a660b7c0508ac28ca02fd65a04a459704fb8140ce2405c17c4d57d7c23ee37355c5f096a27604d6648bbdc563a01b5574a7746e1d4a4f6e78884bae71b04eb96f26440df49725a540c36b7932eda0ea932253a4af15be71cd2137272f71df87783c5e473ecfda0f38f35a47bcd5e980"
	acct7 := "f8679e209fcb8333409cfcec7afda187755ee5973280c811102b4ee1fa4ce66f07b846f8440180a04653bf8a1185cd66e7e42cc25fcc73425b974457b7f90d7c73a425338880f62aa0d65d8429488d77be04ef6c0db2a37547c1e049bbe599dfc35dd8959b62cf4a32"

	//byteslist := make([][]byte, 7)
	//byteslist[0]= common.Hex2Bytes(acct1)
	//byteslist[1]= common.Hex2Bytes(acct2)
	//byteslist[2]= common.Hex2Bytes(acct3)
	//byteslist[3]= common.Hex2Bytes(acct4)
	//byteslist[4]= common.Hex2Bytes(acct5)
	//byteslist[5]= common.Hex2Bytes(acct6)
	//byteslist[6]= common.Hex2Bytes(acct7)

	nl := new(light.NodeList)
	nl.Put(nil,common.Hex2Bytes(acct1))
	nl.Put(nil,common.Hex2Bytes(acct2))
	nl.Put(nil,common.Hex2Bytes(acct3))
	nl.Put(nil,common.Hex2Bytes(acct4))
	nl.Put(nil,common.Hex2Bytes(acct5))
	nl.Put(nil,common.Hex2Bytes(acct6))
	nl.Put(nil,common.Hex2Bytes(acct7))

	ns := nl.NodeSet()

	fmt.Printf("ns :%v\n",ns)
	//key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")
	key := common.Hex2Bytes("0")
	var bss []byte
	rlp.DecodeBytes(key,&bss)
	fmt.Printf("bss:%v\n",bss)

	val,_,err:=trie.VerifyProof(common.HexToHash(stateroot),key,ns)
	if err != nil{
		t.Fatal("err:%s\n",err.Error())
	}
	fmt.Printf("val is %v\n",val)


}

type proofList [][]byte

func (n *proofList) Put(key []byte, value []byte) error {
	*n = append(*n, value)
	return nil
}

func (n *proofList) Delete(key []byte) error {
	panic("not supported")
}


func TestProoflist(t *testing.T){
	mtrie := new(trie.Trie)

	key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")
	val := common.Hex2Bytes("162e")

	updateString(mtrie, string(key), string(val))
	updateString(mtrie, "otherkey", "otherval")
	updateString(mtrie, "otherkey2", "otherval2")

	for i, key := range []string{string(key)} {
		//proof := memorydb.New()
		//proof :=  light.NewNodeSet()
		var proof proofList
		mtrie.Prove([]byte(key), 0, &proof)
		bs,_:= rlp.EncodeToBytes(proof)
		fmt.Printf("bs:%v\n",bs)
		fmt.Printf("proof:%v\n",proof)

		//it := proof.NewIterator()
		//for it.Next() {
		//	fmt.Printf("key :%v, val :%s\n",it.Key(), it.Value())
		//	fmt.Printf("key :%v, val :%v\n",it.Key(), it.Value())
		//}

		fmt.Printf("hash:%v\n",mtrie.Hash().Bytes())
		tmpp := memorydb.New()
		rlp.DecodeBytes(bs,tmpp)

		//if proof.Len() != 1 {
		//	t.Errorf("test %d: proof should have one element", i)
		//}

		fmt.Printf("%+v:", common.ToHexArray(proof))

		nl := new(light.NodeList)
		nl.Put(nil,proof[0])
		nl.Put(nil,proof[1])


		val, _, err := trie.VerifyProof(mtrie.Hash(), []byte(key), nl.NodeSet())
		//val, _, err := trie.VerifyProof(mtrie.Hash(), []byte(key), tmpp)
		if err != nil {
			t.Fatalf("test %d: failed to verify proof: %v\nraw proof: %x", i, err, proof)
		}
		//if val != nil {
		//	t.Fatalf("test %d: verified value mismatch: have %x, want nil", i, val)
		//}
		fmt.Printf("%v\n",val)
	}
}