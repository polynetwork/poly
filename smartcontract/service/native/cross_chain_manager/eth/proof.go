package eth

import (
	"math/big"
)

type Proof struct {
	AssetAddress string
	FromAddress  string
	ToChainID    uint64
	ToAddress    string
	Amount       *big.Int
	Decimal      int
}

func (this *Proof) Deserialize(raw string) error {
	//todo add deserialize logic
	return nil
}
