package vanilla_chain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/gob"
)

type Transaction struct {
	From   string
	To     string
	Amount uint64
	Fee    uint64
	PubKey ed25519.PublicKey

	Signature []byte `json:"-"`
}

func NewTransaction(from, to string, amount, fee uint64, key ed25519.PublicKey, signature []byte) *Transaction {
	return &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		PubKey:    key,
		Signature: signature,
	}
}
func (transaction Transaction) Hash() (string, error) {
	b, err := Bytes(transaction)
	if err != nil {
		return "", err
	}
	return Hash(b)
}

func (transaction Transaction) Bytes() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).Encode(transaction)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (transaction *Transaction) Sign(key ed25519.PrivateKey) error {
	hash, err := transaction.Hash()
	if err != nil {
		return err
	}
	message, err := Bytes(hash)
	if err != nil {
		return err
	}
	transaction.Signature = ed25519.Sign(key, message)
	return nil
}

func (transaction *Transaction) Verify(validator ed25519.PublicKey) error {
	hash, err := transaction.Hash()
	if err != nil {
		return err
	}
	message, err := Bytes(hash)
	if err != nil {
		return err
	}
	verify := ed25519.Verify(validator, message, transaction.Signature)
	if !verify {
		return ErrVerifyNotPassed
	}
	return nil
}
