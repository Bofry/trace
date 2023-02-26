package trace

import "testing"

func Test__InvalidKeyValue(t *testing.T) {
	ok := __InvalidKeyValue.Valid()
	expectedOk := false
	if ok != expectedOk {
		t.Errorf("__InvalidKeyValue.Valid(): expect %v, but got %v.", expectedOk, ok)
	}
}
