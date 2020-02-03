package ont

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
)

// Transfers
type Transfers struct {
	States []State
}

func (this *Transfers) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.States)))
	for _, v := range this.States {
		v.Serialization(sink)
	}
}

func (this *Transfers) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("Transfers.Deserialization error: %s", "NextVarUint error")
	}
	for i := 0; uint64(i) < n; i++ {
		var state State
		if err := state.Deserialization(source); err != nil {
			return err
		}
		this.States = append(this.States, state)
	}
	return nil
}

type State struct {
	From  common.Address
	To    common.Address
	Value uint64
}

func (this *State) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.From)
	sink.WriteAddress(this.To)
	sink.WriteVarUint(this.Value)
}

func (this *State) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.From, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode From address error")
	}

	this.To, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode To address error")
	}

	this.Value, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode Value error")
	}

	return nil
}

type TransferFrom struct {
	Sender common.Address
	From   common.Address
	To     common.Address
	Value  uint64
}

func (this *TransferFrom) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.Sender)
	sink.WriteAddress(this.From)
	sink.WriteAddress(this.To)
	sink.WriteVarUint(this.Value)
}

func (this *TransferFrom) Deserialization(source *common.ZeroCopySource) error {
	var eof bool

	this.Sender, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode Sender address error")
	}

	this.From, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode From address error")
	}

	this.To, eof = source.NextAddress()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode To address error")
	}

	this.Value, eof = source.NextVarUint()
	if eof {
		return fmt.Errorf("State.Deserialization error: %s", "decode Value error")
	}

	return nil
}
