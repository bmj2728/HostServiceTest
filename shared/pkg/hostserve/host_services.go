package hostserve

type IHostServices interface {
	IHostFS
	IHostEnv
}

type HostServices struct {
	IHostFS
	IHostEnv
}

func NewHostServices(fs IHostFS, env IHostEnv) *HostServices {
	return &HostServices{
		IHostFS:  fs,
		IHostEnv: env,
	}
}
