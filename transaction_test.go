package vanilla_chain

import (
	"crypto/ed25519"
	"errors"
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
				Amount:    uint64(500),
				Fee:       uint64(10),
				PubKey:    validatorPublicKey,
				signature: nil,
			},
			res: nil,
		},
		{
			name: "NotHasTo",
			trans: Transaction{
				From:      "a",
				To:        "",
				Amount:    uint64(500),
				Fee:       uint64(10),
				PubKey:    validatorPublicKey,
				signature: nil,
			},
			res: ErrTransToEmpty,
		},
		{
			name: "NotHasFrom",
			trans: Transaction{
				From:      "",
				To:        "b",
				Amount:    uint64(500),
				Fee:       uint64(10),
				PubKey:    validatorPublicKey,
				signature: nil,
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
				signature: nil,
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
				signature: []byte{'a'},
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
				signature: nil,
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
				signature: nil,
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
				signature: nil,
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
				signature: nil,
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
			node.state.Store("a", uint64(1000))
			node.state.Store("b", uint64(500))
			node.state.Store("c", uint64(10000))
			node.state.Store("d", uint64(5000))
			if test.trans.signature == nil {
				err := test.trans.Sign(validatorPrivateKey)
				if err != nil {
					if err != test.res {
						t.Fatal()
					}
					return
				}
			} else {
				test.trans.signature = nil
			}

			err = node.checkTransaction(test.trans)
			if !errors.Is(err, test.res) {
				t.Fatalf("an unexpected error.\nGot\t%v\nwanted\t%v\n", err, test.res)
			}
		})
	}
}
