package vanilla_chain

import (
	"context"
)

type connectedPeer struct {
	Address string
	In      chan Message
	Out     chan Message
	cancel  context.CancelFunc
}

func (cp connectedPeer) Send(ctx context.Context, m Message) {
	//todo timeout using context + done check
	cp.Out <- m
}

//func (peer *connectedPeer) GetBlockById(idBlock uint64) Block {
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
