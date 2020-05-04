package vanilla_chain

import (
	"context"
	"fmt"
	"reflect"
)

type Message struct {
	From string
	Data interface{}
}

type NodeInfoResp struct {
	NodeName string
	BlockNum uint64
}

type BlockByNumResp struct {
	NodeName     string
	BlockNum     uint64
	LastBlockNum uint64
	Block        Block
}

type AddBlockResp struct {
	NodeName string
	Block    Block
}

type AddTransactionResp struct {
	NodeName    string
	Transaction Transaction
}
type DelTransResp struct {
	NodeName     string
	Transactions []Transaction
}

func (c *Node) nodeInfoResp(m NodeInfoResp, address string, ctx context.Context) error {
	c.blockMut.Lock()
	defer c.blockMut.Unlock()
	if c.lastBlockNum < m.BlockNum && !c.handshake {
		c.handshake = true
		c.peerMut.Lock()
		c.peers[address].Send(ctx, Message{
			From: c.address,
			Data: BlockByNumResp{
				NodeName: c.address,
				BlockNum: c.lastBlockNum + 1,
			},
		})
		c.peerMut.Unlock()
	}
	return nil
}
func (c *Node) blockByNumResp(m BlockByNumResp, address string, ctx context.Context) error {
	if !reflect.DeepEqual(m.Block, Block{}) {
		err := c.AddBlock(m.Block)
		if err != nil {
			return err
		}
		c.blockMut.Lock()
		defer c.blockMut.Unlock()
		if c.lastBlockNum < m.LastBlockNum {
			c.peerMut.Lock()
			c.peers[address].Send(ctx, Message{
				From: c.address,
				Data: BlockByNumResp{
					NodeName: c.address,
					BlockNum: c.lastBlockNum + 1,
				},
			})
			c.peerMut.Unlock()
		} else {
			c.handshake = false
		}
	} else {
		c.peerMut.Lock()
		c.peers[address].Send(ctx, Message{
			From: c.address,
			Data: BlockByNumResp{
				NodeName:     c.address,
				BlockNum:     m.BlockNum,
				LastBlockNum: c.lastBlockNum,
				Block:        c.GetBlockByNumber(m.BlockNum),
			},
		})
		c.peerMut.Unlock()
	}
	return nil
}

func (c *Node) addBlockResp(m AddBlockResp, address string, ctx context.Context) error {

	err := c.AddBlock(m.Block)
	fmt.Println(err)
	if err != nil {
		if err == ErrBlockAlreadyExist {
			return nil
		}
		return err
	}
	c.Broadcast(ctx, Message{
		From: c.address,
		Data: m,
	})

	return nil
}

func (c *Node) addTransctionResp(m AddTransactionResp, address string, ctx context.Context) error {
	err := c.AddTransaction(m.Transaction)
	if err != nil {
		if err == ErrTransAlreadyExist {
			return nil
		}
		return err
	}
	c.Broadcast(ctx, Message{
		From: c.address,
		Data: m,
	})

	return nil
}
