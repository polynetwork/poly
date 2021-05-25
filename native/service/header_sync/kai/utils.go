package kai

import (
	"bytes"
	"fmt"

	kaiclient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func VerifyHeader(myHeader *kaiclient.FullHeader, info *EpochSwitchInfo) error {
	valset := myHeader.ValidatorSet
	// now verify this header
	if !bytes.Equal(info.NextValidatorsHash, valset.Hash().Bytes()) {
		return fmt.Errorf("VerifyHeader, block validator is not right, next validator hash: %s, "+
			"validator set hash: %s", info.NextValidatorsHash.String(), valset.Hash().Hex())
	}
	if !myHeader.Header.ValidatorsHash.Equal(valset.Hash()) {
		return fmt.Errorf("VerifyHeader, block validator is not right!, header validator hash: %s, "+
			"validator set hash: %s", myHeader.Header.ValidatorsHash.String(), valset.Hash().Hex())
	}
	if err := valset.VerifyCommit("", myHeader.Header.LastBlockID, myHeader.Header.Height-1, myHeader.Commit); err != nil {
		return err
	}
	return nil
}

func GetEpochSwitchInfo(service *native.NativeService, chainId uint64) (*EpochSwitchInfo, error) {
	val, err := service.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(hscommon.EPOCH_SWITCH), utils.GetUint64Bytes(chainId)))
	if err != nil {
		return nil, fmt.Errorf("failed to get epoch switching height: %v", err)
	}
	raw, err := cstates.GetValueFromRawStorageItem(val)
	if err != nil {
		return nil, fmt.Errorf("deserialize bytes from raw storage item err: %v", err)
	}
	info := &EpochSwitchInfo{}
	if err = info.Deserialization(common.NewZeroCopySource(raw)); err != nil {
		return nil, fmt.Errorf("failed to deserialize EpochSwitchInfo: %v", err)
	}
	return info, nil
}

func PutEpochSwitchInfo(service *native.NativeService, chainId uint64, info *EpochSwitchInfo) {
	sink := common.NewZeroCopySink(nil)
	info.Serialization(sink)
	service.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(hscommon.EPOCH_SWITCH), utils.GetUint64Bytes(chainId)),
		cstates.GenRawStorageItem(sink.Bytes()))
	notifyEpochSwitchInfo(service, chainId, info)
}

func notifyEpochSwitchInfo(native *native.NativeService, chainID uint64, info *EpochSwitchInfo) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States: []interface{}{chainID, info.BlockHash.String(), info.Height,
				info.NextValidatorsHash.String(), info.ChainID, native.GetHeight()},
		})
}
