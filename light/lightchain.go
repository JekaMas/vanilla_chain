package light

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"errors"
	"log"
	"vanilla_chain"
)

type LightBlock struct {
	BlockHash     string `json:"-"`
	PrevBlockHash string
	StateHash     string
}

type LightNode struct {
	address string
	key     ed25519.PrivateKey
	peers   vanilla_chain.Peers
	blocks  []LightBlock
	state   vanilla_chain.State // map[string]uint64 // балансы

	balanceChan chan uint64
}

func (n *LightNode) Initialize() {
}

func (n *LightNode) NodeKey() crypto.PublicKey {
	return n.key.Public()
}

func (n *LightNode) NodeGetType() vanilla_chain.NodeType {
	return vanilla_chain.LightClient
}

func (n *LightNode) NodeAddress() string {
	return n.address
}

func (n *LightNode) Connection(address string, in chan vanilla_chain.Message, out chan vanilla_chain.Message) chan vanilla_chain.Message {
	if out == nil {
		out = make(chan vanilla_chain.Message, vanilla_chain.MSGBusLen)
	}
	ctx, cancel := context.WithCancel(context.Background())
	n.peers.Set(address, out, in, ctx, cancel)
	peer, ok := n.peers.Get(address)
	if !ok {
		return nil
	}
	go n.peerLoop(ctx, peer)
	return peer.Out
}

func (n *LightNode) AddPeer(peer vanilla_chain.Blockchain) error {
	remoteAddress, err := vanilla_chain.PubKeyToAddress(peer.NodeKey())
	if err != nil {
		return err
	}

	if n.address == remoteAddress {
		return errors.New("self connection")
	}

	if _, ok := n.peers.Get(remoteAddress); ok {
		return nil
	}

	out := make(chan vanilla_chain.Message, vanilla_chain.MSGBusLen)
	in := peer.Connection(n.address, out, nil)
	n.Connection(remoteAddress, in, out)
	return nil
}

func (n *LightNode) RemovePeer(peer vanilla_chain.Blockchain) error {
	err := n.peers.Delete(peer.NodeAddress())
	return err
}

func (n *LightNode) GetBalance(account string) (uint64, error) {
	balance, _ := n.state.LoadOrStore(account, uint64(0))
	return balance.(uint64), nil
}

func (n *LightNode) AddTransaction(transaction vanilla_chain.Transaction) error {
	n.peers.Broadcast(vanilla_chain.Message{
		From: n.address,
		Data: vanilla_chain.AddTransactionResp{
			NodeName:    n.address,
			Transaction: transaction,
		},
	})
	return nil
}

func (n *LightNode) SignTransaction(transaction vanilla_chain.Transaction) (vanilla_chain.Transaction, error) {
	err := transaction.Sign(n.key)
	if err != nil {
		return vanilla_chain.Transaction{}, err
	}
	return transaction, nil
}

func (n *LightNode) GetBlockByNumber(ID uint64) vanilla_chain.Block {

	return vanilla_chain.Block{}
}

func (n *LightNode) NodeInfo() vanilla_chain.NodeInfoResp {
	return vanilla_chain.NodeInfoResp{
		NodeName: n.NodeAddress(),
		BlockNum: 0,
	}
}

func (n *LightNode) AddBlock(block vanilla_chain.Block) error {
	return nil
}

func (n *LightNode) peerLoop(ctx context.Context, peer vanilla_chain.connectedPeer) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-peer.In:
			err := n.processMessage(ctx, peer.Address, msg)
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

func (n *LightNode) processMessage(ctx context.Context, address string, msg vanilla_chain.Message) error {
	var err error
	//switch m := msg.Data.(type) {
	//case NodeInfoResp:
	//	err = c.nodeInfoResp(m, address, ctx)
	//case BlockByNumResp:
	//	err = c.blockByNumResp(m, address, ctx)
	//case AddBlockResp:
	//	err = c.addBlockResp(m, address, ctx)
	//case AddTransactionResp:
	//	err = c.addTransctionResp(m, address, ctx)
	//}
	if err != nil {
		return err
	}
	return nil
}
