package test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/types"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransaction(t *testing.T) {
	acc := account.NewAccount("123")
	header := types.Header{
		Version:          0,
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStatesRoot:  common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        12,
		Height:           12,
		ConsensusData:    12,
		ConsensusPayload: []byte{1, 2},
		NextBookkeeper:   common.ADDRESS_EMPTY,
		Bookkeepers:      []keypair.PublicKey{acc.PublicKey},
		SigData:          [][]byte{{1, 2, 3}},
	}

	buf := bytes.NewBuffer(nil)
	err := header.Serialize(buf)

	assert.NoError(t, err)

	var h types.Header
	err = h.Deserialize(buf)

	assert.NoError(t, err)

	assert.Equal(t, header, h)
}

func BenchmarkT1(b *testing.B) {
	acc := account.NewAccount("123")
	header := Header{
		Version:          0,
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStatesRoot:  common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        12,
		Height:           12,
		ConsensusData:    12,
		ConsensusPayload: []byte{1, 2},
		NextBookkeeper:   common.ADDRESS_EMPTY,
		Bookkeepers:      []keypair.PublicKey{acc.PublicKey},
		SigData:          [][]byte{{1, 2, 3}},
	}
	buf := NewZeroCopy(nil)
	header.Serialization(buf)
	for i := 0; i < b.N; i++ {
		var h Header
		err := h.Deserialization(NewZeroCopy(buf.Bytes()))
		assert.NoError(b, err)
	}
}

func BenchmarkT3(b *testing.B) {
	acc := account.NewAccount("123")
	header := types.Header{
		Version:          0,
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStatesRoot:  common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        12,
		Height:           12,
		ConsensusData:    12,
		ConsensusPayload: []byte{1, 2},
		NextBookkeeper:   common.ADDRESS_EMPTY,
		Bookkeepers:      []keypair.PublicKey{acc.PublicKey},
		SigData:          [][]byte{{1, 2, 3}},
	}

	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(nil)
		header.Serialize(buf)
		var h types.Header
		h.Deserialize(buf)
	}
}

func BenchmarkT2(b *testing.B) {
	acc := account.NewAccount("123")
	header := types.Header{
		Version:          0,
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStatesRoot:  common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        12,
		Height:           12,
		ConsensusData:    12,
		ConsensusPayload: []byte{1, 2},
		NextBookkeeper:   common.ADDRESS_EMPTY,
		Bookkeepers:      []keypair.PublicKey{acc.PublicKey},
		SigData:          [][]byte{{1, 2, 3}},
	}
	for i := 0; i < b.N; i++ {
		buf := common.NewZeroCopySink(nil)
		header.Serialization(buf)
		var h types.Header
		h.Deserialization(common.NewZeroCopySource(buf.Bytes()))
	}
}

type Header struct {
	Version          uint32
	ChainID          uint64
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	CrossStatesRoot  common.Uint256
	BlockRoot        common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload []byte
	NextBookkeeper   common.Address

	//Program *program.Program
	Bookkeepers []keypair.PublicKey
	SigData     [][]byte

	hash *common.Uint256
}

func (bd *Header) Serialization(sink *ZeroCopy) error {
	bd.serializationUnsigned(sink)
	sink.WriteVarUint(uint64(len(bd.Bookkeepers)))

	for _, pubkey := range bd.Bookkeepers {
		sink.WriteVarBytes(keypair.SerializePublicKey(pubkey))
	}

	sink.WriteVarUint(uint64(len(bd.SigData)))
	for _, sig := range bd.SigData {
		sink.WriteVarBytes(sig)
	}

	return nil
}

//Serialize the blockheader data without program
func (bd *Header) serializationUnsigned(sink *ZeroCopy) {
	if bd.Version > types.CURR_HEADER_VERSION {
		panic(fmt.Errorf("invalid header %d over max version:%d", bd.Version, types.CURR_HEADER_VERSION))
	}
	sink.WriteUint32(bd.Version)
	sink.WriteUint64(bd.ChainID)
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteBytes(bd.CrossStatesRoot[:])
	sink.WriteBytes(bd.BlockRoot[:])
	sink.WriteUint32(bd.Timestamp)
	sink.WriteUint32(bd.Height)
	sink.WriteUint64(bd.ConsensusData)
	sink.WriteVarBytes(bd.ConsensusPayload)
	sink.WriteBytes(bd.NextBookkeeper[:])
}

func (bd *Header) Deserialization(source *ZeroCopy) error {
	err := bd.deserializationUnsigned(source)
	if err != nil {
		return err
	}

	n, eof := source.NextVarUint()
	if eof {
		return errors.New("[Header] deserialize bookkeepers length error")
	}

	for i := 0; i < int(n); i++ {
		buf, eof := source.NextVarBytes()
		if eof {
			return errors.New("[Header] deserialize bookkeepers public key error")
		}
		pubkey, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		bd.Bookkeepers = append(bd.Bookkeepers, pubkey)
	}

	m, eof := source.NextVarUint()
	if eof {
		return errors.New("[Header] deserialize sigData length error")
	}

	for i := 0; i < int(m); i++ {
		sig, eof := source.NextVarBytes()
		if eof {
			return errors.New("[Header] deserialize sigData error")
		}
		bd.SigData = append(bd.SigData, sig)
	}

	return nil
}

func (bd *Header) deserializationUnsigned(source *ZeroCopy) error {
	var eof bool
	bd.Version, eof = source.NextUint32()
	if eof {
		return errors.New("[Header] read version error")
	}
	if bd.Version > types.CURR_HEADER_VERSION {
		return fmt.Errorf("[Header] header version %d over max version %d", bd.Version, types.CURR_HEADER_VERSION)
	}
	bd.ChainID, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read chainID error")
	}
	bd.PrevBlockHash, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read prevBlockHash error")
	}
	bd.TransactionsRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read transactionsRoot error")
	}
	bd.CrossStatesRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read crossStatesRoot error")
	}
	bd.BlockRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read blockRoot error")
	}
	bd.Timestamp, eof = source.NextUint32()
	if eof {
		return errors.New("[Header] read timestamp error")
	}
	bd.Height, eof = source.NextUint32()
	if eof {
		return errors.New("[Header] read height error")
	}
	bd.ConsensusData, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read consensusData error")
	}
	bd.ConsensusPayload, eof = source.NextVarBytes()
	if eof {
		return errors.New("[Header] read consensusPayload error")
	}
	bd.NextBookkeeper, eof = source.NextAddress()
	if eof {
		return errors.New("[Header] read nextBookkeeper error")
	}
	return nil
}
