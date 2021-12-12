package starcoin

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	stc "github.com/starcoinorg/starcoin-go/client"
	"github.com/starcoinorg/starcoin-go/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestVerifyEventProof(t *testing.T) {
	type args struct {
		proof   *TransactionInfoProof
		data    types.HashValue
		address []byte
	}
	transactionInfo := stc.TransactionInfo{
		BlockHash:        "0xe79298f969573db42415b949ec8daa4e19abc680aa8c7e02cc22a2ad9f02fbf2",
		BlockNumber:      "0",
		TransactionHash:  "0x4978f0c2a31e2b927041bb050fd710aacc9e0575f3f383d7b8439e5d3996c41c",
		TransactionIndex: 4,
		StateRootHash:    "0x7f534a6e7f8312658c6ff0184c070da25d3f99ed1c85ac44a50f516c66e16f5f",
		EventRootHash:    "0xad7b6bbaefc8bf183a927001ee195a2c854fbf778f763bc3f72d6c1cfcf83f24",
		GasUsed:          "0",
		Status:           "Executed",
	}
	bytes3, _ := hex.DecodeString("85d30ad69568d0269f7b2ad327b9acf3585869403a399155b23d29cb2bffc173")
	txnAccumulatorRoot := types.HashValue(bytes3)

	proof := TransactionInfoProof{
		TransactionInfo: transactionInfo,
		Proof:           "{\"siblings\":[\"0x49749e95c8df29d59fad98cc1cc5fb049a5cb016f400724e16c5389b0433be21\",\"0x71cafa34205ab0d7c0b087ab5407bf056b471b4cbe0273059f214ca37e421b9f\",\"0x272385029ccb0b5035d7189cbb6effbf8c78fabc1de2464e324d566cb85580ec\",\"0xc0d1b51dbbf6f194ae2a05af749ca3570e8b202db7e46f22f2b2ef853d487cde\"]}",
		eventIndex:      0,
		EventWithProof:  "{\"event\":{\"V0\":{\"key\":\"0x0e0000000000000000000000000000000000000000000001\",\"sequence_number\":0,\"type_tag\":{\"Struct\":{\"address\":\"0x00000000000000000000000000000001\",\"module\":\"Token\",\"name\":\"MintEvent\",\"type_params\":[]}},\"event_data\":[0,0,79,1,81,224,51,44,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,3,83,84,67,3,83,84,67]}},\"proof\":{\"siblings\":[\"0xe886665a7b12325d365c4991b3b7e306492f0ad565f692f66a0d41dce519e65a\",\"0xa94cc80c5a33aca23b6cfdc72c76acbd17adf9824acf8003e013ac02318c6580\",\"0xeb6ee4ac69b77d980aed9cc8f63a2cc5755e67efd77479e567f7d7a547f9c746\"]}}",
		StateWithProof:  "{\"state\":[32,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,24,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,24,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,24,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0],\"proof\":{\"account_state\":{\"blob\":[2,1,32,58,226,188,223,103,199,252,196,31,4,212,81,21,143,93,79,149,162,70,123,31,238,107,125,247,44,159,65,152,129,133,161,1,32,164,13,61,251,134,239,253,169,107,217,46,216,205,61,35,206,212,129,117,216,182,217,11,11,123,133,198,187,215,146,14,127]},\"account_proof\":{\"leaf\":[\"0xf1232a8cea9ae1d59147fcde30471108ec89c04579567e7aababc351b77ef1b6\",\"0xdfeab852f4f71945e94a5409c8283fcb32addef08ddd3d0ae9814de93a83efb3\"],\"siblings\":[\"0x360c7b79a7d2cc064e3af7befe431ac36613f62419afe2319a571ce2719f6202\",\"0x5350415253455f4d45524b4c455f504c414345484f4c4445525f484153480000\",\"0x5350415253455f4d45524b4c455f504c414345484f4c4445525f484153480000\"]},\"account_state_proof\":{\"leaf\":[\"0x9b079b5aef808c36133c95bacbc3d3411d1e8cce4fbc4b524ffab31202eaaa11\",\"0x91354856a1026edb028bc4aa72877a508483f3c9daef045b1d499cbf227fd0f7\"],\"siblings\":[\"0x318db4c3c46dc7b422e5dbda750952df1aa70387ab1331a8e0e39492a7fab69f\",\"0x5350415253455f4d45524b4c455f504c414345484f4c4445525f484153480000\",\"0x038cd2d75dbad3944f7476cb8e0c808e22a538e058ec0dd6a4508ba98d3a82b4\",\"0x77b69f571a7672871e87967a0d714613efa88ce20f28fa65b9f2c4b4748e2ee1\",\"0x8349b139ef1233c993be3376f07a5a9fa0b5973e22867c3142636d7d5d908e22\",\"0xcb174d26a2565684d3673ab2a01a3bf6273f38bf8d8c329beb85a9c3729449e1\"]}}}",
		accessPath:      "000000000000000000000000000000010100000000000000000000000000000001074163636f756e74074163636f756e7400",
	}
	eventData := []byte{0, 0, 79, 1, 81, 224, 51, 44, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 83, 84, 67, 3, 83, 84, 67}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"test event proof",
			args{
				proof: &proof,
				data:  txnAccumulatorRoot,
			},
			eventData,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyEventProof(tt.args.proof, tt.args.data, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyEventProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("VerifyEventProof() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_internalHash(t *testing.T) {
	type args struct {
		index    int
		elements []byte
		sibling  []byte
	}
	left, _ := hex.DecodeString("21b129ad214b45d4d5f46f6b3a1853181797d871de09cf4dc234195bebf0820f")
	right, _ := hex.DecodeString("1d8a1b360faa73a3e86a2fba321edad502035e915f3b75b71791fc68d052e27b")
	internal, _ := hex.DecodeString("17b38596329fb9bebb56bf9e890645e3f72b688749c9bbeca57b633406b17169")
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test internal hash",
			args: args{
				index:    0,
				elements: left,
				sibling:  right,
			},
			want: internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := internalHash(tt.args.index, tt.args.elements, tt.args.sibling); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent(t *testing.T) {
	event := "{\"V0\":{\"key\":\"0x0e0000000000000000000000000000000000000000000001\",\"sequence_number\":1}}"
	var contract ContractEvent
	err := json.Unmarshal([]byte(event), &contract)
	if err != nil {
		println(err)
	}
	proof := "{\"event\":{\"V0\":{\"key\":\"0x0e0000000000000000000000000000000000000000000001\",\"sequence_number\":0,\"type_tag\":{\"Struct\":{\"address\":\"0x00000000000000000000000000000001\",\"module\":\"Token\",\"name\":\"MintEvent\",\"type_params\":[]}},\"event_data\":[0,0,79,1,81,224,51,44,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,3,83,84,67,3,83,84,67]}},\"proof\":{\"siblings\":[\"0xe886665a7b12325d365c4991b3b7e306492f0ad565f692f66a0d41dce519e65a\",\"0xa94cc80c5a33aca23b6cfdc72c76acbd17adf9824acf8003e013ac02318c6580\",\"0xeb6ee4ac69b77d980aed9cc8f63a2cc5755e67efd77479e567f7d7a547f9c746\"]}}"
	var eventWith EventWithProof
	err = json.Unmarshal([]byte(proof), &eventWith)
	assert.Equal(t, eventWith.Event.V.TypeTag.Value.Address, "0x00000000000000000000000000000001")
}

func Test_verifyAccumulator(t *testing.T) {
	type args struct {
		proof        AccumulatorProof
		expectedRoot types.HashValue
		hash         types.HashValue
		index        int
	}
	sibling0, _ := hex.DecodeString("49749e95c8df29d59fad98cc1cc5fb049a5cb016f400724e16c5389b0433be21")
	sibling1, _ := hex.DecodeString("71cafa34205ab0d7c0b087ab5407bf056b471b4cbe0273059f214ca37e421b9f")
	sibling2, _ := hex.DecodeString("272385029ccb0b5035d7189cbb6effbf8c78fabc1de2464e324d566cb85580ec")
	sibling3, _ := hex.DecodeString("c0d1b51dbbf6f194ae2a05af749ca3570e8b202db7e46f22f2b2ef853d487cde")

	left, _ := hex.DecodeString("4978f0c2a31e2b927041bb050fd710aacc9e0575f3f383d7b8439e5d3996c41c")
	internal, _ := hex.DecodeString("85d30ad69568d0269f7b2ad327b9acf3585869403a399155b23d29cb2bffc173")

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test verifyAccumulator",
			args: args{
				index:        4,
				expectedRoot: internal,
				hash:         left,
				proof:        AccumulatorProof{siblings: []types.HashValue{types.HashValue(sibling0), types.HashValue(sibling1), types.HashValue(sibling2), types.HashValue(sibling3)}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verifyAccumulator(tt.args.proof, tt.args.expectedRoot, tt.args.hash, tt.args.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyAccumulator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verifyAccumulator() got = %v, want %v", got, tt.want)
			}
		})
	}
}
