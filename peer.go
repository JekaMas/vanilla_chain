package vanilla_chain

import (
	"context"
	"sync"
)

type Peers struct {
	peers map[string]ConnectedPeer
	sync.RWMutex
}

func (peers *Peers) Delete(key string) error {
	peer, ok := peers.Get(key)
	if !ok {
		return ErrPeerNotFound
	}
	peer.Cancel()
	delete(peers.peers, key)
	return nil
}
func (peers *Peers) Get(key string) (ConnectedPeer, bool) {
	peers.RLock()
	defer peers.RUnlock()
	peer, ok := peers.peers[key]
	return peer, ok
}
func (peers *Peers) Set(address string, out chan Message, in chan Message, ctx context.Context, cancel context.CancelFunc) {
	peers.Lock()
	defer peers.Unlock()

	peers.peers[address] = ConnectedPeer{
		Address: address,
		Out:     out,
		In:      in,
		ctx:     ctx,
		Cancel:  cancel,
	}
}

func (peers *Peers) Broadcast(msg Message) {
	peers.RLock()
	defer peers.RUnlock()
	for _, v := range peers.peers {
		if v.Address != msg.From {
			v.Send(msg)
		}
	}
}

type ConnectedPeer struct {
	Address string
	In      chan Message
	Out     chan Message
	ctx     context.Context
	Cancel  context.CancelFunc
}

func (cp ConnectedPeer) Send(m Message) {
	//todo timeout using context + done check
	cp.Out <- m
}

//func (peer *ConnectedPeer) GetBlockById(idBlock uint64) Block {
//	for {
//		select {
//		case block := <-peer.channels.blockChannel:
//			//err := json.Unmarshal(data, &block)
//			//if err != nil{
//			//	return Block{}
//			//}
//			if block.BlockNum == idBlock {
//				return block
//			}
//		case <-time.After(time.Minute):
//			return Block{}
//
//		}
//
//	}
//}
