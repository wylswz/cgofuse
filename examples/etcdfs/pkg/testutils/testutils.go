package testutils

import (
	"reflect"
	"testing"
)

func AssertEq(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected %v, actual: %v", a, b)
	}
}
