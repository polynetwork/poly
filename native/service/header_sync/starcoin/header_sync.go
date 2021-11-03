package starcoin

import "github.com/polynetwork/poly/native"

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// SyncGenesisHeader ...
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	return nil
}

func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	return nil
}

func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
