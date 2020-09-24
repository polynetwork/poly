package fisco

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	hcom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/fisco"
)

type FiscoHandler struct{}

func NewFiscoHandler() *FiscoHandler {
	return &FiscoHandler{}
}

func (this *FiscoHandler) MakeDepositProposal(ns *native.NativeService) (*common.MakeTxParam, error) {
	params := new(common.EntranceParam)
	if err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput())); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, contract params deserialize error: %v", err)
	}
	val := &common.MakeTxParam{}
	if err := val.Deserialization(pcom.NewZeroCopySource(params.Extra)); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, failed to deserialize MakeTxParam: %v", err)
	}
	if err := common.CheckDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, check done transaction error: %v", err)
	}
	if err := common.PutDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, PutDoneTx error: %v", err)
	}

	root, err := fisco.GetFiscoRoot(ns, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, failed to get the fisco root cert: %v", err)
	}
	now := ns.GetBlockTime()
	if now.After(root.RootCA.NotAfter) || now.Before(root.RootCA.NotBefore) {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, Fisco root CA need to update for chain %d: (start: %d, end: %d, block_time: %d)",
			params.SourceChainID, root.RootCA.NotBefore.Unix(), root.RootCA.NotAfter.Unix(), now.Unix())
	}
	certs := &hcom.CertTrustChain{}
	if err := certs.Deserialization(pcom.NewZeroCopySource(params.HeaderOrCrossChainMsg)); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, failed to deserialize CertTrustChain: %v", err)
	}
	if err := certs.Validate(ns); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, validate not pass: %v", err)
	}
	if err := certs.CheckSigWithRootCert(root.RootCA, params.Extra, params.Proof); err != nil {
		return nil, fmt.Errorf("Fisco MakeDepositProposal, failed to check sig: %v", err)
	}
	PutFiscoLatestHeightInProcessing(ns, params.SourceChainID, val.FromContractAddress, params.Height)

	return val, nil
}
