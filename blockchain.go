package vanilla_chain

import ()

// first block with blockchain settings

type Chain struct {
	blocks       []Block
	lastBlockNum uint64
	//peer name - > peer info
	peers map[string]connectedPeer
	//hash(state) - хеш от упорядоченного слайса ключ-значение
	state map[string]uint64
	//transaction hash - > transaction
	transactionPool map[string]Transaction
	//genesis
	genesis Genesis
}

func NewChain(allcos map[string]uint64) *Chain {
	//genesis := NewGenesis(allcos, nil)

	chain := Chain{
		blocks:          []Block{},
		lastBlockNum:    0,
		peers:           nil,
		state:           nil,
		transactionPool: nil,
	}
	chain.addBlock(&Block{})
	return &chain
}

func (c *Chain) ConnectToPeer(peer connectedPeer) {
	if c.lastBlockNum < peer.lastBlockNum {
		c.Sync(peer)
	}
}
func (c *Chain) NextBlock(prevBlock Block) error {
	if len(prevBlock.PrevBlockHash) == 0 {
		return ErrPrevHashEmpty
	}
	block := NewBlock(prevBlock.BlockNum+1,
		nil,
		prevBlock.PrevBlockHash)
	block.BlockHash = block.Hash()
	c.addBlock(block)
	return nil
}

func (c *Chain) ExecuteBlock(block *Block) error {
	return nil
}

func (c *Chain) addBlock(block *Block) {
	c.blocks = append(c.blocks, *block)
}

func (c *Chain) Consensus() {
	for _, peer := range c.peers {
		peer.SendBlock()
	}
}
func (c *Chain) Sync(newPeer connectedPeer) {
	for i := newPeer.lastBlockNum - c.lastBlockNum; i < c.lastBlockNum; i++ {
		block := newPeer.GetBlockById(i)
		err := c.ExecuteBlock(block)
		if err == nil {
			c.addBlock(block)
		}
	}
}
