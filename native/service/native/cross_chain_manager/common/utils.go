package btc

import (
	"fmt"

	"encoding/hex"

	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native/service/native"
	crosscommon "github.com/ontio/multi-chain/native/service/native/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/native/governance"
	"github.com/ontio/ontology-crypto/keypair"
)

func ValidateVote(service *native.NativeService, vote *crosscommon.Vote) error {
	consesusPeers, err := governance.GetConsensusPeers(service)
	if err != nil {
		return fmt.Errorf("ValidateVote, governance.GetConsensusPeers error: %v", err)
	}
	sum := 0
	for _, v := range consesusPeers {
		b, err := hex.DecodeString(v)
		if err != nil {
			return fmt.Errorf("ValidateVote, hex.DecodeString consensus public key error: %v", err)
		}
		pk, err := keypair.DeserializePublicKey(b)
		if err != nil {
			return fmt.Errorf("ValidateVote, keypair.DeserializePublicKey consensus public key error: %v", err)
		}
		address := types.AddressFromPubKey(pk)
		_, ok := vote.VoteMap[address.ToBase58()]
		if ok {
			sum = sum + 1
		}
	}

	if sum != (2*len(consesusPeers)+2)/3 {
		return fmt.Errorf("ValidateVote, not enough vote")
	}
	return nil
}
