package vanilla_chain

import (
	"crypto/ed25519"
	"reflect"
	"testing"
)

func TestNode_CheckTransaction(t *testing.T) {
	validatorPublicKey, validatorPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	validatorPublicKey2, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		trans Transaction
		res   error
	}{
		{
			name: "OK",
			trans: Transaction{
				From:      "a",
				To:        "b",
				Amount:    500,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: nil,
			},
			res: nil,
		},
		{
			name: "NotHasTo",
			trans: Transaction{
				From:      "a",
				To:        "",
				Amount:    500,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: nil,
			},
			res: ErrTransToEmpty,
		},
		{
			name: "NotHasFrom",
			trans: Transaction{
				From:      "",
				To:        "b",
				Amount:    500,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: nil,
			},
			res: ErrTransFromEmpty,
		},
		{
			name: "NotHasAmount",
			trans: Transaction{
				From:      "a",
				To:        "b",
				Amount:    0,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: nil,
			},
			res: ErrTransAmountNotValid,
		},
		{
			name: "NotHasSignature",
			trans: Transaction{
				From:      "a",
				To:        "b",
				Amount:    500,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: []byte{'a'},
			},
			res: ErrTransNotHasSignature,
		},
		{
			name: "VerifyFail",
			trans: Transaction{
				From:      "a",
				To:        "b",
				Amount:    500,
				Fee:       10,
				PubKey:    validatorPublicKey2,
				Signature: nil,
			},
			res: ErrVerifyNotPassed,
		},
		{
			name: "NotSelectPubKey",
			trans: Transaction{
				From:      "a",
				To:        "b",
				Amount:    500,
				Fee:       10,
				PubKey:    nil,
				Signature: nil,
			},
			res: ErrNotHasPublicKey,
		},
		{
			name: "NotNotHasBeSum",
			trans: Transaction{
				From:      "ErrTransNotHasNeedSum",
				To:        "b",
				Amount:    500,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: nil,
			},
			res: ErrTransNotHasNeedSum,
		},
		{
			name: "NotNotHasBeSum2",
			trans: Transaction{
				From:      "a",
				To:        "b",
				Amount:    100_000,
				Fee:       10,
				PubKey:    validatorPublicKey,
				Signature: nil,
			},
			res: ErrTransNotHasNeedSum,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := &Node{
				//state: State{state:State{
				//	"a": 1000,
				//	"b": 500,
				//	"c": 10000,
				//	"d": 5000,
				//	},
				//},
			}
			node.state.state.Store("a", uint64(1000))
			node.state.state.Store("b", uint64(500))
			node.state.state.Store("c", uint64(10000))
			node.state.state.Store("d", uint64(5000))
			if test.trans.Signature == nil {
				err := test.trans.Sign(validatorPrivateKey)
				if err != nil {
					if err != test.res {
						t.Fatal()
					}
					return
				}
			} else {
				test.trans.Signature = nil
			}
			err = node.checkTransaction(test.trans)
			if !reflect.DeepEqual(err, test.res) {
				t.Fatal(err, " != ", test.res)
			}
		})
	}
}
