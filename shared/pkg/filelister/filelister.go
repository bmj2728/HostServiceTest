package filelister

type FileLister interface {
	ListFiles(dir string, hostService uint32) ([]string, error)
}
