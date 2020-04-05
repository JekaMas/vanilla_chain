package vanilla_chain

import "crypto/ed25519"

type Transaction struct {
	From   string
	To     string
	Amount uint64
	Fee    uint64
	PubKey ed25519.PublicKey

	Signature []byte
}

//impliment me
func (t *Transaction) Hash() string {
	return ""
}
