package vanilla_chain

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"
	"reflect"
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
		fmt.Println(c.lastBlockNum, " and ", m.BlockNum)
		if c.lastBlockNum < m.BlockNum {
			c.peers[address].Send(ctx, Message{
				From: c.address,
				Data: BlockByNumResp{
					NodeName: c.address,
					BlockNum: m.BlockNum - c.lastBlockNum,
				},
			})
		}
	case BlockByNumResp:
		if !reflect.DeepEqual(m.Block, Block{}) {
			err := c.AddBlock(m.Block)
			if err != nil {
				return err
			}
			if c.lastBlockNum < m.BlockNum {
				c.peers[address].Send(ctx, Message{
					From: c.address,
					Data: BlockByNumResp{
						NodeName: c.address,
						BlockNum: m.BlockNum - c.lastBlockNum,
					},
				})
			}
		} else {
			c.peers[address].Send(ctx, Message{
				From: c.address,
				Data: BlockByNumResp{
					NodeName: c.address,
					BlockNum: m.BlockNum,
					Block:    c.GetBlockByNumber(m.BlockNum),
				},
			})
		}
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

func (c *Node) GetBlockByNumber(id uint64) Block {
	if id > c.lastBlockNum || len(c.blocks) == 0 {
		return Block{}
	}
	return c.blocks[id]
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
	fmt.Println("Node "+c.address+" sync with ", peer.Address)
	for i := blockNum - c.lastBlockNum; i <= blockNum; i++ {
		peer.Send(ctx, Message{
			From: c.address,
			Data: BlockByNumResp{
				NodeName: c.address,
				BlockNum: i,
			},
		})
	NextBlock:
		for {
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
					break NextBlock
				default:
				}
			}
		}
	}
	return nil
}

func (c *Node) AddBlock(block Block) error {
	b := c.GetBlockByNumber(block.BlockNum)
	if !reflect.DeepEqual(b, Block{}) {
		return ErrBlockAlreadyExist
	}

	//todo check
	c.blocks = append(c.blocks, block)
	c.lastBlockNum = uint64(len(c.blocks)) - 1
	return nil
}
