// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package controller implements the metal agent controller.
package controller

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/jhump/grpctunnel"
	"github.com/jhump/grpctunnel/tunnelpb"
	agentpb "github.com/siderolabs/talos-metal-agent/api/agent"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Controller controls servers by establishing a reverse GRPC tunnel with them and by sending them commands.
type Controller struct {
	logger        *zap.Logger
	grpcServer    *grpc.Server
	tunnelHandler *grpctunnel.TunnelServiceHandler

	apiHost string
	apiPort int
}

// New creates a new Controller.
func New(apiHost string, apiPort int, logger *zap.Logger) *Controller {
	recoveryOption := recovery.WithRecoveryHandler(recoveryHandler(logger))

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(recovery.UnaryServerInterceptor(recoveryOption)),
		grpc.ChainStreamInterceptor(recovery.StreamServerInterceptor(recoveryOption)),
		grpc.Creds(insecure.NewCredentials()),
	)

	return &Controller{
		apiHost:    apiHost,
		apiPort:    apiPort,
		logger:     logger,
		grpcServer: grpcServer,
		tunnelHandler: grpctunnel.NewTunnelServiceHandler(
			grpctunnel.TunnelServiceHandlerOptions{},
		),
	}
}

// Run runs the controller, establishing starting the reverse GRPC proxy server.
func (c *Controller) Run(ctx context.Context) error {
	listenAddress := net.JoinHostPort(c.apiHost, strconv.Itoa(c.apiPort))

	c.logger.Info("starting metal agent grpcServer", zap.String("listen_address", listenAddress))

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen address %q: %w", listenAddress, err)
	}

	recoveryOption := recovery.WithRecoveryHandler(recoveryHandler(c.logger))

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(recovery.UnaryServerInterceptor(recoveryOption)),
		grpc.ChainStreamInterceptor(recovery.StreamServerInterceptor(recoveryOption)),
		grpc.Creds(insecure.NewCredentials()),
	)

	tunnelpb.RegisterTunnelServiceServer(server, c.tunnelHandler.Service())

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		serveErr := server.Serve(listener)
		if serveErr == nil || errors.Is(serveErr, context.Canceled) {
			return nil
		}

		return fmt.Errorf("failed to serve: %w", serveErr)
	})

	eg.Go(func() error {
		<-ctx.Done()

		c.logger.Info("stopping metal agent grpcServer")

		stopGRPCServer(server)

		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("failed to wait: %w", err)
	}

	return nil
}

// SetIPMICredentials sets the IPMI credentials on the server with the given ID and returns the password was set.
func (c *Controller) SetIPMICredentials(ctx context.Context, id string) (string, error) {
	channel := c.tunnelHandler.KeyAsChannel(id)
	cli := agentpb.NewAgentServiceClient(channel)

	response, err := cli.SetIPMICredentials(ctx, &agentpb.SetIPMICredentialsRequest{})
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

func recoveryHandler(logger *zap.Logger) recovery.RecoveryHandlerFunc {
	return func(p any) error {
		if logger != nil {
			logger.Error("grpc panic", zap.Any("panic", p), zap.Stack("stack"))
		}

		return status.Errorf(codes.Internal, "%v", p)
	}
}

// stopGRPCServer stops the GRPC server with a timeout.
func stopGRPCServer(server *grpc.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stopped := make(chan struct{})

	go func() {
		server.GracefulStop()

		close(stopped)
	}()

	select {
	case <-shutdownCtx.Done():
	case <-stopped:
	}

	server.Stop()
}
