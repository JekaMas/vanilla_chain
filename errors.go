package vanilla_chain

import "errors"

var (
	ErrPrevHashEmpty = errors.New("prev hash is empty")

	ErrBlockAlreadyExist = errors.New("block already exist")
)
