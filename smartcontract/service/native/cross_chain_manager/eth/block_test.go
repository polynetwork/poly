package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
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
		proof := light.NewNodeSet()
		mtrie.Prove([]byte(key), 0, proof)
		bs, _ := rlp.EncodeToBytes(proof)
		fmt.Printf("============bs:%v\n", bs)
		fmt.Printf("proof:%v\n", proof)

		//it := proof.NewIterator()
		//for it.Next() {
		//	fmt.Printf("key :%v, val :%s\n",it.Key(), it.Value())
		//	fmt.Printf("key :%v, val :%v\n",it.Key(), it.Value())
		//}

		fmt.Printf("hash:%v\n", mtrie.Hash().Bytes())
		tmpp := memorydb.New()
		rlp.DecodeBytes(bs, tmpp)

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
		fmt.Printf("%v\n", val)
	}
}

func TestProof(t *testing.T) {
	byteslist := make([][]byte, 3)
	//byteslist[0] = common.Hex2Bytes("f8518080a036bb5f2fd6f99b186600638644e2f0396989955e201672f7e81e8c8f466ed5b9808080a0533b860a9589c48890ae7f18ee2a3f65eb9123ef1e780a8695c092a7fcebe26c80808080808080808080")
	//byteslist[1] = common.Hex2Bytes("f85180808080808080a02be56bb5d291730e9de9ca6fa80d60f17ed17267992bbd2e249d94918fc7116f8080808080a0d521a3267eeddff95e711f89799bd3ecf5866a01456fc2f930b367248b183b14808080")
	//byteslist[2] = common.Hex2Bytes("e5a020302b501e50f675161fd6ec96cb939560d754e19bc3353e35a1db177ed5cfdc8382162e")

	byteslist[0] = common.Hex2Bytes("f8518080a036bb5f2fd6f99b186600638644e2f0396989955e201672f7e81e8c8f466ed5b9808080a043444909a619758931a7b2476a75741bb838d304fe657ccfcf4d1f0aef21299680808080808080808080")
	byteslist[1] = common.Hex2Bytes("f87180808080808080a02be56bb5d291730e9de9ca6fa80d60f17ed17267992bbd2e249d94918fc7116f8080808080a0d521a3267eeddff95e711f89799bd3ecf5866a01456fc2f930b367248b183b1480a0b9594e73a5545f57e2430ae055187a5d1fc3b03ea4307f616e07719ecd2fee1080")
	byteslist[2] = common.Hex2Bytes("e5a020302b501e50f675161fd6ec96cb939560d754e19bc3353e35a1db177ed5cfdc8382162e")

	//fmt.Printf("b[0]:%v\n",byteslist[0])
	//fmt.Printf("b[1]:%v\n",byteslist[1])
	//fmt.Printf("b[2]:%v\n",byteslist[2])

	bs, _ := rlp.EncodeToBytes(byteslist)
	fmt.Printf("bs:%v\n", crypto.Keccak256(bs))

	nl := new(light.NodeList)
	nl.Put(nil, byteslist[0])
	nl.Put(nil, byteslist[1])
	nl.Put(nil, byteslist[2])
	fmt.Printf("nl:%v\n", nl)

	//acctroot := "f90211a05b21d1f1051be9fb67ff88a36df78e9b1c326c29e485397962dcff3eb7e61e23a0ab68109b7b4f2085a800ab7575e7b1ab684e4aa480301f42967fa5536657125aa0654b077fc666742555a1bcbc995547164006d152f92a9ef2e32147f6c9c92a50a0db47b38efa5fbdab84f28fbb74bb786b93a4406a3f1a4de783d97c252fb5879da01d2701d3eb708750c3be767d09e4fdcb23e3640d9825bbea687bb35af018b29ea0cca6d38758f96b7487e882adc790ed68b4297fd4fa47cc1968cddadc4db91c27a08b7d55a61e7b3f946673dc87f5deafe43bd3dbf54e555bbd2344d15179d09963a05b0f2e027ec07003084a10fa85ef87e8968bb4bf9ccbf53ea15d461eaadd787ba0cf6f749ac553066bd7d8aa9817ee64ca77e8dc3a1a27c10303ce4ef4e187583ba07507a94ff248c915e3254122b3244eb294e2b4c11e9e807e061573a6a301d84da0556fb84eb16c3af4ccbb057c555093e5f9d3248f5aeb0d28aa92638df3fe4a22a0f51b5515e2de5976c554234626fc85aa6914f058501addd9f0f8d12f07e7f662a0f4a5c69dd7ae4b9493b79df2222681b6923f1ecb6d7c97357d50e51adf5bc867a07ff02fa14e0af071ace651ed1ef21fe4476e5646f0a08aa17a0253e5726b16e7a0d3e28f565dd687acee52e2f20bd1ff36177e1d30ae923f03b9d0b88de3104a55a0d0b03880ffdab312806c20997e7e94c74717ccb65039b2cf82312067101b42ec80"
	//accthash := common.Hex2Bytes(acctroot)
	//fmt.Printf("accthash:%v\n",crypto.Keccak256(accthash))
	//fmt.Printf("accthash:%v\n",common.HexToHash(acctroot).Hex())

	//db := memorydb.New()

	orignkey := common.Hex2Bytes("fa98bb293724fa6b012da0f39d4e185f0fe4a749")
	fmt.Printf("orignkey:%v\n", orignkey)

	key := crypto.Keccak256(common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5"))

	storagehash := common.HexToHash("4653bf8a1185cd66e7e42cc25fcc73425b974457b7f90d7c73a425338880f62a")
	//db.Put(stateroothash[:],bs)
	fmt.Printf("stateroothashbs:%v\n", crypto.Keccak256(storagehash[:]))

	//stateroot:=common.HexToHash("a9b7cc1629394efd7181e20559b42638a5d9a8caed9c648c73333ac5b01c4168")
	//fmt.Printf("stateroot:%v\n",stateroot[:])
	//fmt.Printf("staterootbs:%v\n",crypto.Keccak256(stateroot[:]))

	ns := nl.NodeSet()
	fmt.Printf("nodeset:%v\n", ns)
	fmt.Printf("key count:%d\n", ns.KeyCount())

	_, v := ns.Get(storagehash[:])
	fmt.Printf("v:%v\n", v)
	fmt.Printf("stateroothash:%v\n", storagehash[:])

	val, _, err := trie.VerifyProof(storagehash, key, ns)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
	}
	fmt.Printf("val is :%v\n", val)
	i := new(big.Int)
	rlp.DecodeBytes(val, i)
	fmt.Printf("reaval is :%v\n", i)

}

func TestStateroot(t *testing.T) {
	acctroot := "f90211a0bb68b56d610d967adc1c099e7e25c4f150fd8330943da74d01805f630e8fa113a0b4b4ad25b4616b5f91f71df6cced5a7fe7e8a302818685ed45784517cd5b391ea07d2f09d6094eb5450a49f84d0d34edf2acabd2b8bf3dd516095f17f559121a9fa025597ee4b6401224eae0a538ad9aeba3c5a0b55d4382d89bc526a1fefddf3365a023b8a5629bd8b9621bb36614ba9d079ed41a7c65bfab1330c019ab73a7772af4a065894adaadb43bca8a6c3f4c0ee7e6c795b3e2c336905bd69117fe70bc73f8d4a007da2870b2b2ab3ae7e076cf5cd1c35636748de8e100a575e630e95d30f36412a0e4375aac8cd2192c990d1c9fa1dcdec8655af8b667d54645637051c25c9a2142a09ae8f60b12c9af2dbf647b48d1ee7f92a92d68de9b01642ee3aa9a7eb49c2dfaa06db5792abd78974c4512c4cc4d9d36389820fea25d998cba9d977b6878c975f4a0f9cd38e1a27bbefe6eebbfe6971a8496d53317a33372cc73ca7fe8cf08484981a0ff5ad71d9c35306fad9562046b32c836138eba4b91f7b3a44faac6031ae75810a0c0d85755f27d9f19e4a95d28163fbe32284cf5514b50f0642a5caf9e1f68ede6a0a05c99fc872fe26c484c7c9b9aa395b829c510102b6568ce410fa23286694067a0ad1d17bad073b48b2564fb4a49e9c854776eb51f127bd8d4f10308327367790da054e77903e1bdcc5735c16933c42be88344719057ea3a95466d29f290b666d28f80"
	hs := common.HexToHash(acctroot).Hex()
	fmt.Printf("hs :%s\n", hs)

	fmt.Printf("hs bytes:%v\n", crypto.Keccak256(common.Hex2Bytes(acctroot)))
	stateroot := "04f683819c3287a46d76a969aac09c3f158f580c275e6f316e9b76df8079f53b"
	fmt.Printf("stateroot:%v\n", common.HexToHash(stateroot))
}

func TestProof2(t *testing.T) {

	stateroot := "04f683819c3287a46d76a969aac09c3f158f580c275e6f316e9b76df8079f53b"
	fmt.Printf("stateroot:%v\n", common.Hex2Bytes(stateroot))

	acct1 := "f90211a0bb68b56d610d967adc1c099e7e25c4f150fd8330943da74d01805f630e8fa113a0b4b4ad25b4616b5f91f71df6cced5a7fe7e8a302818685ed45784517cd5b391ea07d2f09d6094eb5450a49f84d0d34edf2acabd2b8bf3dd516095f17f559121a9fa025597ee4b6401224eae0a538ad9aeba3c5a0b55d4382d89bc526a1fefddf3365a023b8a5629bd8b9621bb36614ba9d079ed41a7c65bfab1330c019ab73a7772af4a065894adaadb43bca8a6c3f4c0ee7e6c795b3e2c336905bd69117fe70bc73f8d4a007da2870b2b2ab3ae7e076cf5cd1c35636748de8e100a575e630e95d30f36412a0e4375aac8cd2192c990d1c9fa1dcdec8655af8b667d54645637051c25c9a2142a09ae8f60b12c9af2dbf647b48d1ee7f92a92d68de9b01642ee3aa9a7eb49c2dfaa06db5792abd78974c4512c4cc4d9d36389820fea25d998cba9d977b6878c975f4a0f9cd38e1a27bbefe6eebbfe6971a8496d53317a33372cc73ca7fe8cf08484981a0ff5ad71d9c35306fad9562046b32c836138eba4b91f7b3a44faac6031ae75810a0c0d85755f27d9f19e4a95d28163fbe32284cf5514b50f0642a5caf9e1f68ede6a0a05c99fc872fe26c484c7c9b9aa395b829c510102b6568ce410fa23286694067a0ad1d17bad073b48b2564fb4a49e9c854776eb51f127bd8d4f10308327367790da054e77903e1bdcc5735c16933c42be88344719057ea3a95466d29f290b666d28f80"
	acct2 := "f90211a0337ac9a0fb3adcfbaa3c74ad5db141204673466e9c87a46f0bd3b797b1f6b036a091c647c754204e7ac319e9adc7cfe50e803a70258cd7feae73964dd4ccd1aee5a02e96f1875817a742583369b521b0dd5c86cf481198bf4d62fe122bcd641ca431a00930a91c1b70cfaf73f8ef9f1d8634a388627bebb0cf12b7e028fc413a8bba59a06ea3ecf3cf143debc65c4dc2199a635f1ec49499e544eff75cd120ec3d36fdf9a0bf35e2ccd72d7ed800f349fd75950aadb991c8eb1d7189b6e95fa242c834e322a0f35007f59dbb1c884ca63a617e81614475decdacf056e0fb7be250185f6d1f75a09b73d4445b54217f2b2d8491f9967103309dff9aa36e6fa256bee22a4964b5d9a03c35959be07ce860a812807ce314fc925f2177f31ffb31ceb964013368ca37eca07bde942f5266655347c33c8272c19a08b7b13ef04afbc81a6d7262fd48b3ac90a0067fbec4d9769a692d880a3d1b34a045a9be08112e2f86b51ec813050ad44f81a0df81a2824481c944f15daa6db8fc06fd2a173e232ab283f3de1fb30dadc2f90ea05f2be0b462b9f7779f2f764a0abcc816ec6b1ecd647fd8e85528ee1c27a4830ea0286e749adb616b9872a09f430142d802f4127ee152400a7d50570453c836102aa04991d9bccf59c2b42ecd776225e7c4f680a6a67b2708cf83a302d8dc118bc38fa030bca81e1935851ec64b89c7f2e1803d81e8f43d47a644faf2865bd7a0b4d32a80"
	acct3 := "f90211a0d09ec1f1f6a880cd7f4cf391b07628ae237ee10f681c0ea3c1e537a23caf988ea0ac4c4291628af89eaf95f027791aebc853a3db924e6ac5f4dff4045385d89605a0ea453b28d358916f3cc41a7b6d9c0e476a614093c681724c48abc3385be504c2a0f3e7303a8a018c0e31a3a999ba9aec225cb913fcb01192532f201bd57eb2ebd3a0a2613ec3a79f00290bb244e66e7bfc30c83d4e87d4247a92e06aee1cb6c82903a09458e805b70c967d5656e6c4a952659a82ecd591ac30b4f419a6642f37b1d574a00d3e10b5535e80a9818dcaf55c5a88911377eaba8d4adde170207fec2fee49d0a031c5a1443e604365e0f66ba55593f974ffdb3da0e51d27020f13054ac3facafea0cb6def6ab5d774ffbd27c534e01ab839b3237dce1a137a12af9885485e5236d0a04f1daec3f868eadc61a7076f22814109bea6b7537a31e710df7067f559f9662ba01a3c33217d5204eaafabfe9e27cfc2329441a09568c7fc8c75e062d231029558a075ff1a6e9ac421b32620a6e7c91d4915b408a946f928c00c8c4a762e6400f2ada0abac0ed2bef7299d427ba45389e2db112ec224f7127b47457e520cdb54abe07da0b9173d31bc2d94a81b9d8387cd0993c515257b04133ad1943c95729490e32306a02f55e18a4b08c84dafb4f3905ff66f899c3d16064d992e039d814e01f79bc82ba0958d4eecdef3b2b841aa5a75ecfaeb9c990bad4cda1d8730e6d767fbf491fa7f80"
	acct4 := "f90211a0db30a189064d9e6763fdfc24d71f2af817f2595ec5ad616c5fa1ef49584592a0a0b91337386786d0679bb853506c3204876c44ab6014a9419ac98bd6646ac4ffcfa0ab9107881c992928a601c40f1234c585d1d9c6a75c0f16a7f97e089e4a899f29a0c1d4df4d396f9485f22f0f0e05749f936663c0bee52e91d3d6be8aaa2f6394d5a033ad0a9ab3d3b0776e67a54671165a40b3bb9a80f16d60df0ce0e39d050c76a5a0ebb08a9bf8f26ca106014282dcd3e00122697e5bb6895de1696fb341042d1626a05c7c6840174407b0f8f7d9c310b38e39d8c0bd109a03cb691da283c5aaef49aea0919eb9b125a7b09c8c337262dfc7253189ef88c5b5189577921fc85c8014a087a03dae4022ae540a88b22f65477072bea173d08570abcdec7e782a7065e45dfd48a080d39b87abf8f42c5ac95c28cb9187b8ee421c6b1c3ecc33229602059321bbb1a0218e22972e3f4a8877846999ed0c8b353955648b405172f560f77a8ca59c4081a0c1c8fce60cf2ab2320c828df3be9223002df7382b5f24fb2565566f550800d67a0e5740a04a21c678564343a90ee0a10571289640c7c4d39a5729747b321b672a5a04dfcb61fd31a1a5a0975691f4a74caa856fa9354febfbb393dd42ac8c8c8eb66a0800f4efdb908f3f69091e01dbe7b9e5aa5cd1837a17efbeb3ea450bae330c4c4a0016de1857a12eb41dbcfa27d1e9682792e27f0edd63cfc1b9efc3fe74a66caf080"
	acct5 := "f90211a0312527d8277fb405cc0f2c0b2d890d98fb3aa34233766f4c986d8147e62faf04a0409fb0d5fcb5931f1cef5b6ce2532ee2e120c0b565bb69056c38a6165dd257d8a03b60cb027dfa4671464dee778058893a76b4df60ac64f2b352973a4bdd756680a05579d60fc9d2709e02e026742ebaf1edebcca72ba2ea21e430da180df551ffa3a080fe56a478c85cc02689e1e8420863c213d6728d6eb89911f193d1a98dd9dae7a02da5b3607504a4be58e5c43b49c955d468fe581ee442b13d1d0941cc6dc7f801a019f42fcd96a3f26cc1e7f2e99db92b11f9b0b9abd9e3d5b93a983f7bda364b37a0c74d2a823a83dcc21404a70fcbfa82987d62e62adccca645634cc86d8da60c25a0701b1b1ba5cbae80d123315825f06a9cdb8daeeb5a48c331900cb29722a169aca0a92dc30e6dc61862d16fd360e6511c6a22f95708a0b7ca8a404e0f6d9652b69ba039f262ff9d22b2cdc91046ca9a44cdc62842e325c041561e5d401230e78c15a9a0d155ae62e6aad845e179408eee718bd5faf4412cd83c49ed1f6fe7fb6bf1d394a02db22011d0135ce9af5556f55b2ae34d8c17a583d6a90bbfb34c06f0882d6a99a00380a27f7843bbaa8644ccf23dd7da051f585f531e12484de8adad888637af8aa0a9cd7b05edd20e1d20bcf74cba8ede21f18c71a04575fc3b8d9d41bcc3da872ba08c3e453438fd4a1dc3ef09e33ec207a8afee313410d2dce9b145b4fa68577b3580"
	acct6 := "f901d1a0df5893b8a8854eb18a36c98c52faa73bbdb333b02e2fc87d6ccb7c879623180e80a070c3184d2cce1451406f7849d5985578db73548f56dc272f92ea47571f03f29ba035232b4a4a43f890151223de6167bd4cf17e8ad751be6068b73a9009de9e1276a0fea614b64bf8e2670135914937b2562f9de6f867111522c1caff5ad5566ef1e2a04fa9c4a46444b1b22a2217e8abd7a7dca17c216a9e36d83bbfe5d9aa7d5fe3ff80a0d74bd7aa610b0083a8806f19748ab644f1a098207560fa58bc21f39b046f6e44a0cd59f1272e11f6e64f5b062d795253ac49be4eb882452ced68ab12b5788e2dc7a012205f811540de8fda82ad09eeef23cadb631a79c4df567d795386d262d4e21aa0b94fb1dc6f341fdb9b3dff958f0d248e688ce0002f9305d351175179677d7b17a062523c69cd949248d8e45cdc24b23e921a9dea7ba33d13465d46600541f9a46aa0bdf1716728836fd2a02b0a0f34c15b1f862932bb5a660b7c0508ac28ca02fd65a04a459704fb8140ce2405c17c4d57d7c23ee37355c5f096a27604d6648bbdc563a01b5574a7746e1d4a4f6e78884bae71b04eb96f26440df49725a540c36b7932eda0ea932253a4af15be71cd2137272f71df87783c5e473ecfda0f38f35a47bcd5e980"
	acct7 := "f8679e209fcb8333409cfcec7afda187755ee5973280c811102b4ee1fa4ce66f07b846f8440180a04653bf8a1185cd66e7e42cc25fcc73425b974457b7f90d7c73a425338880f62aa0d65d8429488d77be04ef6c0db2a37547c1e049bbe599dfc35dd8959b62cf4a32"

	nl := new(light.NodeList)
	nl.Put(nil, common.Hex2Bytes(acct1))
	nl.Put(nil, common.Hex2Bytes(acct2))
	nl.Put(nil, common.Hex2Bytes(acct3))
	nl.Put(nil, common.Hex2Bytes(acct4))
	nl.Put(nil, common.Hex2Bytes(acct5))
	nl.Put(nil, common.Hex2Bytes(acct6))
	nl.Put(nil, common.Hex2Bytes(acct7))

	ns := nl.NodeSet()

	//fmt.Printf("ns :%v\n",ns)
	//key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")
	storagehash := common.HexToHash("4653bf8a1185cd66e7e42cc25fcc73425b974457b7f90d7c73a425338880f62a")
	codehash := common.HexToHash("d65d8429488d77be04ef6c0db2a37547c1e049bbe599dfc35dd8959b62cf4a32")

	acctkey := crypto.Keccak256(common.Hex2Bytes("aea963a10ab42d20071f69df96e9e0d27cefb76d"))

	val, _, err := trie.VerifyProof(common.HexToHash(stateroot), acctkey, ns)
	if err != nil {
		t.Fatal("err:%s\n", err.Error())
	}
	fmt.Printf("val is %v\n", val)
	acct := &proofAcct{
		Nounce:   big.NewInt(1),
		Balance:  big.NewInt(0),
		Storage:  storagehash,
		Codehash: codehash,
	}

	rlp_account, _ := rlp.EncodeToBytes(acct)
	fmt.Printf("rlp_account:%v\n", rlp_account)

	assert.Equal(t, rlp_account, val, "not equal")

}

type proofList [][]byte

func (n *proofList) Put(key []byte, value []byte) error {
	*n = append(*n, value)
	return nil
}

func (n *proofList) Delete(key []byte) error {
	panic("not supported")
}

type proofAcct struct {
	Nounce   *big.Int
	Balance  *big.Int
	Storage  common.Hash
	Codehash common.Hash
}

func TestProoflist(t *testing.T) {
	mtrie := new(trie.Trie)

	key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")
	val := common.Hex2Bytes("162e")

	updateString(mtrie, string(key), string(val))

	key2 := common.Hex2Bytes("fa98bb293724fa6b012da0f39d4e185f0fe4a749")
	val2 := common.Hex2Bytes("5678")
	updateString(mtrie, string(key2), string(val2))

	key3 := common.Hex2Bytes("1234567890123456789012345678901234567890")
	val4 := common.Hex2Bytes("6789")
	updateString(mtrie, string(key3), string(val4))

	for i, key := range []string{string(key)} {
		//proof := memorydb.New()
		//proof :=  light.NewNodeSet()
		var proof proofList
		mtrie.Prove([]byte(key), 0, &proof)
		bs, _ := rlp.EncodeToBytes(proof)
		fmt.Printf("bs:%v\n", bs)
		fmt.Printf("proof:%v\n", proof)

		//it := proof.NewIterator()
		//for it.Next() {
		//	fmt.Printf("key :%v, val :%s\n",it.Key(), it.Value())
		//	fmt.Printf("key :%v, val :%v\n",it.Key(), it.Value())
		//}

		fmt.Printf("hash:%v\n", mtrie.Hash().Bytes())
		tmpp := memorydb.New()
		rlp.DecodeBytes(bs, tmpp)

		//if proof.Len() != 1 {
		//	t.Errorf("test %d: proof should have one element", i)
		//}

		fmt.Printf("%+v:", common.ToHexArray(proof))

		nl := new(light.NodeList)
		nl.Put(nil, proof[0])
		nl.Put(nil, proof[1])

		val, _, err := trie.VerifyProof(mtrie.Hash(), []byte(key), nl.NodeSet())
		//val, _, err := trie.VerifyProof(mtrie.Hash(), []byte(key), tmpp)
		if err != nil {
			t.Fatalf("test %d: failed to verify proof: %v\nraw proof: %x", i, err, proof)
		}
		//if val != nil {
		//	t.Fatalf("test %d: verified value mismatch: have %x, want nil", i, val)
		//}
		fmt.Printf("%v\n", val)
	}
}
//
func TestAddress(t *testing.T) {
	p1 := "01"
	//p1 := "1234567890123456789012345678901234567890"
	//p1 := "2345678901234567890123456789012345678901"
	p2 := "00"
	v, err := MappingKeyAt(p1, p2)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
	}
	fmt.Printf("=========%v\n", hex.EncodeToString(v))

}
