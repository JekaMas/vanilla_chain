package vanilla_chain

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

const MSGBusLen = 1_000_000
const MaxTransactionBlock = 10

const (
	User = iota
	Validator
	LightClient
)

type NodeType byte

type Node struct {
	send         int
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
	peerMut sync.Mutex
	peers   map[string]ConnectedPeer
	//hash(state) - хеш от упорядоченного слайса ключ-значение
	state State // map[string]uint64 // балансы

	validators []ed25519.PublicKey

	//transaction hash - > transaction
	transMut        sync.Mutex
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
		peers:           make(map[string]ConnectedPeer, 0),
		transactionPool: make(map[string]Transaction),
		nodeType:        nodeType,
	}, err
}

func (c *Node) Initialize() {
	block := c.genesis.ToBlock()
	for _, transaction := range block.Transactions {
		err := c.state.executeTransaction(transaction, "0")
		// todo: не стоит игнорировать ошибки
		if err != nil {
			panic(err)
		}
	}

	c.blocks = append(c.blocks, block)
	for _, validator := range c.genesis.Validators {
		c.validators = append(c.validators, validator.(ed25519.PublicKey))
	}

	if c.nodeType == Validator {
		// todo это лучше выделить в отдельный метод вроде StartMining
		go c.miningLoop(context.Background())
	}
}

func (c *Node) NodeKey() crypto.PublicKey {
	return c.key.Public()
}

func (c *Node) NodeGetType() NodeType {
	return c.nodeType
}

func (c *Node) Connection(address string, in chan Message, out chan Message) chan Message {
	if out == nil {
		out = make(chan Message, MSGBusLen)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.peerMut.Lock()
	c.peers[address] = ConnectedPeer{
		Address: address,
		Out:     out,
		In:      in,
		Cancel:  cancel,
	}
	c.peerMut.Unlock()
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
		if !c.checkValidatorTurn() {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		transactions := make([]Transaction, 0, MaxTransactionBlock)

		c.transMut.Lock()
		for _, transaction := range c.transactionPool {
			if len(transactions) == MaxTransactionBlock {
				break
			}

			err := c.checkTransaction(transaction)
			c.state.executeTransaction(transaction, c.address)
			if err != nil {
				continue
			}
			transactions = append(transactions, transaction)

			hash, err := transaction.Hash()
			if err != nil {
				continue
			}

			delete(c.transactionPool, hash)
		}
		c.transMut.Unlock()
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
		c.blockMut.Lock()
		c.addBlock(*block)
		c.blockMut.Unlock()

		c.state.Add(c.address, 1000)
		c.Broadcast(ctx, Message{
			From: c.address,
			Data: AddBlockResp{Block: *block},
		})
	}
}

func (c *Node) peerLoop(ctx context.Context, peer ConnectedPeer) {
	peer.Send(Message{
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
	c.peerMut.Lock()
	defer c.peerMut.Unlock()
	for _, v := range c.peers {
		if v.Address != c.address && v.Address != msg.From {
			v.Send(msg)
		}
	}
}

// todo: тут не бывает ошибки
func (c *Node) RemovePeer(peer Blockchain) error {
	c.peers[peer.NodeAddress()].Cancel()
	delete(c.peers, peer.NodeAddress())
	return nil
}

// todo: тут не бывает ошибки
func (c *Node) GetBalance(account string) (uint64, error) {
	balance, _ := c.state.LoadOrStore(account, uint64(0))
	return balance.(uint64), nil
}

func (c *Node) AddTransaction(transaction Transaction) error {
	c.transMut.Lock()
	defer c.transMut.Unlock()
	hash, err := transaction.Hash()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(c.transactionPool[hash], Transaction{}) {
		if reflect.DeepEqual(c.transactionPool[hash], transaction) {
			return ErrTransAlreadyExist
		}
		return ErrTransNotEqual
	}

	err = c.checkTransaction(transaction)
	if err != nil {
		return err
	}

	c.transactionPool[hash] = transaction
	ctx := context.Background()

	c.Broadcast(ctx, Message{
		From: c.address,
		Data: AddTransactionResp{
			NodeName:    c.address,
			Transaction: transaction,
		},
	})
	return nil
}

func (c *Node) GetBlockByNumber(id uint64) Block {
	//c.blockMut.Lock()
	//defer c.blockMut.Unlock()
	if id > c.lastBlockNum || len(c.blocks) < int(id+1) {
		return Block{}
	}
	return c.blocks[id]
}

func (c *Node) NodeInfo() NodeInfoResp {
	c.blockMut.Lock()
	defer c.blockMut.Unlock()
	return NodeInfoResp{
		NodeName: c.address,
		BlockNum: c.lastBlockNum,
	}
}

func (c *Node) NodeAddress() string {
	return c.address
}

func (c *Node) SignTransaction(transaction Transaction) (Transaction, error) {
	err := transaction.Sign(c.key)
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func (c *Node) AddBlock(block Block) error {
	c.blockMut.Lock()
	defer c.blockMut.Unlock()
	blockCheck := c.GetBlockByNumber(block.BlockNum)

	if !reflect.DeepEqual(blockCheck, Block{}) {
		if reflect.DeepEqual(block, blockCheck) {
			return ErrBlockAlreadyExist
		}
		return ErrBlocksNotEqual
	}
	validator := c.validators[int(c.lastBlockNum)%len(c.validators)]
	err := block.Verify(validator)
	if err != nil {
		return err
	}

	validatorAddress, err := PubKeyToAddress(validator)
	if err != nil {
		return nil
	}
	for _, transaction := range block.Transactions {
		err := c.checkTransaction(transaction)
		if err != nil {
			//skip
			continue
		}
		//state.executeTransaction(transaction, block.)

		c.state.executeTransaction(transaction, validatorAddress)
		hash, err := transaction.Hash()
		if err != nil {
			continue
		}
		c.transMut.Lock()
		delete(c.transactionPool, hash)
		c.transMut.Unlock()
	}
	c.state.Add(validatorAddress, 1000)
	c.addBlock(block)

	return nil
}

func (c *Node) addBlock(block Block) {
	c.blocks = append(c.blocks, block)
	c.lastBlockNum += 1
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
	if transaction.signature == nil {
		return ErrTransNotHasSignature
	}
	if transaction.PubKey == nil {
		return ErrNotHasPublicKey
	}
	if err := transaction.Verify(transaction.PubKey); err != nil {
		return err
	}
	balance, err := c.GetBalance(transaction.From)
	if err != nil {
		// todo: ты теряешь тут и в подобных местах изначальное сообщение об ошибке
		return fmt.Errorf("%w: %s", ErrTransNotHasNeedSum, err)
	}
	if balance < transaction.Amount+transaction.Fee {
		return fmt.Errorf("%w: balance %d. amount+fee %d", ErrTransNotHasNeedSum, balance, transaction.Amount+transaction.Fee)
	}

	return nil
}

func (c *Node) checkValidatorTurn() bool {
	c.blockMut.Lock()
	defer c.blockMut.Unlock()
	validatorAddr, err := PubKeyToAddress(c.validators[int(c.lastBlockNum)%len(c.validators)])
	if err != nil {
		// todo теряешь саму произошедшую ошибку. будет трудно отлаживать
		return false
	}
	if validatorAddr != c.address {
		return false
	}
	c.transMut.Lock()
	defer c.transMut.Unlock()
	if len(c.transactionPool) == 0 {
		return false
	}

	return true
}
