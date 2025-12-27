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

func (s *GRPCServer) GetEnvDemo(ctx context.Context, request *hostdemov1.GetEnvDemoReq) (*hostdemov1.GetEnvDemoResp, error) {
	if request == nil {
		return nil, fmt.Errorf("nil request for GetEnvDemo")
	}
	if request.Key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}
	demo, err := s.Impl.GetEnvDemo(ctx, request.Key)
	if err != nil {
		return nil, fmt.Errorf("failed get env demo: %w", err)
	}
	return &hostdemov1.GetEnvDemoResp{Resp: demo}, nil
}

func (s *GRPCServer) EnvDemo(ctx context.Context, request *hostdemov1.EnvDemoReq) (*hostdemov1.EnvDemoResp, error) {
	if request == nil {
		return nil, fmt.Errorf("nil request for EnvDemo")
	}
	demo, err := s.Impl.EnvDemo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed env demo: %w", err)
	}
	return &hostdemov1.EnvDemoResp{Resp: demo}, nil
}
