package vanilla_chain

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"reflect"
	"testing"
	"time"
)

func TestSendTransactionSuccess(t *testing.T) {
	numOfPeers := 5
	numOfValidators := 3
	initialBalance := uint64(100000)
	peers := make([]Blockchain, numOfPeers)

	genesis := Genesis{
		make(map[string]uint64),
		make([]crypto.PublicKey, 0, numOfValidators),
	}

	keys := make([]ed25519.PrivateKey, numOfPeers)
	for i := range keys {
		_, key, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatal(err)
		}
		keys[i] = key
		if numOfValidators > 0 {
			genesis.Validators = append(genesis.Validators, key.Public())
			numOfValidators--
		}

		address, err := PubKeyToAddress(key.Public())
		if err != nil {
			t.Error(err)
		}
		genesis.Alloc[address] = initialBalance
	}

	var err error
	for i := 0; i < numOfPeers; i++ {
		peers[i], err = NewNode(keys[i], genesis, Validator)
		if err != nil {
			t.Error(err)
		}
		peers[i].Initialize()
	}

	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			err = peers[i].AddPeer(peers[j])
			if err != nil {
				t.Error(err)
			}
		}
		if peers[i].NodeGetType() == Validator {

		}
	}

	tr := Transaction{
		From:   peers[3].NodeAddress(),
		To:     peers[4].NodeAddress(),
		Amount: 100,
		Fee:    10,
		PubKey: keys[3].Public().(ed25519.PublicKey),
	}

	tr, err = peers[3].SignTransaction(tr)
	if err != nil {
		t.Fatal(err)
	}
	// test
	err = tr.Verify(tr.PubKey)

	err = peers[3].AddTransaction(tr)
	if err != nil {
		t.Fatal(err)
	}

	//wait transaction processing
	time.Sleep(time.Second * 5)

	//check "from" balance
	balance, err := peers[0].GetBalance(peers[3].NodeAddress())
	if err != nil {
		t.Fatal(err)
	}

	if balance != initialBalance-100-10 {
		t.Fatal("Incorrect from balance")
	}

	//check "to" balance
	balance, err = peers[0].GetBalance(peers[4].NodeAddress())
	if err != nil {
		t.Fatal(err)
	}

	if balance != initialBalance+100 {
		t.Fatal("Incorrect to balance")
	}

	//check validators balance
	for i := 0; i < 3; i++ {
		balance, err = peers[0].GetBalance(peers[i].NodeAddress())
		if err != nil {
			t.Error(err)
		}

		if balance > initialBalance {
			t.Error("Incorrect validator balance")
		}
	}
}

func TestHased(t *testing.T) {
	block := Block{
		BlockNum:      0,
		Timestamp:     11,
		Transactions:  nil,
		BlockHash:     "",
		PrevBlockHash: "00000",
		StateHash:     "",
		Signature:     nil,
	}

	t.Log(block.Hash())
}

func TestNode_Sync(t *testing.T) {
	numOfPeers := 5
	peers := make([]*Node, numOfPeers)

	genesis := Genesis{
		Alloc:      make(map[string]uint64),
		Validators: make([]crypto.PublicKey, 0, numOfPeers),
	}
	keys := make([]ed25519.PrivateKey, numOfPeers)
	for i := range keys {
		_, key, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatal(err)
		}
		keys[i] = key

		address, err := PubKeyToAddress(key.Public())
		if err != nil {
			t.Error(err)
		}
		genesis.Alloc[address] = 100
		genesis.Validators = append(genesis.Validators, key.Public())
	}
	var err error
	for i := 0; i < numOfPeers; i++ {
		peers[i], err = NewNode(keys[i], genesis, Validator)
		if err != nil {
			t.Error(err)
		}
		peers[i].Initialize()
	}
	block := NewBlock(1, nil, peers[2].GetBlockByNumber(0).BlockHash)
	err = block.Sign(peers[0].key)
	if err != nil {
		t.Fatal(err)
	}
	peers[0].addBlock(*block)

	if err != nil {
		t.Error(err)
	}
	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			err := peers[i].AddPeer(peers[j])
			if err != nil {
				t.Error(err)
			}
		}
	}

	time.Sleep(time.Second * 5)
	if !checkEqualsBlocks(peers) {
		t.Fatal("blocks not equal")
	}
}

func checkEqualsBlocks(peers []*Node) bool {
	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			if !reflect.DeepEqual(peers[i].blocks, peers[j].blocks) {
				return false
			}
		}
	}
	return true
}

func TestMinig(t *testing.T) {
	genesis := Genesis{
		Alloc: make(map[string]uint64),
	}
	pKey, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	genesis.Validators = append(genesis.Validators, pKey)

	peer, err := NewNode(key, genesis, Validator)
	peer.state.state.Store(peer.address, uint64(10_000))
	//peer.state.state[peer.address] = 10_000
	peer.validators = append(peer.validators, pKey)

	transaction := NewTransaction(peer.address, peer.address, 1000, 10, pKey, nil)
	*transaction, err = peer.SignTransaction(*transaction)

	hash, err := transaction.Hash()
	if err != nil {
		t.Fatal(err)
	}
	peer.blocks = append(peer.blocks, genesis.ToBlock())
	peer.transactionPool[hash] = *transaction
	ctx := context.Background()

	lastLenBefore := len(peer.blocks)
	lastBlockBefore := peer.lastBlockNum
	go peer.miningLoop(ctx)

	time.Sleep(time.Millisecond * 50)
	if peer.lastBlockNum == lastBlockBefore && len(peer.blocks) == lastLenBefore {
		t.Fatal("blocks not update")
	}
}
