package vanilla_chain

type connectedPeer struct {
	lastBlockNum uint64
}

func (peer *connectedPeer) SendBlock() {

}

func (peer *connectedPeer) GetBlockById(idBlock uint64) *Block {
	return &Block{}
}
