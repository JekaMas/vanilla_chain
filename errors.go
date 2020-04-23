package vanilla_chain

import "errors"

var (
	ErrPrevHashEmpty = errors.New("prev hash is empty")

	ErrBlockAlreadyExist = errors.New("block already exist")
)

//Transaction
var (
	ErrTransToEmpty         = errors.New("field 'to' not be empty")
	ErrTransFromEmpty       = errors.New("field 'from' not be empty")
	ErrTransAmountNotValid  = errors.New("field 'amount' not valid value")
	ErrTransNotHasSignature = errors.New("not has signature")
	ErrTransNotHasNeedSum   = errors.New("not has need sum")
)
