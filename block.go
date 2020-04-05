package vanilla_chain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Block struct {
	BlockNum      uint64
	Timestamp     int64
	Transactions  []Transaction
	BlockHash     string
	PrevBlockHash string
	StateHash     string
	Signature     []byte
}

func NewBlock(num uint64, transactions []Transaction, previousHash string) *Block {
	return &Block{
		BlockNum:      num,
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: previousHash,
	}
}

//impliment me hash(BlockNum, Timestamp, Transactions, PrevBlockHash)
func (b *Block) Hash() string {
	hash := sha256.New()

	hash.Write([]byte(string(b.BlockNum) +
		string(b.Timestamp) +
		/* + transactions   +*/
		b.PrevBlockHash))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}
