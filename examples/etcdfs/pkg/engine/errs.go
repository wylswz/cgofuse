package engine

import "fmt"

var (
	ErrExists   = fmt.Errorf("exists")
	ErrNotExist = fmt.Errorf("not exist")
	ErrIsNotDir = fmt.Errorf("is not dir")
	ErrSys      = fmt.Errorf("cannot access etcd")
	ErrDup      = fmt.Errorf("target file already exists")

	ErrBadUrl       = fmt.Errorf("bad url")
	ErrNotSupported = fmt.Errorf("implementation not supported")
)
