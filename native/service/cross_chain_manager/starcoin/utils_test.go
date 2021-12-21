package starcoin

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	stc "github.com/starcoinorg/starcoin-go/client"
	"github.com/starcoinorg/starcoin-go/types"
)

func TestVerifyEventProof(t *testing.T) {
	type args struct {
		proof   *TransactionInfoProof
		data    types.HashValue
		address []byte
	}
	transactionInfo := stc.TransactionInfo{
		BlockHash:              "0x429e6ad8617937da569474a3eca407ec54dd458b4d854f6bd577d152d1428478",
		BlockNumber:            "292517",
		TransactionHash:        "0xd328475d67b5e3b7fef2fc4c3fb694a8904ddf90d91de80796280f5187cc66f0",
		TransactionIndex:       1,
		TransactionGlobalIndex: "324294",
		StateRootHash:          "0x4df4cb2116088fc2da8cfbc2c47175b54ef4a59c3784ddc5982eff4afe6f706c",
		EventRootHash:          "0x081e3db4b09ac60312ec8deb66455b9cfc7aa4201582cfd896b0d3b24ab2a04d",
		GasUsed:                "621382",
		Status:                 "Executed",
	}
	var proofSiblings Siblings
	json.Unmarshal([]byte("{\"siblings\":[\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x5e9ef122f7f51479a3383f13301ad065ba1da4a2746ccec29527e81cf909e197\",\"0x795a33fb6327e2a07249975fb2b712fd242545ed1ff4aec6442aaabc9f14b844\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x295fe28e3363cc6203b3537a17e3c9c4d1fe842a3ccedde809f68a9b29a099a4\",\"0xf475b977f927ca0c487f9afaeac3e37cd9998c61cc686585e68d61f68cbf5202\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x30fe3ea6eb484e75084640e6cde700dd26550219eabd8447cc7e70c4285c04ea\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0xf6afb2287bd40fbc8057c567e04c9cfc48afd1250acf04a4a61edbe23667e0ea\",\"0x044289f44c7a8f4e08e44a32bbe494bed1daa1fce79008921acbed97bf14d265\",\"0xb113cdabab3d7dbe3dc61cc7f76e6effc9da4f056225520fd72cc81229bee02c\",\"0x71c20f978f598a01425d24cf74a907de329e0762e690656eeded1c0dcd099520\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000\",\"0x85dc2fe854246326b7af19488cd939150967b222127f0b39397f69c2b7e666e7\"]}"), &proofSiblings)
	var eventProof EventWithProof
	json.Unmarshal([]byte("{\"event\":\"0x001805000000000000003809644a7409cca52138ce747c56eaf20100000000000000073809644a7409cca52138ce747c56eaf21143726f7373436861696e4d616e616765720f43726f7373436861696e4576656e7400de02103809644a7409cca52138ce747c56eaf2100000000000000000000000000000000135307833383039363434613734303963636135323133386365373437633536656166323a3a43726f7373436861696e4d616e61676572da0000000000000034307833383039363434613734303963636135323133386365373437633536656166323a3a43726f7373436861696e536372697074c701100000000000000000000000000000000120f9b035fd78c030ab69af197a3e00a55f2eb4be1541750210b1eda92f3b008594103809644a7409cca52138ce747c56eaf2da0000000000000034307833383039363434613734303963636135323133386365373437633536656166323a3a43726f7373436861696e53637269707406756e6c6f636b3f0d3078313a3a5354433a3a53544310e498d62f5d1f469d2f72eb3e9dc8f230c7353a4200000000000000000000000000000000000000000000000000000000\",\"proof\":{\"siblings\":[\"0x8c601f513bf5b84fd18d66fc2983bd2b7e518492027d60cc7b73f548d5eb4661\",\"0xd04feed16570f47af77ad5e626e4de335fc170142c2424ec3840c20eef6a0a55\"]}}"), &eventProof)
	eventIdx := 1
	accessPath := ""
	proof := TransactionInfoProof{
		TransactionInfo: transactionInfo,
		Proof:           proofSiblings,
		EventIndex:      &eventIdx,
		EventWithProof:  eventProof,
		//StateWithProof:  "{\"state\":[32,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,24,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,24,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,24,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0],\"proof\":{\"account_state\":{\"blob\":[2,1,32,58,226,188,223,103,199,252,196,31,4,212,81,21,143,93,79,149,162,70,123,31,238,107,125,247,44,159,65,152,129,133,161,1,32,164,13,61,251,134,239,253,169,107,217,46,216,205,61,35,206,212,129,117,216,182,217,11,11,123,133,198,187,215,146,14,127]},\"account_proof\":{\"leaf\":[\"0xf1232a8cea9ae1d59147fcde30471108ec89c04579567e7aababc351b77ef1b6\",\"0xdfeab852f4f71945e94a5409c8283fcb32addef08ddd3d0ae9814de93a83efb3\"],\"siblings\":[\"0x360c7b79a7d2cc064e3af7befe431ac36613f62419afe2319a571ce2719f6202\",\"0x5350415253455f4d45524b4c455f504c414345484f4c4445525f484153480000\",\"0x5350415253455f4d45524b4c455f504c414345484f4c4445525f484153480000\"]},\"account_state_proof\":{\"leaf\":[\"0x9b079b5aef808c36133c95bacbc3d3411d1e8cce4fbc4b524ffab31202eaaa11\",\"0x91354856a1026edb028bc4aa72877a508483f3c9daef045b1d499cbf227fd0f7\"],\"siblings\":[\"0x318db4c3c46dc7b422e5dbda750952df1aa70387ab1331a8e0e39492a7fab69f\",\"0x5350415253455f4d45524b4c455f504c414345484f4c4445525f484153480000\",\"0x038cd2d75dbad3944f7476cb8e0c808e22a538e058ec0dd6a4508ba98d3a82b4\",\"0x77b69f571a7672871e87967a0d714613efa88ce20f28fa65b9f2c4b4748e2ee1\",\"0x8349b139ef1233c993be3376f07a5a9fa0b5973e22867c3142636d7d5d908e22\",\"0xcb174d26a2565684d3673ab2a01a3bf6273f38bf8d8c329beb85a9c3729449e1\"]}}}",
		StateWithProof: StateWithProofJson{},
		//accessPath:      "000000000000000000000000000000010100000000000000000000000000000001074163636f756e74074163636f756e7400",
		AccessPath: &accessPath,
	}
	fmt.Println("---------------------------- proof -----------------------------")
	p, err := json.Marshal(proof)
	if err != nil {
		t.FailNow()
	}
	fmt.Println(string(p))
	fmt.Println("---------------------------- end of proof -----------------------------")

	//eventData := []byte{0, 0, 79, 1, 81, 224, 51, 44, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 83, 84, 67, 3, 83, 84, 67}
	eventData := []byte{16, 56, 9, 100, 74, 116, 9, 204, 165, 33, 56, 206, 116, 124, 86, 234, 242, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 53, 48, 120, 51, 56, 48, 57, 54, 52, 52, 97, 55, 52, 48, 57, 99, 99, 97, 53, 50, 49, 51, 56, 99, 101, 55, 52, 55, 99, 53, 54, 101, 97, 102, 50, 58, 58, 67, 114, 111, 115, 115, 67, 104, 97, 105, 110, 77, 97, 110, 97, 103, 101, 114, 218, 0, 0, 0, 0, 0, 0, 0, 52, 48, 120, 51, 56, 48, 57, 54, 52, 52, 97, 55, 52, 48, 57, 99, 99, 97, 53, 50, 49, 51, 56, 99, 101, 55, 52, 55, 99, 53, 54, 101, 97, 102, 50, 58, 58, 67, 114, 111, 115, 115, 67, 104, 97, 105, 110, 83, 99, 114, 105, 112, 116, 199, 1, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 32, 249, 176, 53, 253, 120, 192, 48, 171, 105, 175, 25, 122, 62, 0, 165, 95, 46, 180, 190, 21, 65, 117, 2, 16, 177, 237, 169, 47, 59, 0, 133, 148, 16, 56, 9, 100, 74, 116, 9, 204, 165, 33, 56, 206, 116, 124, 86, 234, 242, 218, 0, 0, 0, 0, 0, 0, 0, 52, 48, 120, 51, 56, 48, 57, 54, 52, 52, 97, 55, 52, 48, 57, 99, 99, 97, 53, 50, 49, 51, 56, 99, 101, 55, 52, 55, 99, 53, 54, 101, 97, 102, 50, 58, 58, 67, 114, 111, 115, 115, 67, 104, 97, 105, 110, 83, 99, 114, 105, 112, 116, 6, 117, 110, 108, 111, 99, 107, 63, 13, 48, 120, 49, 58, 58, 83, 84, 67, 58, 58, 83, 84, 67, 16, 228, 152, 214, 47, 93, 31, 70, 157, 47, 114, 235, 62, 157, 200, 242, 48, 199, 53, 58, 66, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	txn_accumulator_root_bytes, _ := hex.DecodeString("b44a27b6f98fa9b04471e83bd40675381712a451299518cebab7f4ba9f137bd4")
	txnAccumulatorRoot := types.HashValue(txn_accumulator_root_bytes)

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"test event proof",
			args{
				proof:   &proof,
				data:    txnAccumulatorRoot,
				address: []byte("0x3809644a7409cca52138ce747c56eaf2::CrossChainManager::CrossChainEvent"),
			},
			eventData,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeEventV0, err := VerifyEventProof(tt.args.proof, tt.args.data, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyEventProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := typeEventV0.EventData
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

	var e Event
	ebyte, _ := hex.DecodeString("0x00180500000000000000e498d62f5d1f469d2f72eb3e9dc8f230020000000000000007e498d62f5d1f469d2f72eb3e9dc8f2301143726f7373436861696e4d616e616765720f43726f7373436861696e4576656e7400de0210e498d62f5d1f469d2f72eb3e9dc8f230100000000000000000000000000000000235307865343938643632663564316634363964326637326562336539646338663233303a3a43726f7373436861696e4d616e61676572da0000000000000034307865343938643632663564316634363964326637326562336539646338663233303a3a43726f7373436861696e536372697074c7011000000000000000000000000000000002208f5e5f785723333b2ab129a7928d1d47129a4df840da708d25086f1a361e0f6910e498d62f5d1f469d2f72eb3e9dc8f230da0000000000000034307865343938643632663564316634363964326637326562336539646338663233303a3a43726f7373436861696e53637269707406756e6c6f636b3f0d3078313a3a5354433a3a535443102d81a0427d64ff61b11ede9085efa5ad1027000000000000000000000000000000000000000000000000000000000000")
	err = json.Unmarshal([]byte(ebyte), &e)
	if err != nil {
		println(err)
	}
	println(e.SequenceNumber)
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

func TestUnmarshalTransactionInfoProof(t *testing.T) {
	//this is a json from on-chain RPC
	j := `
	{"transaction_info":{"block_hash":"0x429e6ad8617937da569474a3eca407ec54dd458b4d854f6bd577d152d1428478","block_number":"292517","transaction_hash":"0xd328475d67b5e3b7fef2fc4c3fb694a8904ddf90d91de80796280f5187cc66f0","transaction_index":1,"transaction_global_index":"324294","state_root_hash":"0x4df4cb2116088fc2da8cfbc2c47175b54ef4a59c3784ddc5982eff4afe6f706c","event_root_hash":"0x081e3db4b09ac60312ec8deb66455b9cfc7aa4201582cfd896b0d3b24ab2a04d","gas_used":"621382","status":"Executed"},"proof":{"siblings":["0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x5e9ef122f7f51479a3383f13301ad065ba1da4a2746ccec29527e81cf909e197","0x795a33fb6327e2a07249975fb2b712fd242545ed1ff4aec6442aaabc9f14b844","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x295fe28e3363cc6203b3537a17e3c9c4d1fe842a3ccedde809f68a9b29a099a4","0xf475b977f927ca0c487f9afaeac3e37cd9998c61cc686585e68d61f68cbf5202","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x30fe3ea6eb484e75084640e6cde700dd26550219eabd8447cc7e70c4285c04ea","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0xf6afb2287bd40fbc8057c567e04c9cfc48afd1250acf04a4a61edbe23667e0ea","0x044289f44c7a8f4e08e44a32bbe494bed1daa1fce79008921acbed97bf14d265","0xb113cdabab3d7dbe3dc61cc7f76e6effc9da4f056225520fd72cc81229bee02c","0x71c20f978f598a01425d24cf74a907de329e0762e690656eeded1c0dcd099520","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x414343554d554c41544f525f504c414345484f4c4445525f4841534800000000","0x85dc2fe854246326b7af19488cd939150967b222127f0b39397f69c2b7e666e7"]},"event_proof":{"event":"0x001805000000000000003809644a7409cca52138ce747c56eaf20100000000000000073809644a7409cca52138ce747c56eaf21143726f7373436861696e4d616e616765720f43726f7373436861696e4576656e7400de02103809644a7409cca52138ce747c56eaf2100000000000000000000000000000000135307833383039363434613734303963636135323133386365373437633536656166323a3a43726f7373436861696e4d616e61676572da0000000000000034307833383039363434613734303963636135323133386365373437633536656166323a3a43726f7373436861696e536372697074c701100000000000000000000000000000000120f9b035fd78c030ab69af197a3e00a55f2eb4be1541750210b1eda92f3b008594103809644a7409cca52138ce747c56eaf2da0000000000000034307833383039363434613734303963636135323133386365373437633536656166323a3a43726f7373436861696e53637269707406756e6c6f636b3f0d3078313a3a5354433a3a53544310e498d62f5d1f469d2f72eb3e9dc8f230c7353a4200000000000000000000000000000000000000000000000000000000","proof":{"siblings":["0x8c601f513bf5b84fd18d66fc2983bd2b7e518492027d60cc7b73f548d5eb4661","0xd04feed16570f47af77ad5e626e4de335fc170142c2424ec3840c20eef6a0a55"]}},"state_proof":null}
	`
	// unmarshal to Txn. info proof object
	proof := &TransactionInfoProof{}
	err := json.Unmarshal([]byte(j), proof)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	// set event index in the proof object
	eventIndex := 1
	proof.EventIndex = &eventIndex
	// verify event proof
	txn_accumulator_root, _ := hex.DecodeString("b44a27b6f98fa9b04471e83bd40675381712a451299518cebab7f4ba9f137bd4")
	typeEventV0, err := VerifyEventProof(proof, types.HashValue(txn_accumulator_root), []byte("0x3809644a7409cca52138ce747c56eaf2::CrossChainManager::CrossChainEvent"))
	eventData := typeEventV0.EventData
	eventKey := typeEventV0.Key
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	fmt.Println(eventData)
	fmt.Println(eventKey)
	fmt.Println(typeEventV0.TypeTag)
}
