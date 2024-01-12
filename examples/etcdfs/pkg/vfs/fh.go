package vfs

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/winfsp/cgofuse/fuse"
)

var cnt = atomic.Uint64{}

var (
	ErrConcurrentWrite = errors.New("concurrent write")
)

type fileHandle struct {
	path string
	no   uint64

	stats *fuse.Stat_t

	// read snapshot buffer
	buf []byte

	// copy on write buffer
	cowBuf    []byte
	writerPid int
	lock      sync.RWMutex

	opens   int
	flagMap map[int]int
}

var store = map[string]*fileHandle{}

var rStore = map[uint64]*fileHandle{}

var mux sync.RWMutex

func defaultStat(isDir bool, fh uint64) *fuse.Stat_t {
	stat := fuse.Stat_t{}
	uid, gid, _ := fuse.Getcontext()
	stat.Uid = uid
	stat.Gid = gid
	stat.Ino = fh

	stat.Flags = 0

	if isDir {
		stat.Size = 4096
		stat.Mode = fuse.S_IRWXG | fuse.S_IRWXU | fuse.S_IRWXO | fuse.S_IFDIR
	} else {
		stat.Mode = 00666 | fuse.S_IFREG
		stat.Size = 4096

	}
	return &stat
}

func (f *fileHandle) Stats() *fuse.Stat_t {
	if f.cowBuf == nil {
		f.stats.Size = int64(len(f.buf))
	} else {
		f.stats.Size = int64(len(f.cowBuf))
	}
	return f.stats
}

func NewFH(argpath string, isDir bool) *fileHandle {
	mux.Lock()
	defer mux.Unlock()
	defer func() {
		cnt.Add(1)
	}()
	if existing, ok := store[argpath]; ok {
		existing.opens++
		return existing
	}
	no := cnt.Load()

	handle := &fileHandle{
		no:      no,
		path:    argpath,
		stats:   defaultStat(isDir, no),
		flagMap: map[int]int{},
	}

	store[argpath] = handle
	rStore[handle.no] = handle

	return handle
}

func PurgeFH(argpath string, fh uint64) {
	mux.Lock()
	defer mux.Unlock()

	delete(store, argpath)
	delete(rStore, fh)
}

// PurgeAll files under the directory and directory itself
func PurgeAll(dirPath string, fh uint64) {
	mux.Lock()
	defer mux.Unlock()

	for path, handle := range store {
		if contains(dirPath, path) {
			delete(store, path)
			delete(rStore, handle.no)
		}
	}
	delete(store, dirPath)
	delete(rStore, fh)
}

func contains(dir, path string) bool {
	return false
}

func DeleteFH(argpath string, fh uint64) {
	mux.Lock()
	defer mux.Unlock()

	handle, ok := store[argpath]
	if ok {
		handle.opens--
		if handle.opens == 0 {
			delete(store, argpath)
			delete(rStore, fh)
		}
	}
}

func LookUpByNo(fh uint64) (*fileHandle, bool) {
	handle, ok := rStore[fh]
	if ok {
		return handle, true
	}
	return nil, false
}

func LookUpFH(argpath string) (fh *fileHandle, ok bool) {
	handle, ok := store[argpath]
	if ok {
		return handle, true
	}
	return nil, false
}

func WithFlag(pid int, flags int) {

}

func (f *fileHandle) sync() func() {
	f.lock.Lock()
	return func() {
		f.lock.Unlock()
	}
}

func (f *fileHandle) IsDir() bool {
	return f.stats.Mode&fuse.S_IFDIR > 0
}

func (f *fileHandle) Rst(buf []byte) {
	defer f.sync()()
	f.buf = make([]byte, len(buf))
	copy(f.buf, buf)
}

func (f *fileHandle) ReadThrough(reader func() ([]byte, int)) int {
	defer f.sync()()
	if f.buf == nil {
		var eno int
		f.buf, eno = reader()
		return eno
	}
	return 0
}

func (f *fileHandle) FlushFinishWrite(pid int, drain func([]byte) int) int {
	if f.cowBuf == nil {
		return 0
	}
	eno := drain(f.cowBuf)
	if eno != 0 {
		return eno
	}
	if f.canWrite(pid) && f.cowBuf != nil {
		f.Rst(f.cowBuf)
	}
	f.cowBuf = nil
	return 0
}

func (f *fileHandle) Read(pid int, buf []byte, ofst int64) int {
	if f.canWrite(pid) && f.cowBuf != nil {
		return copy(buf, f.cowBuf[ofst:])
	}
	return copy(buf, f.buf[ofst:])
}

func (f *fileHandle) Write(pid int, content []byte, ofst int64) int {
	if !f.canWrite(pid) {
		return -fuse.EPERM
	}

	if f.cowBuf == nil {
		f.cowBuf = make([]byte, len(f.buf))
		copy(f.cowBuf, f.buf)
	}

	n := copy(f.cowBuf[ofst:], content)
	f.cowBuf = append(f.cowBuf, content[n:]...)
	return len(content)

}

func (f *fileHandle) canWrite(pid int) bool {
	return f.writerPid == 0 || f.writerPid == pid
}

func (f *fileHandle) Truncate(pid int) error {
	if !f.canWrite(pid) {
		return ErrConcurrentWrite
	}
	f.cowBuf = nil
	return nil
}

func (f *fileHandle) AcquireForWrite(pid int) error {
	defer f.sync()()
	if !f.canWrite(pid) {
		return ErrConcurrentWrite
	}
	f.writerPid = pid
	return nil
}

func (f *fileHandle) Release(pid int) {
	defer f.sync()()
	if f.writerPid == pid || pid == 0 {
		f.writerPid = 0
	}
}

func concat(arrs ...[]byte) []byte {
	var res []byte
	for _, arr := range arrs {
		res = append(res, arr...)
	}
	return res
}

func deleteAll(src []uint64, item uint64) []uint64 {
	var res []uint64
	for _, v := range src {
		if v != item {
			res = append(res, v)
		}
	}
	return res
}
