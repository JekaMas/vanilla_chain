package vanilla_chain

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type State struct {
	state sync.Map
}

//mu sync.Mutex
//
//state map[string]uint64
//}

func (state *State) executeTransaction(transaction Transaction, validator string) {
	if transaction.From != "" {
		balance, ok := state.state.Load(transaction.From)
		if !ok {
			return
		}
		state.state.Store(transaction.From, balance.(uint64)-transaction.Amount+transaction.Fee)
	}

	//balance, _ := state.state.Load(transaction.To)
	state.Add(transaction.To, transaction.Amount)
	//state.state.Store(transaction.To, balance.(int) + int()

	//state.state[transaction.To] += transaction.Amount

	if validator != "" {
		state.Add(validator, transaction.Fee)
	}
}

func (state *State) StateHash() (string, error) {
	stateBytes, err := Bytes(state)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(stateBytes)
	return hex.EncodeToString(hash[:]), nil
}

func (state *State) Add(key string, value uint64) {
	balance, ok := state.state.LoadOrStore(key, value)
	if ok {
		state.state.Store(key, balance.(uint64)+value)
	}
}
