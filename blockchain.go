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
	state      State // map[string]uint64 // балансы
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

		for _, transaction := range c.transactionPool {
			if len(transactions) == MaxTransactionBlock {
				break
			}

			err := c.checkTransaction(transaction, transaction.PubKey)
			if err == nil {
				transactions = append(transactions, transaction)
			}
		}

		block := NewBlock(uint64(c.lastBlockNum+1), transactions, c.blocks[c.lastBlockNum].PrevBlockHash)
		var err error
		block.StateHash, err = c.state.StateHash()
		if err != nil {
			continue
		}

		block.BlockHash, err = block.Hash()
		if err != nil {
			continue
		}
		err = block.Sign(c.key)
		if err != nil {
			continue
		}
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
	var err error
	switch m := msg.Data.(type) {
	case NodeInfoResp:
		err = c.nodeInfoResp(m, address, ctx)
	case BlockByNumResp:
		err = c.blockByNumResp(m, address, ctx)
		if err != nil {
			return err
		}
	case AddBlockResp:
		err = c.addBlockResp(m, address, ctx)
	case AddTransactionResp:
		err = c.addTransctionResp(m, address, ctx)
	}
	if err != nil {
		return err
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
	hash, err := transaction.Hash()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(c.transactionPool[hash], Transaction{}) {
		if reflect.DeepEqual(c.transactionPool[hash], transaction) {
			return nil
		}
		return ErrTransNotEqual
	}

	err = c.checkTransaction(transaction, transaction.PubKey)
	if err != nil {
		return err
	}

	c.transactionPool[hash] = transaction
	ctx := context.Background()

	c.Broadcast(ctx, Message{
		From: c.address,
		Data: AddTransactionResp{
			NodeName:    c.address,
			Transaction: Transaction{},
		},
	})
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

func (c *Node) AddBlock(block Block, validator string) error {
	c.blockMut.Lock()
	defer c.blockMut.Unlock()
	blockCheck := c.GetBlockByNumber(block.BlockNum)

	if !reflect.DeepEqual(blockCheck, Block{}) {
		if reflect.DeepEqual(block, blockCheck) {
			return ErrBlockAlreadyExist
		}
		return ErrBlocksNotEqual
	}

	err := block.Verify(c.validators[int(c.lastBlockNum)%len(c.validators)])
	if err != nil {
		return err
	}

	for _, transaction := range block.Transactions {
		err := c.checkTransaction(transaction, transaction.PubKey)
		if err != nil {
			//skip
			continue
		}
		//state.executeTransaction(transaction, block.)
		c.state.executeTransaction(transaction, validator)

	}

	c.blocks = append(c.blocks, block)
	c.lastBlockNum += 1

	return nil
}

func (c *Node) checkTransaction(transaction Transaction, key ed25519.PublicKey) error {
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
	if key == nil {
		return ErrNotHasPublicKey
	}

	if err := transaction.Verify(key); err != nil {
		return err
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
