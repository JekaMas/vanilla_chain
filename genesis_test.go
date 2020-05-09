package vanilla_chain

import (
	"reflect"
	"testing"
)

// todo добавь тестов: пустой Alloc, nil Alloc,
func TestGenesis_ToBlock(t *testing.T) {
	var tests = []struct {
		name    string
		genesis Genesis
		block   Block
	}{
		{
			name: "Simple",
			genesis: Genesis{
				Alloc: map[string]uint64{
					"c": 2,
					"a": 1,
					"b": 3,
				},
			},
			block: Block{Transactions: []Transaction{
				{
					From:   "0",
					To:     "a",
					Amount: 1,
				},
				{
					From:   "0",
					To:     "b",
					Amount: 3,
				},
				{
					From:   "0",
					To:     "c",
					Amount: 2,
				},
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := test.genesis.ToBlock()
			if !reflect.DeepEqual(block.Transactions, test.block.Transactions) {
				// todo если попадаем в ошибку, то как нам отлаживать? Попробуй тут же печатать, что ожидали, а что получили.
				t.Fatal("genesis not true")
			}
		})
	}
}
