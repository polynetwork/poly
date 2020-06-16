package cosmos

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/account"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/core/store/leveldbstore"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	ccmcom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	synccom "github.com/ontio/multi-chain/native/service/header_sync/cosmos"
	"github.com/ontio/multi-chain/native/service/utils"
	"github.com/ontio/multi-chain/native/storage"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var (
	acct *account.Account = account.NewAccount("")
	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}
)

func init() {
	setBKers()
}


const (
	SUCCESS = iota
	HEADER_NOT_EXIST
	PROOF_FORMAT_ERROR
	VERIFY_PROOT_ERROR
	TX_HAS_COMMIT
	UNKNOWN
)

func typeOfError(e error) int {
	if e == nil {
		return SUCCESS
	}
	errDesc := e.Error()
	if strings.Contains(errDesc, "GetHeaderByHeight, height is too big") {
		return HEADER_NOT_EXIST
	} else if strings.Contains(errDesc, "unmarshal proof error:") {
		return PROOF_FORMAT_ERROR
	} else if strings.Contains(errDesc, "verify proof value hash failed") {
		return VERIFY_PROOT_ERROR
	} else if strings.Contains(errDesc, "check done transaction error:checkDoneTx, tx already done") {
		return TX_HAS_COMMIT
	}
	return UNKNOWN
}

func NewNative(args []byte, tx *types.Transaction, db *storage.CacheDB) *native.NativeService {
	if db == nil {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
	}
	ns := native.NewNativeService(db, tx, 0, 0, common.Uint256{0}, 0, args, false)

	contaractAddr, _ := hex.DecodeString("48A77F43C0D7A6D6f588c4758dbA22bf6C5D95a0")
	side := &side_chain_manager.SideChain{
		Name:         "cosmos",
		ChainId:     5,
		BlocksToWait: 1,
		Router:       1,
		CCMCAddress:  contaractAddr,
	}
	sink := common.NewZeroCopySink(nil)
	_ = side.Serialization(sink)
	ns.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(side_chain_manager.SIDE_CHAIN), utils.GetUint64Bytes(2)), cstates.GenRawStorageItem(sink.Bytes()))
	return ns
}

func TestProofHandle(t *testing.T) {
	cosmosHeaderSync := synccom.NewCosmosHandler()
	cosmosProofHandler := NewCosmosHandler()
	var native *native.NativeService
	{
		header10000, _ := hex.DecodeString("0aa8020a02080a120774657374696e6718904e220c08eae29ff70510ffc5bfd5022a480a20a952550d320e34bd910fdf47d4b0d1b9b2c691d0432ebe5c2f25174c484663781224080112206fc745af7c59194f8e841a4e8bf89e6b7e873d1dfb807829deafc1093fcbe33e3220e728973f68379b28c499cfa4a6234a79ca6e7579290fa0c5013a65ce7af69db2422058df3ad01815e689d296705e219563932f8edd3637c1cd8f4a785906ca8883794a2058df3ad01815e689d296705e219563932f8edd3637c1cd8f4a785906ca8883795220048091bc7ddc283f77bfbf91d73c44da58c3df8a9cbc867405d8b7f3daada22f5a2021b3a41b885e7248c813ee0290f540cd6a1227d21e263a875497a6aa7972e4ec72146ff75a0ce1ed3596eb34a107bcfc1bebd1ea947812b70108904e1a480a20d7147608a68aa6e72f26d196dd5ea13ab1e580eee72cb1d84b3e6e49c5a5ffd3122408011220e9dc35792711de4bb8b19ecb7ce888cbd0835bec6b16d7dfc33fe4667bccb8092268080212146ff75a0ce1ed3596eb34a107bcfc1bebd1ea94781a0c08efe29ff70510cd98fefe022240e3807de6d7d219a0c2ca6187b6a447711de395f8a51053ed7fde07f522b1977987411d74128a31a03c0df91790b10596776fe5381363706086ca4086cb96ca061a3f0a146ff75a0ce1ed3596eb34a107bcfc1bebd1ea947812251624de6420760145874ef07a40698eea7afdf7d89719c76c96a5517ac2cf1162bb2e0a70a21864")
		param := new(scom.SyncGenesisHeaderParam)
		param.ChainID = 5
		param.GenesisHeader = header10000
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		tx := &types.Transaction{
			SignedAddr: []common.Address{acct.Address},
		}

		native = NewNative(sink.Bytes(), tx, nil)
		err := cosmosHeaderSync.SyncGenesisHeader(native)
		if err != nil {
			fmt.Printf("err: %s", err.Error())
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
	{
		header13572, _ := hex.DecodeString("0aa8020a02080a120774657374696e6718846a220c08f2f0a0f70510f689da87022a480a2086df1c2b681272092ad44ffde52b2b1b87ded88a2747981aa2d75bf54d3f1639122408011220be753dc5028ccb846c1e4659d3eedb823aa7f446eb96372c967fecc8ff19e32a3220676caa12e9e97fa8f9daa3b275db1ed228db9786a7ae65a6b33bf5ef4fa39c37422058df3ad01815e689d296705e219563932f8edd3637c1cd8f4a785906ca8883794a2058df3ad01815e689d296705e219563932f8edd3637c1cd8f4a785906ca8883795220048091bc7ddc283f77bfbf91d73c44da58c3df8a9cbc867405d8b7f3daada22f5a202ae963bb0275e55a29242189f7de13bb9305e8b9df55fba999be2a3e9341dca672146ff75a0ce1ed3596eb34a107bcfc1bebd1ea947812b70108846a1a480a206d7721829ae2780537f5d59777c758138a3e42e5e763d975a21702a295c525c2122408011220bf25427d0caa57a50665de59a6448d8f0f1cf322090c0fef9adde0785476bdc82268080212146ff75a0ce1ed3596eb34a107bcfc1bebd1ea94781a0c08f7f0a0f70510c7b887ac022240d8050c71f6e507e62dfd91c33876936fdc8aab73f89a3a18f2fd6499e81f81ddbeaae9b1034c7c9d248f8465c316b2b7f995ba2683eb8aabe45e39de12dda60e1a3f0a146ff75a0ce1ed3596eb34a107bcfc1bebd1ea947812251624de6420760145874ef07a40698eea7afdf7d89719c76c96a5517ac2cf1162bb2e0a70a21864")
		header13573, _ := hex.DecodeString("0aa8020a02080a120774657374696e6718856a220c08f7f0a0f70510c7b887ac022a480a206d7721829ae2780537f5d59777c758138a3e42e5e763d975a21702a295c525c2122408011220bf25427d0caa57a50665de59a6448d8f0f1cf322090c0fef9adde0785476bdc83220ea9121668dc77fab1d69e1c0b880a7bc4722e769038e7268ee1c64e0d4d722ad422058df3ad01815e689d296705e219563932f8edd3637c1cd8f4a785906ca8883794a2058df3ad01815e689d296705e219563932f8edd3637c1cd8f4a785906ca8883795220048091bc7ddc283f77bfbf91d73c44da58c3df8a9cbc867405d8b7f3daada22f5a20e24e0f6493e0224e6e6264d3863e3b4e7188ece078d8ad48ec2c9b39d25a9b0b72146ff75a0ce1ed3596eb34a107bcfc1bebd1ea947812b70108856a1a480a20ee655d95e37b85aae16d4021b377c696fda2d3add78eeee3048af897a2081b9412240801122000200af6fa3e089d99b3302889e7fba744557ee40d86a0a2a87c8c40b041b0ce2268080212146ff75a0ce1ed3596eb34a107bcfc1bebd1ea94781a0c08fcf0a0f70510e2c88ad5022240ac1bcc58718c4214c5425008ef6b75219266bc5a57a748bea26d0703b0c3b116ac48a197ff8f9082dafefc73e5ca8731aa5ed34a7de0f2bf859d1b1fce9844021a3f0a146ff75a0ce1ed3596eb34a107bcfc1bebd1ea947812251624de6420760145874ef07a40698eea7afdf7d89719c76c96a5517ac2cf1162bb2e0a70a21864")
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 5
		param.Address = acct.Address
		param.Headers = append(param.Headers, header13572)
		param.Headers = append(param.Headers, header13573)
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), nil, native.GetCacheDB())
		err := cosmosHeaderSync.SyncBlockHeader(native)
		if err != nil {
			fmt.Printf("err: %s", err.Error())
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
	{
		param := new(ccmcom.EntranceParam)
		proof, _ := hex.DecodeString("0ae6010a066961766c3a761215010d861307436c8d0b313d419a88852e74986b05571ac401c2010abf010a290808100918846a2a2092e7afcb12751217f584d3ccf731fc1a8d8027d9036d50a4194281f15f5d90c60a290804100418b8192a20f2e4aceedab0c1f8ae9060f1635552259dac30df29e7b4b7d5e960eeb39d03d80a290802100218b8192a201ac754249ca8068d2df0f5487f013563d4e1fc63f8bb4d2fb9b18f552f9aa5e91a3c0a15010d861307436c8d0b313d419a88852e74986b0557122066d27c6c53fcb53c305dd79a97c56fb763a9617e3072db4eda801f73443db69618b8190adb040a0a6d756c746973746f726512036163631ac704c5040ac2040a370a0c646973747269627574696f6e12270a2508846a12208655480c4b76b57527befbb33fdae9fc9e63d576b55f3c0599a055e790efb7ab0a310a06706172616d7312270a2508846a1220b9fee0c6611681cf122afde370e094c85b1534e8bdd19639480b92cd44e1bee90a2e0a0361636312270a2508846a12204693fe6fb0b911e01685fe900bb6daea6d8dba4658fbcda0ba1fce60ca26d4800a320a077374616b696e6712270a2508846a1220896549478b3e57487a0805a803390a83e202137b164f9024e0fa6cb43b450c4d0a100a077570677261646512050a0308846a0a120a096c6f636b70726f787912050a0308846a0a2f0a046d696e7412270a2508846a12209403885b7c0863462505036cf5d391c75510b71aba7c0549521b14e5ff90a5250a0c0a0363636d12050a0308846a0a0b0a02667412050a0308846a0a2e0a03676f7612270a2508846a122060cd137c1962ecac616389d68034833f2921509c2d285aa5f5153997ce968a740a2f0a046d61696e12270a2508846a1220b19cf098b43b3c7d54799fef0944aa95099f82ddc219fa0c8092e452a282bf410a130a0a68656164657273796e6312050a0308846a0a330a08736c617368696e6712270a2508846a1220bd2295b6de7a73e278e27e7f3c91b96b73a7c2b9efe8a5b7d6e0cf6ce6b61d960a310a06737570706c7912270a2508846a122031ec31dd7558883c3d260b3f013a5b849f1f332f164fbd4ebfae37d8a02fb61c0a0d0a046274637812050a0308846a0a110a0865766964656e636512050a0308846a")
		value, _ := hex.DecodeString("0a322f6163632f253031253044253836253133253037436c253844253042313d412539412538382538352e742539386b25303557127cf6e4f8380a140d861307436c8d0b313d419a88852e74986b055712160a057374616b65120d39393939383939393930303030121e0a0e76616c696461746f72746f6b656e120c3130303030303030303030301a26eb5ae98721037291f71277854c4ca5a06012019d612a5ebdeb689ba994e63f02a74a798583d22802")
		param.SourceChainID = 5
		param.Height = 13573
		param.Proof = proof
		param.RelayerAddress = acct.Address[:]
		param.Extra = value
		param.HeaderOrCrossChainMsg = []byte{}
		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)

		native = NewNative(sink.Bytes(), nil, native.GetCacheDB())
		_, err := cosmosProofHandler.MakeDepositProposal(native)
		if err != nil {
			fmt.Printf("%v", err)
		}
		assert.Equal(t, SUCCESS, typeOfError(err))
	}
}
