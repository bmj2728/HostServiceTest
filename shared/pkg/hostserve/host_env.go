package hostserve

import "os"

type HostEnv struct {
	//TBD fields
}

func NewHostEnv() *HostEnv {
	return &HostEnv{}
}

func (he *HostEnv) GetEnv(key string) string {
	return os.Getenv(key)
}
