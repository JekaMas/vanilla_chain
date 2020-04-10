package vanilla_chain

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"errors"
	"log"
	"reflect"
	"time"
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
	state      map[string]uint64 // балансы
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
			err := c.processMessage(ctx, peer.Address, msg)
			if err != nil {
				log.Println("Process peer error", err)
				continue
			}

			//broadcast to connected peers
			c.Broadcast(ctx, msg)
		}
	}
}

func (c *Node) processMessage(ctx context.Context, address string, msg Message) error {
	switch m := msg.Data.(type) {
	//example
	case NodeInfoResp:
		if c.lastBlockNum <= m.BlockNum {
			c.Sync(c.peers[address], m.BlockNum)
		}
	case BlockByNumResp:
		c.peers[address].Send(ctx, Message{
			From: c.address,
			Data: BlockByNumResp{
				NodeName: c.address,
				BlockNum: m.BlockNum,
				Block:    c.GetBlockByNumber(m.BlockNum),
			},
		})
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
	if ID > c.lastBlockNum {
		return Block{}
	}
	return c.blocks[ID]
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
func (c *Node) Sync(ctx context.Context, peer connectedPeer, blockNum uint64) error {
	for i := blockNum - c.lastBlockNum + 1; i < blockNum; i++ {
		peer.Send(ctx, Message{
			From: c.address,
			Data: BlockByNumResp{
				NodeName: c.address,
				BlockNum: i,
			},
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-peer.In:
			switch m := msg.Data.(type) {
			case BlockByNumResp:
				err := c.AddBlock(m.Block)
				if err != nil {
					return err
				}
			}
		}

	}
}

func (c *Node) AddBlock(block Block) error {
	block := c.GetBlockByNumber(block.BlockNum)
	if !reflect.DeepEqual(block, Block{}) {
		return ErrBlockAlreadyExist
	}

}
