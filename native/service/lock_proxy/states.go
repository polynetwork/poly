package lock_proxy

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"io"
	"math/big"
)

// Args for lock and unlock
type Args struct {
	TargetAssetHash []byte
	ToAddress       []byte
	Value           uint64
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TargetAssetHash)
	sink.WriteVarBytes(this.ToAddress)
	sink.WriteVarUint(this.Value)
}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	if this.TargetAssetHash, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("Args.Deserialization NextVarBytes TargetAssetHash error:%s", io.ErrUnexpectedEOF)
	}

	if this.ToAddress, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("Args.Deserialization NextVarBytes ToAddress error:%s", io.ErrUnexpectedEOF)
	}

	if this.Value, eof = source.NextVarUint(); eof {
		return fmt.Errorf("Args.Deserialization NextVarUint Value error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}

type LockParam struct {
	SourceAssetHash common.Address
	FromAddress     common.Address
	ToChainID       uint64
	ToAddress       []byte
	Value           uint64
}

func (this *LockParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.SourceAssetHash)
	sink.WriteAddress(this.FromAddress)
	sink.WriteVarUint(this.ToChainID)
	sink.WriteVarBytes(this.ToAddress)
	sink.WriteVarUint(this.Value)
}

func (this *LockParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	if this.SourceAssetHash, eof = source.NextAddress(); eof {
		return fmt.Errorf("LockParam.Deserialization NextAddress AssetHash error:%s", io.ErrUnexpectedEOF)
	}
	if this.FromAddress, eof = source.NextAddress(); eof {
		return fmt.Errorf("LockParam.Deserialization NextAddress FromAddress error:%s", io.ErrUnexpectedEOF)
	}
	if this.ToChainID, eof = source.NextVarUint(); eof {
		return fmt.Errorf("LockParam.Deserialization NextVarUint ToChainID error:%s", io.ErrUnexpectedEOF)
	}
	if this.ToAddress, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("LockParam.Deserialization NextVarBytes ToAddress error:%s", io.ErrUnexpectedEOF)
	}
	if this.Value, eof = source.NextVarUint(); eof {
		return fmt.Errorf("LockParam.Deserialization NextVarUint Value error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}

type UnlockParam struct {
	ArgsBs             []byte
	FromContractHashBs []byte
	FromChainId        uint64
}

func (this *UnlockParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.ArgsBs)
	sink.WriteVarBytes(this.FromContractHashBs)
	sink.WriteVarUint(this.FromChainId)
}

func (this *UnlockParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	if this.ArgsBs, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("UnlockParam.Deserialization NextVarBytes ArgsBs error:%s", io.ErrUnexpectedEOF)
	}
	if this.FromContractHashBs, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("UnlockParam.Deserialization NextVarBytes FromContractHashBs error:%s", io.ErrUnexpectedEOF)
	}
	if this.FromChainId, eof = source.NextVarUint(); eof {
		return fmt.Errorf("UnlockParam.Deserialization NextVarUint FromChainId error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}

type BindProxyParam struct {
	TargetChainId uint64
	TargetHash    []byte
}

func (this *BindProxyParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.TargetChainId)
	sink.WriteVarBytes(this.TargetHash)
}

func (this *BindProxyParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	if this.TargetChainId, eof = source.NextVarUint(); eof {
		return fmt.Errorf("BindProxyParam.Deserialization NextVarUint TargetChainId error:%s", io.ErrUnexpectedEOF)
	}
	if this.TargetHash, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("BindProxyParam.Deserialization NextVarBytes TargetHash error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}

type BindAssetParam struct {
	SourceAssetHash    common.Address
	TargetChainId      uint64
	TargetAssetHash    []byte
	Limit              *big.Int
	IsTargetChainAsset bool
}

func (this *BindAssetParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.SourceAssetHash)
	sink.WriteVarUint(this.TargetChainId)
	sink.WriteVarBytes(this.TargetAssetHash)
	sink.WriteVarBytes(this.Limit.Bytes())
	sink.WriteBool(this.IsTargetChainAsset)
}

func (this *BindAssetParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	if this.SourceAssetHash, eof = source.NextAddress(); eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextAddress SourceAssetAddress error:%s", io.ErrUnexpectedEOF)
	}
	if this.TargetChainId, eof = source.NextVarUint(); eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextVarUint TargetChainId error:%s", io.ErrUnexpectedEOF)
	}
	if this.TargetAssetHash, eof = source.NextVarBytes(); eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextVarBytes TargetAssetHash error:%s", io.ErrUnexpectedEOF)
	}
	limitBigIntBs, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextVarBytes Limit error:%s", io.ErrUnexpectedEOF)
	}
	this.Limit = big.NewInt(0).SetBytes(limitBigIntBs)
	if this.IsTargetChainAsset, eof = source.NextBool(); eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextBool IsTargetChainAsset error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}
