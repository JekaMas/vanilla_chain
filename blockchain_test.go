package vanilla_chain

import (
	"crypto/ed25519"
	"testing"
)

type Blockchain interface {
	NodeKey() ed25519.PublicKey
	PublicAPI
}

type PublicAPI interface {
	//network
	AddPeer(Blockchain)
	RemovePeer(Blockchain)

	//for clients
	GetBalance(account string) (uint64, error)
	//add to transaction pool
	AddTransaction(transaction Transaction) error

	//sync
	GetBlockByNumber(ID uint64)
	NodeInfo() NodeInfoResp
}

type NodeInfoResp struct {
	NodeName string
	BlockNum uint64
}

func TestHased(t *testing.T) {
	block := Block{
		BlockNum:      0,
		Timestamp:     11,
		Transactions:  nil,
		BlockHash:     "",
		PrevBlockHash: "00000",
		StateHash:     "",
		Signature:     nil,
	}

	t.Log(block.Hash())
}

func TestName(t *testing.T) {
	var peers []Blockchain

	for i := 0; i < len(peers); i++ {
		for j := i; j < len(peers); j++ {
			peers[i].AddPeer(peers[j])
		}
	}

	ed25519.GenerateKey(nil)
}
