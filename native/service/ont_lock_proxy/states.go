package ont_lock_proxy

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
)

// Args for lock and unlock
type Args struct {
	ToAddress []byte
	Value     uint64
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.ToAddress)
	sink.WriteVarUint(this.Value)
}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	var eof bool

	this.ToAddress, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args.Deserialization error: decode ToAddress var bytes ")
	}

	this.Value, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("Args.Deserialization error: decode Value uint64 ")
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
		return fmt.Errorf("LockParam.Deserialization error: %s", "decode ToChainID error")
	}
	this.FromAddress, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("LockParam.Deserialization error: %s", "decode FromAddress error")
	}
	this.Fee, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("LockParam.Deserialization error: %s", "decode Fee error")
	}
	err := this.Args.Deserialization(source)
	if err != nil {
		return err
	}
	return nil
}
