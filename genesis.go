package vanilla_chain

import (
	"crypto"
	"sort"
)

// first block with blockchain settings
type Genesis struct {
	Alloc      map[string]uint64  // Account -> funds
	Validators []crypto.PublicKey // list of validators public keys
}

// todo: не используешь, лучше сразу удаляй в отдельной ветке. будет нужно, то в ветке посмотришь.
//func NewGenesis(allocs map[string]uint64, validators []ed25519.PublicKey) *Genesis {
//	return &Genesis{
//		Alloc:      allocs,
//		Validators: validators,
//	}
//}

func (g Genesis) ToBlock() Block {
	transactions := make([]Transaction, 0, len(g.Alloc))

	var state State
	for address, amount := range g.Alloc {
		transactions = append(transactions,
			*NewTransaction("0", address, amount, 0, nil, nil))
		state.Set(address, amount)
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].To < transactions[j].To
	})

	prevHash, err := Hash([]byte("0")) // todo неясно, что это за хэш и почему он такой нужен. Может его выделить в константу?
	if err != nil {
		// todo тут и далее теряются ошибки
		return Block{}
	}

	block := NewBlock(0, transactions, prevHash)
	block.StateHash, err = state.StateHash()
	if err != nil {
		return Block{}
	}

	block.BlockHash, err = block.Hash()
	if err != nil {
		return Block{}
	}
	return *block
}
