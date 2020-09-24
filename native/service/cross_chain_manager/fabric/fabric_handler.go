package fabric

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/common"
)

type FabricHandler struct{}

func NewFabricHandler() *FabricHandler {
	return &FabricHandler{}
}

func (this *FabricHandler) MakeDepositProposal(ns *native.NativeService) (*common.MakeTxParam, error) {
	params := new(common.EntranceParam)
	if err := params.Deserialization(pcom.NewZeroCopySource(ns.GetInput())); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, contract params deserialize error: %v", err)
	}
	val := &common.MakeTxParam{}
	if err := val.Deserialization(pcom.NewZeroCopySource(params.Extra)); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to deserialize MakeTxParam: %v", err)
	}
	if err := common.CheckDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, check done transaction error: %v", err)
	}
	if err := common.PutDoneTx(ns, val.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Fabric MakeDepositProposal, PutDoneTx error: %v", err)
	}
	//rootCerts, err := fabric.GetFabricRoot(ns, params.SourceChainID)
	//if err != nil {
	//	return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to get the Fabric root certs: %v", err)
	//}
	//l := len(rootCerts.Certs)
	//rootCerts = rootCerts.ValidCAs(ns)
	//validL := len(rootCerts.Certs)
	//if validL < l {
	//	if err := fabric.PutFabricRoot(ns, rootCerts, params.SourceChainID); err != nil {
	//		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to put valid fabric root CAs: %v", err)
	//	}
	//}
	//if validL == 0 {
	//	return nil, fmt.Errorf("Fabric MakeDepositProposal, no valid root CA in poly's storage for Fabric chain %d", params.SourceChainID)
	//}
	//certs := hcom.MultiCertTrustChain(make([]*hcom.CertTrustChain, 0))
	//if err := certs.Deserialization(pcom.NewZeroCopySource(params.HeaderOrCrossChainMsg)); err != nil {
	//	return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to deserialize CertTrustChain: %v", err)
	//}
	//if err := certs.ValidateAll(ns); err != nil {
	//	return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to validate CAs: %v", err)
	//}
	//
	//isChecked := make(map[string]bool)
	//for i, trustChain := range certs {
	//	rawKey := hex.EncodeToString(trustChain.Certs[0].Raw)
	//	if isChecked[rawKey] {
	//		continue
	//	}
	//	var isExist bool
	//	for _, root := range rootCerts.Certs {
	//		if root.Equal(trustChain.Certs[0]) {
	//			isExist = true
	//			break
	//		}
	//	}
	//	if !isExist {
	//		continue
	//	}
	//	// TODO: need the signed raw info and the sig
	//	lastElem := trustChain.Certs[len(trustChain.Certs)-1]
	//	for _, ou := range lastElem.Subject.OrganizationalUnit {
	//		if strings.Contains(strings.ToLower(ou), "peer") {
	//			isExist = false
	//		}
	//	}
	//	if isExist {
	//		continue
	//	}
	//	if err := trustChain.CheckSig(params.Extra, params.Proof); err != nil {
	//		return nil, fmt.Errorf("Fabric MakeDepositProposal, failed to check signature for No.%d chain: %v", i, err)
	//	}
	//	isChecked[rawKey] = true
	//}
	//if len(isChecked) != len(rootCerts.Certs) {
	//	return nil, fmt.Errorf("Fabric MakeDepositProposal, only %d valid signature commited but %d needed.", len(isChecked), len(rootCerts.Certs))
	//}
	return val, nil
}
