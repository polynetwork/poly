package cross_chain_manager

import (
	"fmt"
	"testing"

	"github.com/polynetwork/poly/common"
)


func TestWhite(t *testing.T) {
	a := new(WhiteAddressParam)

	a.Addresses = []string{"0x111222"}

	sink := new(common.ZeroCopySink)
	a.Serialization(sink)

	b := new(WhiteAddressParam)

	b.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	fmt.Printf("%v", b)
}