package vanilla_chain

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"
)

const MSGBusLen = 100

type Node struct {
	key          ed25519.PrivateKey
	address      string
	genesis      Genesis
	lastBlockNum uint64

	//state
	blocks []Block
	//peer address - > peer info
	peers map[string]connectedPeer
	//hash(state) - хеш от упорядоченного слайса ключ-значение
	//todo hash()
	state      map[string]uint64
	validators []ed25519.PublicKey

	//transaction hash - > transaction
	transactionPool map[string]Transaction
}

func NewNode(key ed25519.PrivateKey, genesis Genesis) (*Node, error) {
	address, err := PubKeyToAddress(key.Public())
	if err != nil {
		return nil, err
	}
	return &Node{
		key:             key,
		address:         address,
		genesis:         genesis,
		blocks:          make([]Block, 0),
		lastBlockNum:    0,
		peers:           make(map[string]connectedPeer, 0),
		state:           make(map[string]uint64),
		transactionPool: make(map[string]Transaction),
	}, err
}

func (c *Node) NodeKey() crypto.PublicKey {
	return c.key.Public()
}

func (c *Node) Connection(address string, in chan Message) chan Message {
	out := make(chan Message, MSGBusLen)
	ctx, cancel := context.WithCancel(context.Background())
	c.peers[address] = connectedPeer{
		Address: address,
		Out:     out,
		In:      in,
		cancel:  cancel,
	}

	go c.peerLoop(ctx, c.peers[address])
	return c.peers[address].Out
}

func (c *Node) AddPeer(peer Blockchain) error {
	remoteAddress, err := PubKeyToAddress(peer.NodeKey())
	if err != nil {
		return err
	}

	if c.address == remoteAddress {
		return errors.New("self connection")
	}

	if _, ok := c.peers[remoteAddress]; ok {
		return nil
	}

	out := make(chan Message, MSGBusLen)
	in := peer.Connection(c.address, out)
	c.Connection(remoteAddress, in)
	return nil
}

func (c *Node) peerLoop(ctx context.Context, peer connectedPeer) {
	//todo handshake
	peer.Send(ctx, Message{
		From: c.address,
		Data: NodeInfoResp{
			NodeName: c.address,
			BlockNum: c.lastBlockNum,
		},
	})
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-peer.In:
			err := c.processMessage(peer.Address, msg)
			if err != nil {
				log.Println("Process peer error", err)
				continue
			}

			//broadcast to connected peers
			c.Broadcast(ctx, msg)
		}
	}
}

//func (c *Chain) Connection(){
//	for _, peer := range c.peers {
//		c.ConnectToPeer(peer)
//	}
//}
//func (c *Chain) ConnectToPeer(peer connectedPeer) {
//	if c.lastBlockNum < peer.lastBlockNum {
//		c.Sync(peer)
//	}
//}
//func (c *Chain) NextBlock(prevBlock Block) error {
//	if len(prevBlock.PrevBlockHash) == 0 {
//		return ErrPrevHashEmpty
//	}
//	block := NewBlock(prevBlock.BlockNum+1,
//		nil,
//		prevBlock.PrevBlockHash)
//	block.BlockHash = block.Hash()
//	c.addBlock(block)
//	return nil
//}

func (c *Node) processMessage(address string, msg Message) error {
	switch m := msg.Data.(type) {
	//example
	case NodeInfoResp:
		fmt.Println(c.address, "connected to ", address, "need sync", c.lastBlockNum < m.BlockNum)
	}
	return nil
}

func (c *Node) Broadcast(ctx context.Context, msg Message) {
	for _, v := range c.peers {
		if v.Address != c.address {
			v.Send(ctx, msg)
		}
	}
}

func (c *Node) RemovePeer(peer Blockchain) error {
	panic("implement me")
	return nil
}

func (c *Node) GetBalance(account string) (uint64, error) {
	panic("implement me")
}

func (c *Node) AddTransaction(transaction Transaction) error {
	panic("implement me")
}

func (c *Node) GetBlockByNumber(ID uint64) Block {
	panic("implement me")
}

func (c *Node) NodeInfo() NodeInfoResp {
	panic("implement me")
}

func (c *Node) NodeAddress() string {
	return c.address
}

func (c *Node) SignTransaction(transaction Transaction) (Transaction, error) {
	b, err := transaction.Bytes()
	if err != nil {
		return Transaction{}, err
	}

	transaction.Signature = ed25519.Sign(c.key, b)
	return transaction, nil
}

//func (c *Chain) ExecuteBlock(block *Block) error {
//	return nil
//}
//
//func (c *Chain) addBlock(block *Block) {
//	c.blocks = append(c.blocks, *block)
//}
//
//func (c *Chain) Consensus() {
//	for _, peer := range c.peers {
//		peer.SendBlock()
//	}
//}
//func (c *Chain) Sync(newPeer connectedPeer) {
//	for i := newPeer.lastBlockNum - c.lastBlockNum; i < c.lastBlockNum; i++ {
//		block := newPeer.GetBlockById(i)
//		err := c.ExecuteBlock(block)
//		if err == nil {
//			c.addBlock(block)
//		}
//	}
//}
