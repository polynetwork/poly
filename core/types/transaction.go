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

package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ontio/multi-chain/common/constants"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/payload"
	"github.com/ontio/ontology-crypto/keypair"
)

const MAX_TX_SIZE = 1024 * 1024 // The max size of a transaction to prevent DOS attacks

type CoinType byte

const (
	ONG CoinType = iota
	ETH
	BTC
)

type Transaction struct {
	Version    byte
	TxType     TransactionType
	Nonce      uint32
	ChainID    uint64
	GasLimit   uint64
	GasPrice   uint64
	Payload    Payload
	Attributes []byte //this must be 0 now, Attribute Array length use VarUint encoding, so byte is enough for extension
	Payer      common.Address
	CoinType   CoinType
	Sigs       []Sig

	Raw []byte // raw transaction data

	hash       common.Uint256
	SignedAddr []common.Address // this is assigned when passed signature verification
}

func (tx *Transaction) SerializeUnsigned(sink *common.ZeroCopySink) error {
	if tx.Version > CURR_TX_VERSION {
		return fmt.Errorf("invalid tx version:%d", tx.Version)
	}
	sink.WriteByte(tx.Version)
	sink.WriteByte(byte(tx.TxType))
	sink.WriteUint32(tx.Nonce)
	sink.WriteUint64(tx.ChainID)
	sink.WriteUint64(tx.GasLimit)
	sink.WriteUint64(tx.GasPrice)
	//Payload
	if tx.Payload == nil {
		return errors.New("transaction payload is nil")
	}
	switch pl := tx.Payload.(type) {
	case *payload.InvokeCode:
		pl.Serialization(sink)
	default:
		return errors.New("wrong transaction payload type")
	}
	if len(tx.Attributes) > MAX_ATTRIBUTES_LEN {
		return fmt.Errorf("attributes length %d over max length %d", tx.Attributes, MAX_ATTRIBUTES_LEN)
	}
	sink.WriteVarBytes(tx.Attributes)
	sink.WriteAddress(tx.Payer)
	sink.WriteByte(byte(tx.CoinType))
	return nil
}

// Serialize the Transaction
func (tx *Transaction) Serialization(sink *common.ZeroCopySink) error {
	if err := tx.SerializeUnsigned(sink); err != nil {
		return err
	}

	sink.WriteVarUint(uint64(len(tx.Sigs)))
	for _, sig := range tx.Sigs {
		if err := sig.Serialize(sink); err != nil {
			return err
		}
	}

	return nil
}

// if no error, ownership of param raw is transfered to Transaction
func TransactionFromRawBytes(raw []byte) (*Transaction, error) {
	if len(raw) > MAX_TX_SIZE {
		return nil, errors.New("execced max transaction size")
	}
	source := common.NewZeroCopySource(raw)
	tx := &Transaction{}
	err := tx.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Transaction has internal reference of param `source`
func (tx *Transaction) Deserialization(source *common.ZeroCopySource) error {
	pstart := source.Pos()
	if err := tx.DeserializationUnsigned(source); err != nil {
		return err
	}
	pos := source.Pos()
	lenUnsigned := pos - pstart
	source.BackUp(lenUnsigned)
	rawUnsigned, eof := source.NextBytes(lenUnsigned)
	if eof {
		return fmt.Errorf("read unsigned code error")
	}
	temp := sha256.Sum256(rawUnsigned)
	tx.hash = sha256.Sum256(temp[:])

	l, eof := source.NextVarUint()
	if eof {
		return errors.New("[Deserialization] read sigs length error")
	}
	sigs := make([]Sig, l)
	for i := 0; i < int(l); i++ {
		var sig Sig
		if err := sig.Deserialize(source); err != nil {
			return err
		}
		sigs[i] = sig
	}
	tx.Sigs = sigs
	pend := source.Pos()
	lenAll := pend - pstart
	if lenAll > MAX_TX_SIZE {
		return fmt.Errorf("execced max transaction size:%d", lenAll)
	}
	source.BackUp(lenAll)
	tx.Raw, _ = source.NextBytes(lenAll)
	return nil
}

func (tx *Transaction) DeserializationUnsigned(source *common.ZeroCopySource) error {
	var eof bool
	tx.Version, eof = source.NextByte()
	if eof {
		return errors.New("[deserializationUnsigned] read version error")
	}
	if tx.Version > CURR_TX_VERSION {
		return fmt.Errorf("[deserializationUnsigned] tx version %d over max version %d", tx.Version, CURR_TX_VERSION)
	}
	txType, eof := source.NextByte()
	if eof {
		return errors.New("[deserializationUnsigned] read txType error")
	}
	tx.TxType = TransactionType(txType)
	tx.Nonce, eof = source.NextUint32()
	if eof {
		return errors.New("[deserializationUnsigned] read nonce error")
	}
	tx.ChainID, eof = source.NextUint64()
	if eof {
		return errors.New("[deserializationUnsigned] read chainid error")
	}
	tx.GasLimit, eof = source.NextUint64()
	if eof {
		return errors.New("[deserializationUnsigned] read gaslimit error")
	}
	tx.GasPrice, eof = source.NextUint64()
	if eof {
		return errors.New("[deserializationUnsigned] read gasprice error")
	}

	switch tx.TxType {
	case Invoke:
		pl := new(payload.InvokeCode)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl
	default:
		return fmt.Errorf("unsupported tx type %v", tx.Type())
	}
	tx.Attributes, eof = source.NextVarBytes()
	if eof {
		return errors.New("[deserializationUnsigned] read attributes error")
	}
	if len(tx.Attributes) > MAX_ATTRIBUTES_LEN {
		return fmt.Errorf("[deserializationUnsigned] attributes length %d over max limit %d", tx.Attributes, MAX_ATTRIBUTES_LEN)
	}
	tx.Payer, eof = source.NextAddress()
	if eof {
		return errors.New("[deserializationUnsigned] read payer error")
	}
	coinType, eof := source.NextByte()
	if eof {
		return errors.New("[deserializationUnsigned] read coinType error")
	}
	tx.CoinType = CoinType(coinType)
	if tx.CoinType != ONG {
		return errors.New("[deserializationUnsigned] unsupported coinType")
	}
	return nil
}

type Sig struct {
	SigData [][]byte
	PubKeys []keypair.PublicKey
	M       uint16
}

func (this *Sig) Serialize(sink *common.ZeroCopySink) error {
	if len(this.PubKeys) == 0 {
		return errors.New("[Sig Serialize] no pubkeys in sig")
	}
	sink.WriteUint16(uint16(len(this.SigData)))
	for _, v := range this.SigData {
		sink.WriteVarBytes(v)
	}
	sink.WriteUint16(uint16(len(this.PubKeys)))
	for _, v := range this.PubKeys {
		key := keypair.SerializePublicKey(v)
		sink.WriteVarBytes(key)
	}
	sink.WriteUint16(this.M)
	return nil
}

func (this *Sig) Deserialize(source *common.ZeroCopySource) error {
	l, eof := source.NextUint16()
	if eof {
		return errors.New("[Sig] deserialize read sigData length error")
	}
	sigData := make([][]byte, l)
	for i := 0; i < int(l); i++ {
		data, eof := source.NextVarBytes()
		if eof {
			return errors.New("[Sig] deserialize read sigData error")
		}
		sigData[i] = data
	}
	l, eof = source.NextUint16()
	if eof {
		return errors.New("[Sig] deserialize read publicKey length error")
	}
	this.SigData = sigData
	pubKeys := make([]keypair.PublicKey, l)
	for i := 0; i < int(l); i++ {
		data, eof := source.NextVarBytes()
		if eof {
			return errors.New("[Sig] deserialize read sigData error")
		}
		pk, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return err
		}
		pubKeys[i] = pk
	}
	this.PubKeys = pubKeys
	m, eof := source.NextUint16()
	if eof {
		return errors.New("[Sig] deserialize read M error")
	}
	this.M = m
	return nil
}

func (self *Transaction) GetSignatureAddresses() ([]common.Address, error) {
	if len(self.SignedAddr) == 0 {
		addrs := make([]common.Address, 0, len(self.Sigs))
		for _, prog := range self.Sigs {
			if len(prog.PubKeys) == 0 {
				return nil, errors.New("[GetSignatureAddresses] no public key")
			} else if len(prog.PubKeys) == 1 {
				buf := keypair.SerializePublicKey(prog.PubKeys[0])
				addrs = append(addrs, common.AddressFromVmCode(buf))
			} else {
				sink := common.NewZeroCopySink(nil)
				if err := EncodeMultiPubKeyProgramInto(sink, prog.PubKeys, prog.M); err != nil {
					return nil, err
				}
				addrs = append(addrs, common.AddressFromVmCode(sink.Bytes()))
			}
		}
		self.SignedAddr = addrs
	}
	return self.SignedAddr, nil
}

type TransactionType byte

const (
	Deploy TransactionType = 0xd0
	Invoke TransactionType = 0xd1
)

// Payload define the func for loading the payload data
// base on payload type which have different structure
type Payload interface {
	Deserialization(source *common.ZeroCopySource) error

	Serialization(sink *common.ZeroCopySink)
}

func (tx *Transaction) ToArray() []byte {
	sink := new(common.ZeroCopySink)
	tx.Serialization(sink)
	return sink.Bytes()
}

func (tx *Transaction) Hash() common.Uint256 {
	return tx.hash
}

func (tx *Transaction) Type() common.InventoryType {
	return common.TRANSACTION
}

func EncodeMultiPubKeyProgramInto(sink *common.ZeroCopySink, pubkeys []keypair.PublicKey, m uint16) error {
	n := len(pubkeys)
	if !(1 <= m && int(m) <= n && n > 1 && n <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		return errors.New("wrong multi-sig param")
	}
	pubkeys = keypair.SortPublicKeys(pubkeys)

	sink.WriteUint16(uint16(len(pubkeys)))
	for _, pubkey := range pubkeys {
		key := keypair.SerializePublicKey(pubkey)
		sink.WriteVarBytes(key)
	}
	sink.WriteUint16(m)

	return nil
}
