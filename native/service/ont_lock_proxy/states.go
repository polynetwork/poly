package ont_lock_proxy

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"io"
)

// Args for lock and unlock
type Args struct {
	AssetHash []byte
	ToAddress []byte
	Value     uint64
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.AssetHash)
	sink.WriteVarBytes(this.ToAddress)
	sink.WriteVarUint(this.Value)
}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.AssetHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args.Deserialization error: decode AssetHash var bytes error:%s", io.ErrUnexpectedEOF)
	}

	this.ToAddress, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args.Deserialization error: decode ToAddress var bytes error:%s", io.ErrUnexpectedEOF)
	}

	this.Value, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("Args.Deserialization error: decode Value uint64 error:%s", io.ErrUnexpectedEOF)
	}

	return nil
}

type LockParam struct {
	ToChainID   uint64
	FromAddress common.Address
	Fee         uint64
	Args        Args
}

func (this *LockParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(this.ToChainID)
	sink.WriteAddress(this.FromAddress)
	sink.WriteVarUint(this.Fee)
	this.Args.Serialization(sink)
}

func (this *LockParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.ToChainID, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("LockParam.Deserialization ToChainID NextVarUint error:%s", io.ErrUnexpectedEOF)
	}
	this.FromAddress, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("LockParam.Deserialization FromAddress NextAddress error:%s", io.ErrUnexpectedEOF)
	}
	this.Fee, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("LockParam.Deserialization Fee NextVarUint error:%s", io.ErrUnexpectedEOF)
	}
	err := this.Args.Deserialization(source)
	if err != nil {
		return fmt.Errorf("LockParam.Deserialization Args Deserialization error:%s", err)
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
	this.TargetChainId, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("BindProxyParam.Deserialization NextVarUint TargetChainId error:%s", io.ErrUnexpectedEOF)
	}
	this.TargetHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("BindProxyParam.Deserialization NextVarBytes TargetHash error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}

type BindAssetParam struct {
	SourceAssetHash common.Address
	TargetChainId   uint64
	TargetAssetHash []byte
}

func (this *BindAssetParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.SourceAssetHash)
	sink.WriteVarUint(this.TargetChainId)
	sink.WriteVarBytes(this.TargetAssetHash)
}

func (this *BindAssetParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.SourceAssetHash, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextAddress SourceAssetAddress error:%s", io.ErrUnexpectedEOF)
	}
	this.TargetChainId, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextVarUint TargetChainId error:%s", io.ErrUnexpectedEOF)
	}
	this.TargetAssetHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("BindAssetParam.Deserialization NextVarBytes TargetAssetHash error:%s", io.ErrUnexpectedEOF)
	}
	return nil
}
