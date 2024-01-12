package vfs

import (
	"fmt"
	"testing"
)

func TestCopy(t *testing.T) {
	a := []int{1, 2, 3}
	n := copy(a[1:], []int{4, 4, 4})
	fmt.Println(n)
	fmt.Println(a)
}
