package test

import (
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/cmd/utils"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/payload"
	"github.com/polynetwork/poly/core/signature"
	"github.com/polynetwork/poly/core/types"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerifyTx(t *testing.T) {
	acc1 := account.NewAccount("123")

	tx := &types.Transaction{
		Version:    0,
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte{},
	}
	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	assert.NoError(t, err)

	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	assert.NoError(t, err)

	err = utils.SignTransaction(acc1, tx)
	assert.NoError(t, err)

	hash := tx.Hash()
	err = signature.Verify(acc1.PublicKey, hash.ToArray(), tx.Sigs[0].SigData[0])
	assert.NoError(t, err)

	addr, err := tx.GetSignatureAddresses()
	assert.NoError(t, err)
	assert.Equal(t, acc1.Address, addr[0])
}

func TestMultiVerifyTx(t *testing.T) {
	acc1 := account.NewAccount("123")
	acc2 := account.NewAccount("123")
	acc3 := account.NewAccount("123")

	accAddr, err := types.AddressFromMultiPubKeys([]keypair.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, 2)
	assert.NoError(t, err)
	tx := &types.Transaction{
		Version:    0,
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte{},
	}
	sink := common.NewZeroCopySink(nil)
	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	assert.NoError(t, err)

	err = utils.MultiSigTransaction(tx, 2, []keypair.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, acc1)
	assert.NoError(t, err)

	err = utils.MultiSigTransaction(tx, 2, []keypair.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, acc2)
	assert.NoError(t, err)

	hash := tx.Hash()
	err = signature.VerifyMultiSignature(hash.ToArray(), tx.Sigs[0].PubKeys, 2, tx.Sigs[0].SigData)
	assert.NoError(t, err)

	addr, err := tx.GetSignatureAddresses()
	assert.NoError(t, err)
	assert.Equal(t, accAddr, addr[0])
}
