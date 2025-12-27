package hostdemo

import (
	"context"
	"fmt"

	"github.com/bmj2728/hst/shared/pkg/hostserve"
	"github.com/bmj2728/hst/shared/protogen/hostdemo/v1"
	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	client        hostdemov1.HostDemoClient
	broker        *plugin.GRPCBroker
	hostServiceID uint32
}

func (c *GRPCClient) SetBroker(broker *plugin.GRPCBroker) {
	c.broker = broker
}

func (c *GRPCClient) EstablishHostServices(hostServiceID uint32) (hostserve.ClientID, error) {
	c.hostServiceID = hostServiceID

	resp, err := c.client.EstablishHostServices(context.Background(),
		&hostdemov1.HostServiceRequest{
			HostService: hostServiceID,
		})
	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err) // CHANGED - return error with context
	}

	return hostserve.ClientID(resp.ClientId), nil
}

func (c *GRPCClient) DisconnectHostServices() {
	// The host manages its own server lifecycle
	// This is called during plugin shutdown to do any cleanup
	// Currently no cleanup needed on the client side
}

func (c *GRPCClient) RegisterHostService(hostServices hostserve.IHostServices) (uint32, error) {
	// Allocate a unique ID for this service using the broker's built-in ID allocator
	serviceID := c.broker.NextId()

	// Start a gRPC server for the host service via the broker at the allocated ID
	go c.broker.AcceptAndServe(serviceID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		hostservev1.RegisterHostServiceServer(server, &hostserve.HostServiceGRPCServer{
			Impl: hostServices,
		})
		return server
	})

	return serviceID, nil
}

func (c *GRPCClient) GetEnvDemo(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}
	resp, err := c.client.GetEnvDemo(ctx, &hostdemov1.GetEnvDemoReq{Key: key})
	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("nil response from GetEnvDemo")
	}
	if resp.Error != nil {
		return "", fmt.Errorf("gRPC call failed: %v", resp.Error)
	}
	return resp.Resp, nil
}

func (c *GRPCClient) EnvDemo(ctx context.Context) (string, error) {
	resp, err := c.client.EnvDemo(ctx, &hostdemov1.EnvDemoReq{})
	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("nil response from EnvDemo")
	}
	if resp.Error != nil {
		return "", fmt.Errorf("gRPC call failed: %v", resp.Error)
	}
	return resp.Resp, nil
}
