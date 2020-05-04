package vanilla_chain

import (
	"crypto"
	"sort"
)

// first block with blockchain settings
type Genesis struct {
	//Account -> funds
	Alloc map[string]uint64
	//list of validators public keys
	Validators []crypto.PublicKey
}

//func NewGenesis(allocs map[string]uint64, validators []ed25519.PublicKey) *Genesis {
//	return &Genesis{
//		Alloc:      allocs,
//		Validators: validators,
//	}
//}

func (g Genesis) ToBlock() Block {
	transactions := make([]Transaction, len(g.Alloc))

	var state State
	i := 0
	for address, amount := range g.Alloc {
		transactions[i] = *NewTransaction("", address, amount, 0, nil, nil)
		state.Add(address, amount)
		i++
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].To < transactions[j].To
	})
	prevHash, err := Hash([]byte("0"))
	if err != nil {
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
