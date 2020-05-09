package vanilla_chain

import "testing"

func TestState_Add(t *testing.T) {
	var state State
	key := "24"
	v := uint64(23)
	err := state.Add(key, v)
	if err != nil {
		t.Fatal(err)
	}

	v1, ok := state.Get(key)
	if !ok || v1 != v {
		t.Fatalf("incorrect stage. Got key: %v. value %d, expected %d", ok, v1, v)
	}

	err = state.Add(key, v)
	if err != nil {
		t.Fatal(err)
	}

	v1, ok = state.Get(key)
	if !ok || v1 != 2*v {
		t.Fatalf("incorrect stage. Got key: %v. value %d, expected %d", ok, v1, v)
	}

	state.Set(key, 1)

	v1, ok = state.Get(key)
	if !ok || v1 != 1 {
		t.Fatalf("incorrect stage. Got key: %v. value %d, expected %d", ok, v1, 1)
	}
}
