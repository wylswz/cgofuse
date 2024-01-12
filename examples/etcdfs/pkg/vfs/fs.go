package vfs

import (
	"strings"

	"github.com/winfsp/cgofuse/fuse"
	"github.com/wylswz/etcdfs/pkg/args"
	"github.com/wylswz/etcdfs/pkg/engine"
)

var (
	_ fuse.FileSystemInterface = &EtcdFS{}

	onee64 uint64 = ^uint64(0)
)

func isDir(m uint32) bool {
	return (((m) & fuse.S_IFMT) == fuse.S_IFDIR)
}

type EtcdFS struct {
	fuse.FileSystemBase
	etcdEps []string
}

func (e EtcdFS) Init() {
	e.Opendir("/")
}

func (e EtcdFS) Destroy() {
	engine.GetEngine().Close()
}

func _p(path string) string {
	if path == "" || path == "/" {
		return path
	}
	if path[len(path)-1] == '/' {
		return path[:len(path)-1]
	}
	return path
}

func (e EtcdFS) Mknod(path string, mode uint32, dev uint64) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Mkdir(path string, mode uint32) int {
	path = _p(path)
	err := engine.GetEngine().Mkdir(path)
	if err == engine.ErrExists {
		return -fuse.EEXIST
	}

	if err == engine.ErrIsNotDir {
		return -fuse.EEXIST
	}

	if err != nil {
		return -fuse.ENOSYS
	}

	errno, _ := e.Opendir(path)
	return errno
}

func (e EtcdFS) Rmdir(path string) int {
	path = _p(path)
	handle, ok := LookUpFH(path)
	err := engine.GetEngine().Rmdir(path)
	//TODO: bugfix - removed dir still remain in memory
	if ok {
		PurgeAll(handle.path, handle.no)
	}
	if err != nil {
		return fuse.ENOSYS
	}

	return 0
}

func (e EtcdFS) Rename(oldpath string, newpath string) int {
	newpath = _p(newpath)
	oldpath = _p(oldpath)

	handle, ok := LookUpFH(oldpath)
	if !ok {
		return -fuse.ENOENT
	}
	var err error
	if !handle.IsDir() {
		err = engine.GetEngine().RenameFile(oldpath, newpath)
		// purge old file
		PurgeFH(handle.path, handle.no)
	} else {
		// we currently don't support rename directories
		subFiles, err := engine.GetEngine().List(oldpath)
		if err != nil {
			return -fuse.ENOSYS
		}
		for _, subFile := range subFiles {
			eno := e.Rename(subFile, strings.Replace(subFile, oldpath, newpath, 1))
			if eno != 0 {
				return eno
			}
		}
		engine.GetEngine().Rmdir(oldpath)
		PurgeAll(handle.path, handle.no)

	}

	// err handler goes here
	if err == engine.ErrDup {
		return -fuse.EEXIST
	}
	if err == engine.ErrNotExist {
		return -fuse.ENOENT
	}

	return 0

}

func (e EtcdFS) Chmod(path string, mode uint32) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Chown(path string, uid uint32, gid uint32) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Utimens(path string, tmsp []fuse.Timespec) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Access(path string, mask uint32) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Create(path string, flags int, mode uint32) (int, uint64) {
	path = _p(path)

	err := engine.GetEngine().NewFile(path)
	if err == engine.ErrExists {
		return -fuse.EEXIST, 0
	}
	if err != nil {
		return -fuse.ENOSYS, 0
	}
	fh := NewFH(path, false)
	return 0, fh.no
}

func shouldCreate(flags int) bool {
	return (flags&fuse.O_CREAT)|(flags&fuse.O_TRUNC)|(flags&fuse.O_EXCL) > 0
}

func openForWrite(flags int) bool {
	return (flags&fuse.O_WRONLY)|(flags&fuse.O_RDWR) > 0
}

func (e EtcdFS) Open(path string, flags int) (int, uint64) {
	path = _p(path)
	eno, handle := e.open(path, flags)
	if eno != 0 {
		return eno, 0
	}
	return 0, handle.no

}

func (e EtcdFS) open(path string, flags int) (int, *fileHandle) {
	path = _p(path)
	var fh *fileHandle
	if shouldCreate(flags) {
		engine.GetEngine().NewFile(path)
	}

	if openForWrite(flags) {
		fh = NewFH(path, engine.GetEngine().IsDir(path))
		_ = e.readthrough(fh)
		return 0, fh
	} else {
		if !engine.GetEngine().FileExist(path) {
			return -fuse.ENOENT, nil
		} else {
			fh = NewFH(path, engine.GetEngine().IsDir(path))
			_ = e.readthrough(fh)
			return 0, fh
		}
	}

}

func (e EtcdFS) readthrough(handle *fileHandle) int {
	eno := handle.ReadThrough(func() ([]byte, int) {
		content, err := engine.GetEngine().Read(handle.path)
		if err != nil {
			return nil, -fuse.ENOSYS
		}
		return content, 0
	})

	if eno != 0 {
		return eno
	}
	return 0
}

func (e EtcdFS) lookupNode(path string, fh uint64) (int, *fileHandle) {
	path = _p(path)

	if ^uint64(0) == fh {
		fh, ok := LookUpFH(path)
		if !ok {
			return -fuse.ENOENT, nil
		}
		return 0, fh
	} else {
		handle, ok := LookUpByNo(fh)
		if !ok {
			return -fuse.ENOENT, nil
		}
		return 0, handle
	}
}

func (e EtcdFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	path = _p(path)

	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		if eno == -fuse.ENOENT {
			if engine.GetEngine().DirExist(path) || engine.GetEngine().FileExist(path) {
				isDir := path == "/" || engine.GetEngine().IsDir(path)
				handle = NewFH(path, isDir)
			} else {
				return -fuse.ENOENT
			}
		} else {
			return eno
		}
	}

	*stat = *handle.Stats()
	return 0
}

func (e EtcdFS) Truncate(path string, size int64, fh uint64) int {
	path = _p(path)

	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		return eno
	}
	fh = handle.no
	_, _, pid := fuse.Getcontext()
	err := handle.Truncate(pid)
	if err == ErrConcurrentWrite {
		return -fuse.EPERM
	}
	return 0
}

func (e EtcdFS) Unlink(path string) int {
	path = _p(path)

	err := engine.GetEngine().Rm(path)
	handle, ok := LookUpFH(path)
	if ok {
		PurgeFH(handle.path, handle.no)
	}
	if err != nil {
		return -fuse.ENOSYS
	}
	return 0
}

func (e EtcdFS) Read(path string, buff []byte, ofst int64, fh uint64) int {
	path = _p(path)

	if engine.GetEngine().IsDir(path) {
		return -fuse.EISDIR
	}
	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		return eno
	}
	fh = handle.no

	_, _, pid := fuse.Getcontext()
	return handle.Read(pid, buff, ofst)
}

func (e EtcdFS) Write(path string, buff []byte, ofst int64, fh uint64) int {
	path = _p(path)

	if engine.GetEngine().IsDir(path) {
		return -fuse.EISDIR
	}
	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		return eno
	}

	_, _, pid := fuse.Getcontext()
	err := handle.AcquireForWrite(pid)
	defer handle.Release(pid)
	if err != nil {
		return -fuse.EPERM
	}
	return handle.Write(pid, buff, ofst)
}

func (e EtcdFS) Flush(path string, fh uint64) int {
	path = _p(path)

	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		return eno
	}
	_, _, pid := fuse.Getcontext()

	err := handle.AcquireForWrite(pid)
	defer handle.Release(pid)
	if err != nil {
		return -fuse.EPERM
	}

	return handle.FlushFinishWrite(pid, func(bts []byte) int {
		err := engine.GetEngine().Write(path, bts)
		if err != nil {
			return -fuse.ENOSYS
		}
		return 0
	})
}

func (e EtcdFS) Release(path string, fh uint64) int {
	path = _p(path)

	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		return 0
	}
	_, _, pid := fuse.Getcontext()
	handle.Release(pid)

	DeleteFH(path, handle.no)
	return 0
}

func (e EtcdFS) Fsync(path string, datasync bool, fh uint64) int {
	path = _p(path)
	// not supported
	return 0
}

func (e EtcdFS) Opendir(path string) (int, uint64) {
	path = _p(path)

	handle := NewFH(path, true)
	return 0, handle.no
}

func (e EtcdFS) Readdir(path string, fill func(name string, stat *fuse.Stat_t, ofst int64) bool, ofst int64, fh uint64) int {
	path = _p(path)

	nodes, err := engine.GetEngine().List(path)
	if err == engine.ErrNotExist {
		return -fuse.EEXIST
	}
	eno, handle := e.lookupNode(path, fh)
	if eno != 0 {
		return eno
	}
	fh = handle.no

	filled := fill(".", defaultStat(true, fh), 0)
	filled = fill("..", nil, 0)

	s := map[string]struct{}{}
	for _, node := range nodes {

		childPath, isDir := trimTail(path, node)

		if _, ok := s[childPath]; ok {
			continue
		}
		s[childPath] = struct{}{}

		// implicitly open children if we haven't done this yet
		// since node information is not preserved with data in etcd, so we have to build from scratch
		// whenever they are needed
		if isDir {
			_, cfh := e.Opendir(appendPath(path, childPath))
			if filled = fill(childPath, defaultStat(isDir, cfh), 0); !filled {
				break
			}
		} else {
			_, cHandle := e.open(appendPath(path, childPath), 0)
			if filled = fill(childPath, cHandle.Stats(), 0); !filled {
				break
			}
		}

	}

	return 0

}

func (e EtcdFS) Releasedir(path string, fh uint64) int {
	path = _p(path)
	// never release root
	if path == "/" {
		return 0
	}
	DeleteFH(path, fh)
	return 0
}

func (e EtcdFS) Fsyncdir(path string, datasync bool, fh uint64) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Setxattr(path string, name string, value []byte, flags int) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Getxattr(path string, name string) (int, []byte) {
	path = _p(path)
	return 0, []byte{}
}

func (e EtcdFS) Removexattr(path string, name string) int {
	path = _p(path)
	return 0
}

func (e EtcdFS) Listxattr(path string, fill func(name string) bool) int {
	path = _p(path)
	return 0
}

func CreateFS(argv []string) (*EtcdFS, error) {
	err := engine.Init(args.ResolveDatasource(argv))
	if err != nil {
		return nil, err
	}
	return &EtcdFS{}, nil
}
