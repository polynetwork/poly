package ripple

import (
	"encoding/json"
	"fmt"
	"github.com/polynetwork/ripple-sdk/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestJsonMarshall(t *testing.T) {
	txJson := "{\"TransactionType\":\"Payment\",\"Account\":\"rsHYGX2AoQ4tXqFywzEeeTDgXFTUfL1Fw9\",\"Sequence\":25336393,\"Fee\":\"150\",\"SigningPubKey\":\"\",\"Signers\":[{\"Account\":\"rLi6oSF38EdP7mzhdccyxhfd8vp8FWbsWF\",\"TxnSignature\":\"3044022048B1FD1B48B149B9E7A66344F758E7992C331D5EFE9A81F3F4D52477C5DBEBD50220453DC7B5A4E617CC59B15F887A7579F2B0BA8F14A65A6C416EB7C1D8A610204A\",\"SigningPubKey\":\"038B71C30DF7D4E9259732247AF169CCFACA1C0210784CEBD2884C0003B91CF33A\"}],\"Memos\":[{\"Memo\":{\"MemoType\":\"706F6C7968617368\",\"MemoData\":\"3E7C59E3954DEE9116A8148EC5CDCDB22485D55A62161B892F68ABDE4BF1A618\",\"MemoFormat\":\"\"}}],\"hash\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"Destination\":\"rT4vRkeJsgaq7t6TVJJPsbrQp5oKMGRfN\",\"Amount\":\"1000000\"}"
	payment := new(types.MultisignPayment)
	err := json.Unmarshal([]byte(txJson), payment)
 	assert.Nil(t, err)
	for _, s := range payment.Signers {
		fmt.Println(s.Signer.Account)
	}
}

func TestStringPrecise(t *testing.T) {
	fee_temp := new(big.Int).SetUint64(150)
	fee := ToStringByPrecise(fee_temp, 6)
	assert.Equal(t, fee, "0.00015")
}