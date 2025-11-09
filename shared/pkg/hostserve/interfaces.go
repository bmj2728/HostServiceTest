package hostserve

import "io/fs"

// IHostServices is an interface that combines IHostFS and IHostEnv to provide file system and environment services.
type IHostServices interface {
	IHostFS
	IHostEnv
}

// IHostFS is an interface that defines methods to interact with the host file system.
type IHostFS interface {

	// ReadDir reads the contents of the directory specified by `path` and returns a slice of directory
	// entries or an error.
	ReadDir(path string) ([]fs.DirEntry, error)

	// ReadFile reads the specified file from the given directory and returns its contents as a byte slice or an error.
	ReadFile(dir, file string) ([]byte, error)
}

// IHostEnv defines a contract for interacting with environment variables in the host system.
type IHostEnv interface {

	// GetEnv fetches the value of an environment variable by its key and returns it as a string.
	GetEnv(key string) string
}
