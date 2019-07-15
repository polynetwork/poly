package test

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"testing"
)

func TestBufer(t *testing.T) {
	bf := common.NewZeroCopySink(nil)

	bf.WriteUint64(uint64(22))
	fmt.Println("buf:", bf.Bytes())
}
