package vfs

import "testing"

func TestTrimSlash(t *testing.T) {
	testCases := []struct {
		arg string
		ret string
	}{
		{"/", "/"},
		{"", ""},
		{"/a/b/c", "/a/b/c"},
		{"/a/", "/a"},
	}

	for _, testCase := range testCases {
		if _p(testCase.arg) != testCase.ret {
			t.Fail()
		}
	}
}
