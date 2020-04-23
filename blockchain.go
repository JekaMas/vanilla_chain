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
const MaxTransactionBlock = 10

const (
	User = iota
	Miner
)

type NodeType byte

type Node struct {
	key          ed25519.PrivateKey
	address      string
	genesis      Genesis
	lastBlockNum uint64
	nodeType     NodeType
	//sync
	handshake bool
	//state
	blocks   []Block
	blockMut sync.Mutex
	//peer address - > peer info
	peers map[string]connectedPeer
	//hash(state) - хеш от упорядоченного слайса ключ-значение
	//todo hash()
	state      map[string]uint64 // балансы
	validators []ed25519.PublicKey

	//transaction hash - > transaction
	transactionPool map[string]Transaction
}

func NewNode(key ed25519.PrivateKey, genesis Genesis, nodeType NodeType) (*Node, error) {
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
		nodeType:        nodeType,
	}, err
}

func (c *Node) NodeKey() crypto.PublicKey {
	return c.key.Public()
}

func (c *Node) Connection(address string, in chan Message, out chan Message) chan Message {
	if out == nil {
		out = make(chan Message, MSGBusLen)
	}
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
	in := peer.Connection(c.address, out, nil)
	c.Connection(remoteAddress, in, out)
	return nil
}

func (c *Node) miningLoop(ctx context.Context) {
	for {
		if !c.checkValidatorTurn() || len(c.transactionPool) == 0 {
			time.Sleep(time.Millisecond * 25)
			continue
		}

		transactions := make([]Transaction, 0, MaxTransactionBlock)
		transactionsDel := make([]Transaction, 0, MaxTransactionBlock)
		for key, transaction := range c.transactionPool {
			if len(transactions) == MaxTransactionBlock {
				break
			}

			err := c.checkTransaction(transaction)
			if err == nil {
				transactions = append(transactions, transaction)
			}

			transactionsDel = append(transactionsDel, transaction)
			delete(c.transactionPool, key)
		}
		c.Broadcast(ctx, Message{
			From: c.address,
			Data: DelTransResp{
				NodeName:     c.address,
				Transactions: transactionsDel,
			},
		})

		if len(transactions) == 0 {
			continue
		}

		block := NewBlock(uint64(c.lastBlockNum+1), transactions, c.blocks[c.lastBlockNum].PrevBlockHash)
		c.blocks = append(c.blocks, *block)
		c.lastBlockNum++
		c.Broadcast(ctx, Message{
			From: c.address,
			Data: AddBlockResp{Block: *block},
		})
	}
}

func (c *Node) peerLoop(ctx context.Context, peer connectedPeer) {
	peer.Send(ctx, Message{
		From: c.address,
		Data: c.NodeInfo(),
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

			////broadcast to connected peers
			//if broadcasting == true {
			//	c.Broadcast(ctx, msg)
			//}
		}
	}
}

func (c *Node) processMessage(ctx context.Context, address string, msg Message) error {
	switch m := msg.Data.(type) {
	case NodeInfoResp:
		//fmt.Println(c.lastBlockNum, "(", c.address , ")", " and ", m.BlockNum)
		if c.lastBlockNum < m.BlockNum && !c.handshake {
			c.handshake = true
			c.peers[address].Send(ctx, Message{
				From: c.address,
				Data: BlockByNumResp{
					NodeName: c.address,
					BlockNum: c.lastBlockNum + 1,
				},
			})
		}
	case BlockByNumResp:
		if !reflect.DeepEqual(m.Block, Block{}) {
			err := c.AddBlock(m.Block)
			if err != nil {
				return err
			}
			if c.lastBlockNum < m.LastBlockNum {
				c.peers[address].Send(ctx, Message{
					From: c.address,
					Data: BlockByNumResp{
						NodeName: c.address,
						BlockNum: c.lastBlockNum + 1,
					},
				})
			} else {
				c.handshake = false
			}
		} else {
			c.peers[address].Send(ctx, Message{
				From: c.address,
				Data: BlockByNumResp{
					NodeName:     c.address,
					BlockNum:     m.BlockNum,
					LastBlockNum: c.lastBlockNum,
					Block:        c.GetBlockByNumber(m.BlockNum),
				},
			})
		}
	}
	return nil
}

func (c *Node) Broadcast(ctx context.Context, msg Message) {
	for _, v := range c.peers {
		if v.Address != c.address && v.Address != msg.From {
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
	err := c.checkTransaction(transaction)
	if err != nil {
		return err
	}

	hash, err := transaction.Hash()
	if err != nil {
		return err
	}
	c.transactionPool[hash] = transaction
	return nil
}

func (c *Node) GetBlockByNumber(id uint64) Block {
	if id > c.lastBlockNum || len(c.blocks) < int(id+1) {
		return Block{}
	}
	return c.blocks[id]
}

func (c *Node) NodeInfo() NodeInfoResp {
	return NodeInfoResp{
		NodeName: c.address,
		BlockNum: c.lastBlockNum,
	}
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

func (c *Node) AddBlock(block Block) error {
	c.blockMut.Lock()
	defer c.blockMut.Unlock()
	blockCheck := c.GetBlockByNumber(block.BlockNum)

	if !reflect.DeepEqual(blockCheck, Block{}) {
		return ErrBlockAlreadyExist
	}
	if len(c.blocks) == 2 {
		x := 2
		x = x
	}
	for _, transaction := range block.Transactions {
		err := c.checkTransaction(transaction)
		if err != nil {
			//skip
		}

	}

	c.blocks = append(c.blocks, block)
	c.lastBlockNum = uint64(len(c.blocks)) - 1

	return nil
}

func (c *Node) checkTransaction(transaction Transaction) error {
	if transaction.To == "" {
		return ErrTransToEmpty
	}
	if transaction.From == "" {
		return ErrTransFromEmpty
	}
	if transaction.Amount <= 0 {
		return ErrTransAmountNotValid
	}
	if transaction.Signature == nil {
		return ErrTransNotHasSignature
	}
	balance, ok := c.state[transaction.From]
	if !ok {
		return ErrTransNotHasNeedSum
	}
	if balance < transaction.Amount+transaction.Fee {
		return ErrTransNotHasNeedSum
	}

	return nil
}

func (c *Node) checkValidatorTurn() bool {
	validatorAddr, err := PubKeyToAddress(c.validators[int(c.lastBlockNum)%len(c.validators)])
	if err != nil {
		return false
	}
	return validatorAddr == c.address
}
