package neo3_state_manager

import (
	"fmt"
	"github.com/polynetwork/poly/common"
)

type StateValidatorListParam struct {
	StateValidators []string       // public key strings in encoded format, each is 33 bytes in []byte
	Address         common.Address // for check witness?
}

func (this *StateValidatorListParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.StateValidators)))
	for _, v := range this.StateValidators {
		sink.WriteString(v)
	}
	sink.WriteVarBytes(this.Address[:])
}

func (this *StateValidatorListParam) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize StateValidators length error")
	}
	stateValidators := make([]string, 0, n)
	for i := 0; uint64(i) < n; i++ {
		ss, eof := source.NextString()
		if eof {
			return fmt.Errorf("source.NextString, deserialize stateValidator error")
		}
		stateValidators = append(stateValidators, ss)
	}

	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	this.StateValidators = stateValidators
	this.Address = addr
	return nil
}

type ApproveStateValidatorParam struct {
	ID      uint64         // StateValidatorApproveID
	Address common.Address // for check witness?
}

func (this *ApproveStateValidatorParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.ID)
	sink.WriteVarBytes(this.Address[:])
}

func (this *ApproveStateValidatorParam) Deserialization(source *common.ZeroCopySource) error {
	ID, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize ID error")
	}

	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	this.ID = ID
	this.Address = addr
	return nil
}
