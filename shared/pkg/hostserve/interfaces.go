package hostserve

import "io/fs"

type IHostServices interface {
	IHostFS
	IHostEnv
}

type IHostFS interface {
	ReadDir(path string) ([]fs.DirEntry, error)
	ReadFile(dir, file string) ([]byte, error)
}

type IHostEnv interface {
	GetEnv(key string) string
}
