package vanilla_chain

import "testing"

func TestState_Add(t *testing.T) {
	var state State
	state.Add("24", 23)
}
