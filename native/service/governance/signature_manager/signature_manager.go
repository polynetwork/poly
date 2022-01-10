package signature_manager

import (
	"crypto/sha256"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//function name
	ADD_SIGNATURE = "addSignature"
)

//Register methods of signature_manager contract
func RegisterSignatureManagerContract(native *native.NativeService) {
	native.Register(ADD_SIGNATURE, AddSignature)
}

func AddSignature(native *native.NativeService) ([]byte, error) {
	params := new(AddSignatureParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddSignature, contract params deserialize error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, params.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddSignature, checkWitness: %s, error: %v", params.Address.ToBase58(), err)
	}

	temp := sha256.Sum256(params.Subject)
	id := temp[:]
	//check consensus signs
	ok, err := CheckSigns(native, id, params.Signature, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddSignature, CheckSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.RelayerManagerContractAddress,
			States:          []interface{}{"AddSignatureQuorum", id, params.Subject},
		})
	return utils.BYTE_TRUE, nil

}
