package eth

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	scom "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, contract params deserialize error: %s", err)
	}
	if err := scom.CheckDoneTx(service, params.TxHash, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, check done transaction error:%s", err)
	}

	value, err := verifyFromEthTx(service, params.Proof, params.Extra, params.TxHash, params.Height)
	if err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, verifyFromEthTx error: %s", err)
	}

	if err := scom.PutDoneTx(service, value.TxHash, value.FromChainID); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, PutDoneTx error:%s", err)
	}
	return value, nil
}