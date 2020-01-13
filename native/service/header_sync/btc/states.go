package btc

import (
	"github.com/btcsuite/btcd/wire"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"bytes"
	"math/big"
	"encoding/binary"
	"github.com/ontio/eth_tools/log"
)


type BTCBlockHeader struct {
	Height uint64
	BlockHeader *wire.BlockHeader
}



func (this *BTCBlockHeader) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.Height)
	buf := bytes.NewBuffer(nil)
	this.BlockHeader.Serialize(buf)
	sink.WriteVarBytes(buf.Bytes())
}

func (this *BTCBlockHeader) Deserialization(source *common.ZeroCopySource) error {
	height, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("BTCBlockHeader deserialize height error")
	}
	blockHeaderBs, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BTCBlockHeader deserialize BlockHeader error")
	}
	this.Height = height
	var blockHeader *wire.BlockHeader
	blockHeader.Deserialize(bytes.NewBuffer(blockHeaderBs))
	this.BlockHeader = blockHeader
	return nil
}

type StoredHeader struct {
	Header    wire.BlockHeader
	Height    uint32
	totalWork *big.Int
}



//TODO, format serialize and deserialize rule
/*----- header serialization ------- */
/* byteLength   desc          at offset
   80	       header	           0
    4	       height             80
   32	       total work         84
*/


func (this *StoredHeader)Serialization(sink *common.ZeroCopySink) {
	buf := bytes.NewBuffer(nil)
	this.Header.Serialize(buf)
	sink.WriteVarBytes(buf.Bytes())
	sink.WriteUint32(this.Height)
	biBytes := this.totalWork.Bytes()
	pad := make([]byte, 32-len(biBytes))
	//serializedBI := append(pad, biBytes...)
	sink.WriteVarBytes(append(pad, biBytes...))
}

func (this *StoredHeader)Deserialization(source *common.ZeroCopySource) error {
	buf, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("StoredHeader get header bytes error")
	}
	blockHeader := new(wire.BlockHeader)
	err := blockHeader.Deserialize(bytes.NewBuffer(buf))
	if err != nil {
		log.Error("deserialize wire.blockheader error: ", err)
		return fmt.Errorf("StoredHeader deserialize header error")
	}
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("StoredHeader get height error")
	}
	totalWorkBytes, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("StoredHeader get total work bytes error")
	}
	totalWork := new(big.Int)
	totalWork.SetBytes(totalWorkBytes)
	this.Header = *blockHeader
	this.Height = height
	this.totalWork = totalWork
	return nil
}



func serializeHeader(sh StoredHeader) ([]byte, error) {
	var buf bytes.Buffer
	err := sh.Header.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.BigEndian, sh.Height)
	if err != nil {
		return nil, err
	}
	biBytes := sh.totalWork.Bytes()
	pad := make([]byte, 32-len(biBytes))
	serializedBI := append(pad, biBytes...)
	buf.Write(serializedBI)
	return buf.Bytes(), nil
}

func deserializeHeader(b []byte) (sh StoredHeader, err error) {
	r := bytes.NewReader(b)
	hdr := new(wire.BlockHeader)
	err = hdr.Deserialize(r)
	if err != nil {
		return sh, err
	}
	var height uint32
	err = binary.Read(r, binary.BigEndian, &height)
	if err != nil {
		return sh, err
	}
	biBytes := make([]byte, 32)
	_, err = r.Read(biBytes)
	if err != nil {
		return sh, err
	}
	bi := new(big.Int)
	bi.SetBytes(biBytes)
	sh = StoredHeader{
		Header:    *hdr,
		Height:    height,
		totalWork: bi,
	}
	return sh, nil
}


