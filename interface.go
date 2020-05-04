package vanilla_chain

import "crypto"

type Blockchain interface {
	Initialize()
	NodeKey() crypto.PublicKey
	NodeGetType() NodeType
	NodeAddress() string
	Connection(address string, in chan Message) chan Message

	PublicAPI
	AddBlock(block Block) error
}

type PublicAPI interface {
	//network
	AddPeer(Blockchain) error
	RemovePeer(Blockchain) error

	//for clients
	GetBalance(account string) (uint64, error)
	//add to transaction pool
	AddTransaction(transaction Transaction) error
	SignTransaction(transaction Transaction) (Transaction, error)

	//sync
	GetBlockByNumber(ID uint64) Block
	NodeInfo() NodeInfoResp
}
