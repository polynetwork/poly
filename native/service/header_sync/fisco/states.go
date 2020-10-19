package fisco

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/tjfoc/gmsm/sm2"
)

type FiscoRoot struct {
	RootCA *sm2.Certificate
}

func (root *FiscoRoot) Serialization(sink *pcom.ZeroCopySink) {
	sink.WriteVarBytes(root.RootCA.Raw)
}

func (root *FiscoRoot) Deserialization(source *pcom.ZeroCopySource) error {
	var (
		err error
	)
	raw, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("failed to deserialize RootCA")
	}
	root.RootCA, err = sm2.ParseCertificate(raw)
	if err != nil {
		return fmt.Errorf("failed to parse cert: %v", err)
	}

	return nil
}
