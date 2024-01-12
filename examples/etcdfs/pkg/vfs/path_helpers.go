package vfs

import (
	"path"
	"strings"
)

const (
	splitter  = "/"
	splitterC = '/'
)

// trim tail of argpath
// parent=/a/b/c, argpath=/a/b/c/d/e/f/g, it should return a
func trimTail(parent string, argpath string) (string, bool) {
	parentDir := parent
	if parent != "/" {
		parentDir = parent + "/"
	}
	rPath := strings.TrimPrefix(argpath, parentDir)
	splitted := strings.Split(rPath, "/")
	return splitted[0], len(splitted) > 1
}

func removeEmpty(ss []string) []string {
	var res []string
	for _, s := range ss {
		if s != "" {
			res = append(res, s)
		}
	}
	return res
}

func appendPath(a, b string) string {
	return path.Join(a, b)
}
