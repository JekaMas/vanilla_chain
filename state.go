package vanilla_chain

import (
	"crypto/sha256"
	"encoding/hex"
)

type State map[string]uint64

func (state State) executeTransaction(transaction Transaction, validator string) {
	state[transaction.From] -= transaction.Amount + transaction.Fee

	state[transaction.To] += transaction.Amount
	//if (validator == )
	state[validator] += transaction.Fee
}

func (state State) StateHash() (string, error) {
	stateBytes, err := Bytes(state)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(stateBytes)
	return hex.EncodeToString(hash[:]), nil
}
