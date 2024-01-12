package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
)

var (
	ctx = context.TODO()
)

type EtcdEngine struct {
	eps    []string
	c      *clientv3.Client
	exLock sync.Mutex
}

type EtcdNode struct {
	SubPath string
	IsDir   bool
}

func dir(s string) string {
	if s[len(s)-1] != '/' {
		return s + "/"
	}
	return s
}

func (e *EtcdEngine) synchronize() func() {
	e.exLock.Lock()
	return func() {
		e.exLock.Unlock()
	}
}

func (e *EtcdEngine) NewFile(path string) error {
	resp, err := e.c.Get(ctx, path)

	if err == nil && resp.Count > 0 {
		return ErrExists
	}

	_, err = e.c.Put(ctx, path, "")
	return err
}

func (e *EtcdEngine) Mkdir(argpath string) error {
	resp, err := e.c.Get(ctx, argpath)

	if err == nil && resp.Count > 0 {
		return ErrIsNotDir
	}
	if resp, err = e.c.Get(ctx, dir(argpath), clientv3.WithPrefix()); err == nil && resp.Count > 0 {
		return ErrExists
	}
	_, err = e.c.Put(ctx, dir(argpath), "")
	return err
}

func (e *EtcdEngine) Read(argpath string) ([]byte, error) {
	resp, err := e.c.Get(ctx, argpath)
	if err != nil {
		return nil, ErrSys
	}
	if resp.Count == 0 {
		return nil, ErrNotExist
	}

	return resp.Kvs[0].Value, nil
}

func (e *EtcdEngine) Write(argpath string, value []byte) error {
	_, err := e.c.Put(ctx, argpath, string(value))
	if err != nil {
		return err
	}
	return nil
}

func (e *EtcdEngine) Rm(argpath string) error {
	_, err := e.c.Delete(ctx, argpath)
	return err
}
func (e *EtcdEngine) Rmdir(argpath string) error {
	_, err := e.c.Delete(ctx, argpath)
	if err != nil {
		return e.resolveErr(err)
	}
	_, err = e.c.Delete(ctx, dir(argpath), clientv3.WithPrefix())
	if err != nil {
		return e.resolveErr(err)
	}
	return nil
}

// List all keys prefixed by ${argpath}/
func (e *EtcdEngine) List(argpath string) ([]string, error) {
	// TODO: handle sub dirs
	resp, err := e.c.Get(ctx, dir(argpath), clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, e.resolveErr(err)
	}
	var res []string
	for _, kv := range resp.Kvs {
		if string(kv.Key) == dir(argpath) {
			continue
		}
		res = append(res, string(kv.Key))
	}
	return res, nil
}

func (e *EtcdEngine) resolveErr(err error) error {
	if err == nil {
		return nil
	}

	rpcErr, ok := err.(rpctypes.EtcdError)
	if !ok {
		return err
	}

	if rpcErr.Code() == codes.NotFound {
		return ErrNotExist
	}

	switch rpcErr.Code() {
	case codes.NotFound:
		return ErrNotExist
	case codes.AlreadyExists:
		return ErrExists
	}

	return err
}
func (e *EtcdEngine) IsDir(p string) bool {
	resp, err := e.c.Get(ctx, dir(p), clientv3.WithCountOnly(), clientv3.WithPrefix())
	return err == nil && resp.Count > 0
}

func (e *EtcdEngine) FileExist(argpath string) bool {
	resp, err := e.c.Get(ctx, argpath, clientv3.WithIgnoreValue(), clientv3.WithCountOnly())
	return err == nil && resp.Count > 0
}

// Rename old dir to new one
// along with all sub paths
func (e *EtcdEngine) RenameDir(oldPath, newPath string) error {
	if e.DirExist(newPath) {
		return ErrDup
	}
	return nil
}

// RenameFile renames a single key
func (e *EtcdEngine) RenameFile(oldPath, newPath string) error {
	defer e.synchronize()()

	return e.doRenameFile(oldPath, newPath)
}

func (e *EtcdEngine) doRenameFile(oldPath, newPath string) error {
	if e.DirExist(newPath) {
		return ErrDup
	}
	content, err := e.Read(oldPath)
	if err != nil {
		return err
	}
	err = e.Write(newPath, content)

	return err
}

func (e *EtcdEngine) DirExist(argpath string) bool {
	resp, err := e.c.Get(ctx, dir(argpath), clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return false
	}
	return resp.Count > 0
}

func (e *EtcdEngine) Close() {
	_ = e.c.Close()
}

func NewEtcdEngine(ds *Datasource) (*EtcdEngine, error) {
	host := ds.GetHost()
	auth := ds.GetAuth()

	var eps []string
	for _, h := range strings.Split(host, ",") {
		eps = append(eps, fmt.Sprintf("%s:%s", h, ds.GetPort()))
	}

	config := clientv3.Config{
		Endpoints: eps,
	}
	if auth != nil {
		config.Username = auth.Username
		config.Password = auth.Password
	}
	c, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}
	return &EtcdEngine{
		c: c,
	}, nil
}
