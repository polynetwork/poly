package signature_manager

import (
	"fmt"

	"github.com/polynetwork/poly/common"
)

type AddSignatureParam struct {
	Address   common.Address
	Subject   []byte
	Signature []byte
}

func (this *AddSignatureParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Address[:])
	sink.WriteVarBytes(this.Subject)
	sink.WriteVarBytes(this.Signature)
}

func (this *AddSignatureParam) Deserialization(source *common.ZeroCopySource) error {

	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}

	subject, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize subject error")
	}

	signature, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize signature error")
	}

	this.Address = addr
	this.Subject = subject
	this.Signature = signature
	return nil
}
