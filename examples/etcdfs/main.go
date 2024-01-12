package main

import (
	"github.com/winfsp/cgofuse/fuse"
	"github.com/wylswz/etcdfs/pkg/vfs"
	"os"
)

func main() {
	args := os.Args

	fs, err := vfs.CreateFS(args)
	if err != nil {
		panic(err)
	}

	host := fuse.NewFileSystemHost(fs)
	host.Mount("", args[1:])
}
