package vanilla_chain

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

type DelTransResp struct {
	NodeName     string
	Transactions []Transaction
}
