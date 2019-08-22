package eth

import (
	"encoding/hex"
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

		fmt.Printf("%+v:", proof)

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

	stateroot := "409a0e505b47a59fa40595b34fcfc992f0d95b04cbe85efe68a2f119121237f7"
	fmt.Printf("stateroot:%v\n", common.Hex2Bytes(stateroot))

	acct1 := "f90211a093f52399a6fdbe6bd21778d99bcce8715ea22f64c58d7d1eddfe619520fcd00ba0df3b44f90b732cb8610d9efab96d4866136553f67e6bbc2a67a01faa70e2616ea09aaa17d04c123b8fe16abc74a2891a289e00f3880abcd342da44012458f53ea5a0173eb44d16cfbdc00da5da6bb0de5c699d92b5a6b26fa3ec53be97985c848145a08063e605bec744866352a1c7adf79e0442ba1c97ae432cffcf3d4ab14f7fa658a069066e84a5c7cfee8d3baea983c268fbff348d9f5bf0bfcb6014c4e9aabfb1f9a020dcd2d1fcf52383d2a62f349ca821621c5779cf421f3f7f0197aa2f8b5e8dbea0f77a7eccabb17a10b1aae85f8db3f139cefd616589d12eac9199dde4ab204c00a0ca858280619fd021a5d20c41ed53db3f8b4804a506be8f8e076e881915050b1ca005012e22940316da8fa348354b9f2e39478ca1e8326bd052a68180753de040e7a0d45c65c4679a65feb27153cff6d7af087bbf06034889706a4607ac0038d62511a08b63300aae91a4bfe127638bcd1ed20bf7bd69fb1d5547a7c1851d8a621ccf21a009e684e4a6e710a74fcac8b527c337a6a54151e3a3ab6ab3cbadebfa9565ba5ea0339bc45bfb7de75f4ee14fd28b72a82d7a3eb7d3277f1667fc3b74ac97ea0931a016d442799c4dca695121139d5a3607525d577db90753731a40fab4214ae489e1a0ffe60a0c0589099bf3fc20b6ca0f2953028449c2c426bb9539dc15fb3a0e358580"
	acct2 := "f90211a02d8b82ad3c65182aefde2602dbd716cd4e0c8502e799688ee9c5736f8e8f8bd4a074a81e2d8f553779eb7f52583ea4fa66f7aa64cf1ad86e214b93c1eb3f66cb8ba0cf5385c8ad4e6d16cb5014130a9407edff7d11f4fbc8403f1163101e0201cc3ca015592e621fb9933e3fec757f41c7d701f6ee57109168250499b064d929c14e5ba01568ce7463eaffcd7e5e67381a7a88f7fca6497ed677209d70c66d0ba2b141c0a01d8c1d7d100b96b8ac1484d7bd8d3769697bbbd47a92d302d37268d552b4f7c3a0a49029b85fa0ccca0eb81271d9156e59d7b793e371c1a6550ba3dee8dac349dfa0cd3369a19124810c50d61d748fa04d42efba3a91cd633ecd3611e156c7c703e3a0c4ceac54d7f05f9617638f77f6ab06aaec9d10ba567c0a835b71b10817e40b3ea0ccaffbc3e4bb39ac92e0f5f521381c00cdf82470bb8b0e8c0ad0ad8a6e16a16fa0cec68ebe4b44a4a7c23c9a9fcdb8748bf6babf6523854b63e01a5d0cdd811c80a0778caa83cad73894ec9dcb7468726ba8d16ca81eddc736467630919159562001a05440e3570f0c156dbb94785dcf0b2abb2808c56528f8a58ec06d695aba31cdfda00ffbdc89afefb2a7bdf981b0acfcb286cd2603cf3dca4f4ed66f0f172d63a3a7a0f09ec8665c69a3f917da74f3a91d0f0725b738a866bb3c2e2c83a852dc24a3c0a02668a5436ad4aee6f85bdb90f45f55d3bfadd0ca045ad71778fc7133767b34d480"
	acct3 := "f90211a0855ff209196d19b85d206615c3326a020925e7d6c529e3fb2e96fbeaac152fe3a0df6fc1a9de0844fc6ae1dcf58178e3067b109202b74769f58e63833a93ffc478a0e4018e72449320b6b5977ab0b2aa5ac259919ef8cd55c6c85b33f41fef9a7b38a041a901cd42a7556530ab6afd84c1458d951a57078c18f1d7e691295a066404fca013de0c44d8e5e4f36b45e875c733da8faf131822ee12bfb461bfd572ebd62375a09f93d4d73a895c8c95f26be7914946ff432ee662bbd206322cee15eb047fa234a0662d54e9eb68fe359fc5f628767bace85977ee351d675a5c7dd973ce5335541ca01219dac8673fc17308a020f96e5f79b27adb2ec13116115ef0b44a1acd2d9e9fa0691e17050cf9f462ea6e386e3ec5ec51143d7b812eb6ae5c620dd3b28f199e73a08356a54e33f3ff8c501b936199c7d5c133af2cc39bb8ab91b44427bfe10e8133a08405fab3e6dc1eb10a2202b31a6b9132b996f523be74f82911b1dd4ebac0e0d5a0ad3308f1aaeecdc76d4c308182430b023a35d342e251252a395149d3b9775702a0f2c9d776a0578a05af1e268cb20321422c622b61075810ca7f8c231ab6c8d474a04b0bd04c5743bb555f42b67c02d2cc7d2b7948a6c71f739976e9e8b838f1511aa090e0403bda985a00bcbb48f09c6052411b6f9037db4dba122bbe9aa19dccce6ba08ebc9dc52e39e5bf06853e555a7e28311a1d894e82fcf90f143c6ca695c7c06180"
	acct4 := "f90211a05dcb205ef2629e7ccfee77a112a09a374ae46d8014e78ce3b7b5722e71c8ecf4a047ae3f970499f0c81524a1670506faf31074300a77f261d689e94c768c11d612a05ce364b7b24688a75bba69ead42cb5038bc8e546029bf61f96de97e358e54ee9a057eb5938ebbcf39310c777735cf5b43b28226f5f861aafbf4c45a4d1686bd26fa0abed0edd2bb2563bae326ca3580f3bbc46b9a6b7c087fa00bec1de4645fbde6aa08593dc56a503d4e38c8d2295018de6470dd0df61be5c38f99051ff87f5a5aeeda0888256ee032d95b89ab443ea62c3b85b1562c6e6a1211eee62c4ebe18ed65da0a0099c2920bd70c2dd799662a5eedd2410106cef9d766437f4f0bb4d945e86d883a063eae253532e8ae16ae8f69b0fc29b3a736cce7a8f29c3fde8629d740e5e9c8aa0a08893f364b32aaf2b9e8f4cd03e45f9d8c2e2f4d7d216247a0d281f6952d893a06def261afa3b293bd046277ed3122d7b41690e2904a6c48aa070fffddde27a32a09729175669d81cf5dc95386c451103471a1819c1ffe32f252608858b434ccd38a077e4ddbfcf8b08d34f4c505d3a4d027040948fdfe71ecd625016c636db18b848a0f5a0204cd7decf2e918ef4ecf570740a8da24814b14d356a4519c19ba5143202a0e591c376e5ac49231d08780df48f0dcdcd5080d9199bbce21012e5207e549cb3a0065c572e02712b3041d87bc3cc8030802c93c0339e6b0b3c1577178143b2642880"
	acct5 := "f90211a05d4d0a6c0d5060a1a41fd85cba3939069dc2a91d284022d40b14a8854f6f39fba057520ed3c0355c1d2761d59bc437d1b6d127a3c3d3edb2e355e74c9fed858c79a0073cfb999e4c710217a2f7cd0d2caf8ad2537ebcc58eb7bf591127ab5e4df0dca07d21e47c216e364976cea9e593f90598a10144b13dd82c240a2fac7db7c0fe4ba0e2030fec2bf49ac36fb6ac8e6059cbdb359c62a144f0e230dfff49cec4796a68a0e02d5cae3c3e6d6a3eeb3ae59673dd3b2b37c7b483e87b931e658ac787dbb437a0c6006c1e91a3e8afb651a69bbcb4d332f3c1ddec712fffb5925ac624a8a9e3f3a0199adc52372322018c7a7b412fe0663db54c28620d21b2a6eebf3084feec6b48a09f4783b9e672af635011218e4d021b99b774eadd4d7375f5ec65e354c0149401a0d613d2c0259166141712539188bddda6caca1dce5e80b6a7ce83e2dce7d7dbcca04a445496fe9410ef3e233538076b968363e85805bcda3aed25bb6b1c5b8b7d89a06deb3d5e4f4b29b5e1621bffbc98a07f1760ba6d3ff32b9145a34393541930efa092dfb292e18be96722cfee5b211f1b1dc39675a7f7aa96a929a764228137b83ca06e711813d36fe386b0009cb420a27f6a9627201ec0cafcdd25a14c2b69fa65e2a0d772b615d6978cdc0a536340f79d86da7873a186d3ce21b343bd8889529b78a1a091b5dca3b5605c4fed708168642832dbda195014d5d8a179d9d07cdc28cc14b380"
	acct6 := "f901918080a03d4407f530e122df0db67c5699f7901f5946fbe5e2c458d7daf3e021347d48e5a00244db8741b579cde8cdc5821e33cf38f5f4cc36145d8b6f38cc8d726e254fc2a020882ad097d70d03b6ff606e4b64334e9c9d4bef941c4c751cff98cff8ead375a0727bcdd3539827bd6168dd8740169cde77f0dab0656730a6d737012051932483a00ff8d9e8b0f04930678575250286612a66bd5c90e3ca6e2e63d50365b7d06bdda08af8ff149f04146270794f7c8b2adefc82d2a46aab2a5f101cbea5f0ffaa5e30a03a14149e8d7e660b32e7cd07e0dc5414b7b569a8e0491fee22025c2abf57eeaca0ef0539f9e08ef9dd7d0e064378c6791eb91a8aedeb1bff1bb5e6d1f722bd600c80a0aacdeb55020c0cd5d6bc1784e6ca8659c67d7cd3bec45204b000494ee10b9a0aa017026eeb23c5814f16cde27bf509d1ef60984da55313318a603b46b5afbc8eb1a0010b36d4c41a8619e34e62e3db6dfbd187348e17e66fb8693763746d298f47f0a0354d4053e179f6f9453172e98fa8e272d96817afbdeb791584ce7e24db0187b68080"
	acct7 := "f89180808080808080a08b136c06b5614e85acf274fe82d6bf29a030936bb65de45a70898e5c3e984266a01e6361b5aa82e2d52c6f0726ef35b8556f960965b81081a8b54dd0025a4a9ac580a018b25d7d6d2f97de38402097f14b5862a05e2ad96ab602f7c764bdd3a98da46d808080a017ac7c55301319f2fef32dde70cd2ecb24036b2cf3e95603d17b6bcbb370960b8080"
	acct8 := "f8669d30e732946eecba85677337e00de24979ccc704d04449c767725f11b2a1b846f8440114a0862b754ca8c11a13231addf702d053e21b47db68efc1aa165d8e1ee0008c0dd6a0e3fed8a43c94ab0ccebe8278319f9aede7f4084b84942825231602bd86fafcf2"
	nl := new(light.NodeList)
	nl.Put(nil, common.Hex2Bytes(acct1))
	nl.Put(nil, common.Hex2Bytes(acct2))
	nl.Put(nil, common.Hex2Bytes(acct3))
	nl.Put(nil, common.Hex2Bytes(acct4))
	nl.Put(nil, common.Hex2Bytes(acct5))
	nl.Put(nil, common.Hex2Bytes(acct6))
	nl.Put(nil, common.Hex2Bytes(acct7))
	nl.Put(nil, common.Hex2Bytes(acct8))

	ns := nl.NodeSet()

	//fmt.Printf("ns :%v\n",ns)
	//key := common.Hex2Bytes("2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5")
	storagehash := common.HexToHash("862b754ca8c11a13231addf702d053e21b47db68efc1aa165d8e1ee0008c0dd6")
	codehash := common.HexToHash("e3fed8a43c94ab0ccebe8278319f9aede7f4084b84942825231602bd86fafcf2")

	acctkey := crypto.Keccak256(common.Hex2Bytes("f6dc652e2f7ab7a20d1cc4156d5a7122a9e966a5"))

	val, _, err := trie.VerifyProof(common.HexToHash(stateroot), acctkey, ns)
	if err != nil {
		t.Fatalf("err:%s\n", err.Error())
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
