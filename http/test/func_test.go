/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package test

import (
	"bytes"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/merkle"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMerkleVerifier(t *testing.T) {
	type merkleProof struct {
		Type             string
		TransactionsRoot string
		BlockHeight      uint32
		CurBlockRoot     string
		CurBlockHeight   uint32
		TargetHashes     []string
	}
	proof := merkleProof{
		Type:             "MerkleProof",
		TransactionsRoot: "4b74e15973ce3964ba4a33ddaf92efbff922ea2225bca7676f62eab05829f11f",
		BlockHeight:      2,
		CurBlockRoot:     "a5094c1daeeceab46319ce62b600c68a7accc806bd9fe2fdb869560bf66b5251",
		CurBlockHeight:   6,
		TargetHashes: []string{
			"c7ac8087b4ce292d654001b1ab1bfe5e68fa6f7b8492a5b2f83560f8ac28f5fa",
			"5205a22b07c6072d60d28b41f1321ab993799d70693a3bb70bab7e58b49acc30",
			"c0de7f3035a7960450ec9a64e7835b958b0fec1ddb90cbeb0779073c0a9a8f53",
		},
	}

	verify := merkle.NewMerkleVerifier()
	var leaf_hash common.Uint256
	bys, _ := common.HexToBytes(proof.TransactionsRoot)
	leaf_hash.Deserialize(bytes.NewReader(bys))

	var root_hash common.Uint256
	bys, _ = common.HexToBytes(proof.CurBlockRoot)
	root_hash.Deserialize(bytes.NewReader(bys))

	var hashes []common.Uint256
	for _, v := range proof.TargetHashes {
		var hash common.Uint256
		bys, _ = common.HexToBytes(v)
		hash.Deserialize(bytes.NewReader(bys))
		hashes = append(hashes, hash)
	}
	res := verify.VerifyLeafHashInclusion(leaf_hash, proof.BlockHeight, hashes, root_hash, proof.CurBlockHeight+1)
	assert.Nil(t, res)

}

func TestTxDeserialize(t *testing.T) {
	bys, _ := common.HexToBytes("00d190ac06ff0000000000000000fddf0300000000000000000000000000000000000000000313496d706f72744f757465725472616e73666572fdb3030300000000000000ad000000fd7501fd720120366d36f82d953816ad92e1eddc83836d28ff6fa5089c07609137ad0e290875f208140000000000000014e553510e9f7eb980b3f225056d144529525cc5ec01000000000000000362746306756e6c6f636bfd1d01226d6a456f79794350734c7a4a3233784d58364d746931337a4d794e33366b7a6e3537ad42000000000000f15521023ac710e73e1410718530b2686ce47f12fa3c470a9eb6085976b70b01c64c9f732102c9dc4d8f419e325bbef0fe039ed6feaf2079a2ef7b27336ddb79be2ea6e334bf2102eac939f2f0873894d8bf0ef2f8bbdd32e4290cbf9632b59dee743529c0af9e802103378b4a3854c88cca8bfed2558e9875a144521df4a75ab37a206049ccef12be692103495a81957ce65e3359c114e6c2fe9f97568be491e3f24d6fa66cc542e360cd662102d43e29299971e802160a92cfcd4037e8ae83fb8f6af138684bebdc5686f3b9db21031e415c04cbc9b81fbee6e04d8c902e8f61109a2c9883a959ba528c52698c055a57ae142186fe74983e0016359c7c1b9063448fc8813b8700fd160200ad00000011cbe13241a6a8f9711d511222d9ba69943aa32ea2cf057945529ff0014bed3c0540211ce0c33a78753097e8f9bc59770ed175c41e7ec6b8cdda325bc18c593306dc77ea250699f3c11a819c6a04872cc29178d5d79d9d65834ab084988d6f48e7e04029f7776bcd13dadce2153bf487f72f7dc216fcb12648e29f1e868feb96c4823f5fcd9ce8c16d332fbab15fc4fbed337514ed0be9bc1a79aede0ff5aa7fcd621440d3afda11df726ce1113357ce78d50a3ef58c1fdf398beee5c13d934fa4ad5af9926ea3cb72844a84c9ac28824be66b84a9a2a300d6bb38893bdd13d95c519d87405265f6e83dc0b6a89d426ed2f6d0acd71b90e042d2deb3658965620e0af25c5ce22b192b7e74bd2249f80fe1388ee613b1d02ff0b5582beaffa719fe9765372440bae0e1f3103f0fed5716dbae10fe4bffe6688992820d504595521ad82d302c5842efd4ce7024922c7aa4f006ff83fc212e10f4cf909852a41ebb727c96e849830521023967bba3060bf8ade06d9bad45d02853f6c623e4d4f52d767eb56df4d364a99f210215865baab70607f4a2413a7a9ba95ab2c3c0202d5b7731c6824eef48e899fc9021035eb654bad6c6409894b9b42289a43614874c7984bde6b03aaf6fc1d0486d9d45210253ccfd439b29eca0fe90ca7c6eaa1f98572a054aa2d1d56e72ad96c466107a852103f1095289e7fddb882f1cb3e158acc1c30d9de606af21c97ba851821e8b6ea5350001010042051c33e385b99bac45dbf50e9e2074f399f38170cd570ddf7be125bbc908d5bf496e7ff97ace139b2774332e858a194af94249e0b23a8bd212f43247c3b1b6a9cc3f010023120503d101838807ec4079a46fe98d6bd9a0690abcbd8ce16e0fbc4520c7c7ef7885db0100")
	_, err := types.TransactionFromRawBytes(bys)
	assert.Nil(t, err)
}
func TestAddress(t *testing.T) {
	pubkey, _ := common.HexToBytes("120203a4e50edc1e59979442b83f327030a56bffd08c2de3e0a404cefb4ed2cc04ca3e")
	_, err := keypair.DeserializePublicKey(pubkey)
	assert.Nil(t, err)
}
func TestMultiPubKeysAddress(t *testing.T) {
	pubkey, _ := common.HexToBytes("120203a4e50edc1e59979442b83f327030a56bffd08c2de3e0a404cefb4ed2cc04ca3e")
	pk, err := keypair.DeserializePublicKey(pubkey)
	assert.Nil(t, err)

	pubkey2, _ := common.HexToBytes("12020225c98cc5f82506fb9d01bad15a7be3da29c97a279bb6b55da1a3177483ab149b")
	pk2, err := keypair.DeserializePublicKey(pubkey2)
	assert.Nil(t, err)

	_, err = types.AddressFromMultiPubKeys([]keypair.PublicKey{pk, pk2}, 1)
	assert.Nil(t, err)
}
