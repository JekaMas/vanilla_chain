package vanilla_chain

import (
	"crypto/ed25519"
	"errors"
	"time"
)

type Block struct {
	BlockNum      uint64
	Timestamp     int64
	Transactions  []Transaction
	BlockHash     string `json:"-"`
	PrevBlockHash string
	StateHash     string
	Signature     []byte `json:"-"`
}

func NewBlock(num uint64, transactions []Transaction, previousHash string) *Block {
	return &Block{
		BlockNum:      num,
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: previousHash,
	}
}

func (block *Block) Hash() (string, error) {
	if block == nil {
		return "", errors.New("empty block")
	}
	b, err := Bytes(block)
	if err != nil {
		return "", err
	}
	return Hash(b)
}

func (block *Block) Sign(key ed25519.PrivateKey) error {
	message, err := Bytes(block.BlockHash)
	if err != nil {
		return err
	}
	block.Signature = ed25519.Sign(key, message)
	return nil
}

func (block *Block) Verify(validator ed25519.PublicKey) error {
	message, err := Bytes(block.BlockHash)
	if err != nil {
		return err
	}
	verify := ed25519.Verify(validator, message, block.Signature)
	if !verify {
		return ErrVerifyNotPassed
	}
	return nil
}
