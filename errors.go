package vanilla_chain

import "errors"

var (
	ErrPrevHashEmpty   = errors.New("prev hash is empty")
	ErrNotHasPublicKey = errors.New("not specified public key")
	ErrTimeDried       = errors.New("time dried")
)

//Block
var (
	ErrBlockAlreadyExist = errors.New("block already exist")
	ErrBlocksNotEqual    = errors.New("blocks not equal")

	ErrVerifyNotPassed = errors.New("verify not passed")
)

//Transaction
var (
	ErrTransToEmpty         = errors.New("field 'to' not be empty")
	ErrTransFromEmpty       = errors.New("field 'from' not be empty")
	ErrTransAmountNotValid  = errors.New("field 'amount' not valid value")
	ErrTransNotHasSignature = errors.New("not has signature")
	ErrTransNotHasNeedSum   = errors.New("not has need sum")

	ErrTransAlreadyExist = errors.New("transaction already exist")
	ErrTransNotEqual     = errors.New("transaction not equal")
)

// Peers
var (
	ErrPeerNotFound = errors.New("peer not found")
)
