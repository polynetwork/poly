package fabric

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutFabricRoot(native *native.NativeService, rootCerts *common.CertTrustChain, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	sink := pcom.NewZeroCopySink(nil)
	rootCerts.Serialization(sink)
	rawCerts := sink.Bytes()
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(common.MULTI_ROOT_CERT), utils.GetUint64Bytes(chainID)),
		states.GenRawStorageItem(rawCerts))

	common.NotifyPutCertificate(native, chainID, rawCerts)
	return nil
}

func GetFabricRoot(native *native.NativeService, chainID uint64) (*common.CertTrustChain, error) {
	store, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(common.MULTI_ROOT_CERT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return nil, fmt.Errorf("GetFabricRoot, get root error: %v", err)
	}
	if store == nil {
		return nil, fmt.Errorf("GetFabricRoot, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return nil, fmt.Errorf("GetFabricRoot, deserialize from raw storage item err: %v", err)
	}
	root := &common.CertTrustChain{}
	if err = root.Deserialization(pcom.NewZeroCopySource(raw)); err != nil {
		return nil, fmt.Errorf("GetFabricRoot, failed to deserialize fabric root CAs: %v", err)
	}
	return root, nil
}
