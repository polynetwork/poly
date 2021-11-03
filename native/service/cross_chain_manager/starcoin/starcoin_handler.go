package starcoin

import "github.com/polynetwork/poly/native"

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// MakeDepositProposal ...
func (h *Handler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	return nil, nil
}
