package hostdemo

import (
	"context"
	"fmt"

	"github.com/bmj2728/hst/shared/pkg/hostconn"
	hostdemov1 "github.com/bmj2728/hst/shared/protogen/hostdemo/v1"
)

type GRPCServer struct {
	Impl HostDemo
	hostdemov1.UnimplementedHostDemoServer
}

func (s *GRPCServer) EstablishHostServices(ctx context.Context,
	request *hostdemov1.HostServiceRequest) (*hostdemov1.HostServiceResponse, error) {

	if hostConn, ok := s.Impl.(hostconn.HostConnection); ok {
		clientID, err := hostConn.EstablishHostServices(request.HostService)
		if err != nil {
			return nil, fmt.Errorf("plugin failed to establish host services: %w", err)
		}
		return &hostdemov1.HostServiceResponse{ClientId: clientID.String()}, nil
	}

	// Plugin doesn't implement HostConnection - not an error
	return &hostdemov1.HostServiceResponse{}, nil
}

func (s *GRPCServer) GetEnvDemo(_ context.Context, request *hostdemov1.GetEnvDemoReq) (*hostdemov1.GetEnvDemoResp, error) {
	if request == nil {
		return nil, fmt.Errorf("nil request for GetEnvDemo")
	}
	if request.Key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}
	demo, err := s.Impl.GetEnvDemo(request.Key)
	if err != nil {
		return nil, fmt.Errorf("failed get env demo: %w", err)
	}
	return &hostdemov1.GetEnvDemoResp{Resp: demo}, nil
}

func (s *GRPCServer) EnvDemo(_ context.Context, request *hostdemov1.EnvDemoReq) (*hostdemov1.EnvDemoResp, error) {
	if request == nil {
		return nil, fmt.Errorf("nil request for EnvDemo")
	}
	demo, err := s.Impl.EnvDemo()
	if err != nil {
		return nil, fmt.Errorf("failed env demo: %w", err)
	}
	return &hostdemov1.EnvDemoResp{Resp: demo}, nil
}

func (s *GRPCServer) TempDemo(_ context.Context, request *hostdemov1.TempDemoReq) (*hostdemov1.TempDemoResp, error) {
	if request == nil {
		return nil, fmt.Errorf("nil request for TempDemo")
	}
	if request.Pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}
	demo, err := s.Impl.TempDemo(request.Pattern, request.TextToWrite)
	if err != nil {
		return nil, fmt.Errorf("failed temp demo: %w", err)
	}
	return &hostdemov1.TempDemoResp{Resp: demo}, nil
}

func (s *GRPCServer) ReadFrankenstein(_ context.Context, request *hostdemov1.ReadFrankensteinReq) (*hostdemov1.ReadFrankensteinResp, error) {
	if request == nil {
		return nil, fmt.Errorf("nil request for ReadFrankenstein")
	}
	demo, err := s.Impl.ReadFrankenstein()
	if err != nil {
		return nil, fmt.Errorf("failed read frankenstein: %w", err)
	}
	return &hostdemov1.ReadFrankensteinResp{Frank: demo}, nil
}
