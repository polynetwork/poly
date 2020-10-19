package fabric

import (
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

type FabricHandler struct{}

func NewFabricHandler() *FabricHandler {
	return &FabricHandler{}
}

func (this *FabricHandler) SyncGenesisHeader(ns *native.NativeService) error {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(ns.GetInput())); err != nil {
		return fmt.Errorf("FabricHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(ns)
	if err != nil {
		return fmt.Errorf("FabricHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	// check witness
	err = utils.ValidateOwner(ns, operatorAddress)
	if err != nil {
		return fmt.Errorf("FabricHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	certs := &scom.CertTrustChain{}
	if err := certs.Deserialization(common.NewZeroCopySource(params.GenesisHeader)); err != nil {
		return fmt.Errorf("FabricHandler SyncGenesisHeader, Deserialize CertArr error: %v", err)
	}
	if err := certs.Validate(ns); err != nil {
		return fmt.Errorf("FabricHandler SyncGenesisHeader, failed to validate CAs: %v", err)
	}

	prevCerts, err := GetFabricRoot(ns, params.ChainID)
	if err == nil && prevCerts != nil {
		cas := prevCerts.ValidCAs(ns)
		if len(cas.Certs) > 0 {
			certs.Certs = append(certs.Certs, cas.Certs...)
		}
	}

	if err = PutFabricRoot(ns, certs, params.ChainID); err != nil {
		return fmt.Errorf("FabricHandler SyncGenesisHeader, failed to put new fabric root CAs into storage: %v", err)
	}

	return nil
}

func (this *FabricHandler) SyncBlockHeader(ns *native.NativeService) error {
	return nil
}

func (this *FabricHandler) SyncCrossChainMsg(ns *native.NativeService) error {
	return nil
}
