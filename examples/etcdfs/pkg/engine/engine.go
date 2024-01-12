package engine

var (
	impl Engine
)

type Engine interface {
	NewFile(path string) error
	Mkdir(argpath string) error
	Read(argpath string) ([]byte, error)
	Write(argpath string, value []byte) error
	Rm(argpath string) error
	Rmdir(argpath string) error
	List(argpath string) ([]string, error)
	IsDir(p string) bool
	FileExist(argpath string) bool
	DirExist(argpath string) bool
	RenameDir(oldPath, newPath string) error
	RenameFile(oldPath, newPath string) error

	Close()
}

func Init(connStr string) error {
	var err error

	ds, err := NewDatasource(connStr)
	if err != nil {
		return err
	}
	switch ds.GetScheme() {
	case EtcdImpl:
		impl, err = NewEtcdEngine(ds)
	default:
		err = ErrNotSupported
	}
	if err != nil {
		return err
	}
	return nil
}

func GetEngine() Engine {
	return impl
}
