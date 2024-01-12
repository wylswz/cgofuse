package vfs

import (
	"testing"

	"github.com/wylswz/etcdfs/pkg/testutils"
)

func TestTrimTail(t *testing.T) {
	testCases := []struct {
		parent   string
		path     string
		resName  string
		resIsDir bool
	}{
		{"/a/b/c", "/a/b/c/d", "d", false},
		{"/", "/a", "a", false},
		{"/", "/a/b", "a", true},
		{"/a", "/a/b", "b", false},
		{"/a", "/a/b/c", "b", true},
	}

	for _, testCase := range testCases {
		name, isDir := trimTail(testCase.parent, testCase.path)
		testutils.AssertEq(t, name, testCase.resName)
		testutils.AssertEq(t, isDir, testCase.resIsDir)
	}

}
