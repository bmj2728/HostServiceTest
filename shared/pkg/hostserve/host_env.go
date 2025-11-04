package hostserve

import "os"

type IHostEnv interface {
	GetEnv(key string) string
}

type HostEnv struct {
	//TBD fields
}

func (he *HostEnv) GetEnv(key string) string {
	return os.Getenv(key)
}
