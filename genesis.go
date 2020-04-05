package vanilla_chain

import "crypto/ed25519"

type Genesis struct {
	//Account -> funds
	Alloc map[string]uint64
	//list of validators public keys
	Validators []ed25519.PublicKey
}

func NewGenesis(allocs map[string]uint64, validators []ed25519.PublicKey) *Genesis {
	return &Genesis{
		Alloc:      allocs,
		Validators: validators,
	}
}
