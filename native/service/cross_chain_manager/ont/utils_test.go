package ont

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/merkle"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStateroot(t *testing.T) {
	crossStatesRoot := "91d969e63a2cf0c0e9bc76ed5aa99d3d024696d249229baf1f6dc688967e240b"
	root, err := common.Uint256FromHexString(crossStatesRoot)
	if err != nil {
		fmt.Println("common.Uint256FromHexString", err)
	}
	proofHex := "80000000000000000107283730613439356165323865646233316432613661613164643439613261643863616335393736643001012a30783030303030303030303030303030303030303030303030303030303030303030303030303030303001022241476a44344d6f32356b7a6353747968317374703774586b55754d6f704434334e54020384"
	proof, err := hex.DecodeString(proofHex)
	if err != nil {
		fmt.Println("hex.DecodeString", err)
	}
	v := merkle.MerkleProve(proof, root)
	assert.NotNil(t, v)
}
