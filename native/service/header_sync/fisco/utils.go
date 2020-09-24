package fisco

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutFiscoRoot(native *native.NativeService, root *FiscoRoot, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	sink := pcom.NewZeroCopySink(nil)
	root.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(common.ROOT_CERT), utils.GetUint64Bytes(chainID)),
		states.GenRawStorageItem(sink.Bytes()))

	common.NotifyPutCertificate(native, chainID, root.RootCA.Raw)
	return nil
}

func GetFiscoRoot(native *native.NativeService, chainID uint64) (*FiscoRoot, error) {
	store, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(common.ROOT_CERT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return nil, fmt.Errorf("GetFiscoRoot, get root error: %v", err)
	}
	if store == nil {
		return nil, fmt.Errorf("GetFiscoRoot, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return nil, fmt.Errorf("GetFiscoRoot, deserialize from raw storage item err: %v", err)
	}
	root := &FiscoRoot{}
	if err = root.Deserialization(pcom.NewZeroCopySource(raw)); err != nil {
		return nil, fmt.Errorf("GetFiscoRoot, failed to deserialize FiscoRoot: %v", err)
	}
	return root, nil
}
