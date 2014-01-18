package test

import (
	"testing"
)

func my_func(pa int) int {
	return pa * 10
}

func TestDeclare(t *testing.T) {
	v := my_func(10)
	if v != 100 {
		t.Errorf("expect 100, actual %v", v)
	}
}


