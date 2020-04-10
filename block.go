package vanilla_chain

import (
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

func (bl *Block) Hash() (string, error) {
	if bl == nil {
		return "", errors.New("empty block")
	}
	b, err := Bytes(bl)
	if err != nil {
		return "", err
	}
	return Hash(b)
}
