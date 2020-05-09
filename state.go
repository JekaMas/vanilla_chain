package vanilla_chain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
)

type State struct {
	sync.Map
	sync.RWMutex // todo да, этот мьютекс убьет весь выигрыш от sync.Map
}

func (state *State) executeTransaction(transaction Transaction, validator string) error {
	// todo хотя каждое отдельное действие с sync.Map атомарно, но между этимы выховами с состоянием может случиться что угодно

	state.Lock()
	defer state.Unlock()

	// todo тут надо отменять изменения, если таковые были сделаны. или не применять никакие изменения, пока всё не проверим.
	txFrom, err := state.SubLazy(transaction.From, transaction.Amount + transaction.Fee)
	if err != nil {
		return err
	}

	txTo, err := state.AddLazy(transaction.To, transaction.Amount)
	if err != nil {
		return err
	}

	txValidator, err := state.AddLazy(validator, transaction.Fee)
	if err != nil {
		return err
	}

	txTo()
	txFrom()
	txValidator()

	return nil
}

func (state *State) StateHash() (string, error) {
	stateBytes, err := Bytes(state)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(stateBytes)
	return hex.EncodeToString(hash[:]), nil
}

// todo: лучше разделить методы Set и Add. У тебя метод Set вводил в заблуждение, потому что он мог делать не Set, а Add
func (state *State) Set(key string, value uint64) {
	state.Store(key, value)
}

func (state *State) Get(key string) (uint64, bool) {
	balance, ok := state.Load(key)
	if ok {
		v, ok := balance.(uint64)
		return v, ok
	}
	return 0, ok
}

func (state *State) Add(key string, value uint64) error {
	fn, err := state.AddLazy(key, value)
	if err != nil {
		return err
	}

	fn()
	return nil
}

func (state *State) Sub(key string, value uint64) error {
	fn, err := state.SubLazy(key, value)
	if err != nil {
		return err
	}

	fn()
	return nil
}

func (state *State) AddLazy(key string, value uint64) (func(), error) {
	if key == "" {
		return nil, errors.New("account is empty")
	}
	balance, ok := state.Load(key)
	if !ok {
		balance = interface{}(uint64(0))
	}

	amount := balance.(uint64) + value
	if amount < 0 {
		return nil, errors.New("not enough funds")
	}

	return func(){state.Store(key, amount)}, nil
}

func (state *State) SubLazy(key string, value uint64) (func(), error) {
	return state.AddLazy(key, -value)
}
