package pixiechain

// PixieChain Will active a hard fork in the future.
/*
import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/polynetwork/poly/native/service/header_sync/eth"
)

var (
	isTest                 bool
	testNextHardforkHeight uint64
)

// isNext returns true if it's the next hard fork of PixieChain
func isNext(h *eth.Header) bool {
	if isTest {
		return h.Number.Uint64() >= testNextHardforkHeight
	}
	return h.BaseFee != nil || h.Number.Uint64() >= config.GetPixieChainNextHardforkHeight(config.DefConfig.P2PNode.NetworkId)
}

// VerifyEip1559Header verifies some header attributes which were changed in EIP-1559,
// - gas limit check
// - basefee check
func VerifyEip1559Header(parent, header *eth.Header) error {
	// Verify that the gas limit remains within allowed bounds
	parentGasLimit := parent.GasLimit

	if err := VerifyGaslimit(parentGasLimit, header.GasLimit); err != nil {
		return err
	}
	// Verify the header is not malformed
	if header.BaseFee == nil {
		return fmt.Errorf("header is missing baseFee")
	}
	// Verify the baseFee is correct based on the parent header.
	expectedBaseFee := CalcBaseFee(parent)
	if header.BaseFee.Cmp(expectedBaseFee) != 0 {
		return fmt.Errorf("invalid baseFee: have %s, want %s, parentBaseFee %s, parentGasUsed %d",
			expectedBaseFee, header.BaseFee, parent.BaseFee, parent.GasUsed)
	}
	return nil
}

// CalcBaseFee calculates the basefee of the header.
func CalcBaseFee(parent *eth.Header) *big.Int {
	return common.Big0
}
*/
