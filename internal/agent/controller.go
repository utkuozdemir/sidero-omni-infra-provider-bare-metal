// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package agent implements the metal agent controller.
package agent

import (
	"context"

	"github.com/jhump/grpctunnel"
	"github.com/jhump/grpctunnel/tunnelpb"
	agentpb "github.com/siderolabs/talos-metal-agent/api/agent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Controller controls servers by establishing a reverse GRPC tunnel with them and by sending them commands.
type Controller struct {
	logger        *zap.Logger
	grpcServer    grpc.ServiceRegistrar
	tunnelHandler *grpctunnel.TunnelServiceHandler
}

// NewController creates a new agent Controller.
func NewController(grpcServer grpc.ServiceRegistrar, logger *zap.Logger) *Controller {
	tunnelHandler := grpctunnel.NewTunnelServiceHandler(
		grpctunnel.TunnelServiceHandlerOptions{},
	)

	tunnelpb.RegisterTunnelServiceServer(grpcServer, tunnelHandler.Service())

	return &Controller{
		logger:        logger,
		grpcServer:    grpcServer,
		tunnelHandler: tunnelHandler,
	}
}

// SetIPMICredentials sets the IPMI credentials on the server with the given ID and returns the password was set.
func (c *Controller) SetIPMICredentials(ctx context.Context, id, username string) (string, error) {
	channel := c.tunnelHandler.KeyAsChannel(id)
	cli := agentpb.NewAgentServiceClient(channel)

	response, err := cli.SetIPMICredentials(ctx, &agentpb.SetIPMICredentialsRequest{
		Username: username,
	})
	if err != nil {
		return "", err
	}

	return response.Password, nil
}

// GetIPMIInfo retrieves the IPMI information from the server with the given ID.
func (c *Controller) GetIPMIInfo(ctx context.Context, id string) (ip string, port int, err error) {
	channel := c.tunnelHandler.KeyAsChannel(id)
	cli := agentpb.NewAgentServiceClient(channel)

	response, err := cli.GetIPMIInfo(ctx, &agentpb.GetIPMIInfoRequest{})
	if err != nil {
		return "", 0, err
	}

	return response.Ip, int(response.Port), nil
}
