package vanilla_chain

import (
	"context"
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
	return nil
}
func (c *Node) blockByNumResp(m BlockByNumResp, address string, ctx context.Context) error {
	if !reflect.DeepEqual(m.Block, Block{}) {
		err := c.AddBlock(m.Block, address)
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
	return nil
}

func (c *Node) addBlockResp(m AddBlockResp, address string, ctx context.Context) error {
	err := c.AddBlock(m.Block, m.NodeName)
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
