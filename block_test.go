package vanilla_chain

import (
	"crypto/ed25519"
	"testing"
)

func TestNode_AddBlock(t *testing.T) {

}

func TestBlock_Verify(t *testing.T) {
	validatorPublicKey, validatorPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	block := &Block{
		BlockNum:      1,
		Timestamp:     1000,
		Transactions:  []Transaction{},
		BlockHash:     "",
		PrevBlockHash: "0",
		StateHash:     "",
		Signature:     nil,
	}
	block.BlockHash, err = block.Hash()
	if err != err {
		t.Fatal(err)
	}
	err = block.Sign(validatorPrivateKey)
	if err != err {
		t.Fatal(err)
	}

	err = block.Verify(validatorPublicKey)
	if err != nil {
		t.Fatal()
	}
	validatorPublicKey2, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	err = block.Verify(validatorPublicKey2)
	if err != ErrVerifyNotPassed {
		t.Fatal()
	}
}
