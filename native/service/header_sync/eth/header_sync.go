package eth

import (
	"encoding/json"
	"fmt"
	cty "github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	var header cty.Header
	err = json.Unmarshal(params.GenesisHeader, &header)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, deserialize header err: %v", err)
	}

	//block header storage
	err = PutBlockHeader(native, header, params.GenesisHeader)
	if err != nil {
		return fmt.Errorf("ETHHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}

	return nil
}

func (this *ETHHandler) SyncBlockHeader(native *native.NativeService) error {
	return nil
}
