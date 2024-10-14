// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package provider implements the bare metal infra provider.
package provider

import (
	"context"
	_ "embed"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/agent"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/config"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/dhcp"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/ipxe"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/omni"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/server"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/tftp"
)

//go:embed data/icon.svg
var icon []byte

// Provider implements the bare metal infra provider.
type Provider struct {
	logger *zap.Logger

	apiHost            string
	imageFactoryPXEURL string
	name               string
	description        string
	omniAPIEndpoint    string

	apiPort        int
	ipxeServerPort int

	insecureSkipTLSVerify bool
}

// New creates a new Provider.
func New(name, description, omniAPIEndpoint, imageFactoryPXEURL string, ipxeServerPort int, apiHost string, apiPort int, insecureSkipTLSVerify bool, logger *zap.Logger) *Provider {
	return &Provider{
		name:                  name,
		description:           description,
		omniAPIEndpoint:       omniAPIEndpoint,
		imageFactoryPXEURL:    imageFactoryPXEURL,
		ipxeServerPort:        ipxeServerPort,
		apiHost:               apiHost,
		apiPort:               apiPort,
		insecureSkipTLSVerify: insecureSkipTLSVerify,
		logger:                logger,
	}
}

// Run runs the provider.
func (p *Provider) Run(ctx context.Context) error {
	omniClient, err := omni.BuildClient(p.omniAPIEndpoint, p.insecureSkipTLSVerify)
	if err != nil {
		return fmt.Errorf("failed to build omni client: %w", err)
	}

	defer omniClient.Close() //nolint:errcheck

	if err = omniClient.EnsureProviderStatus(ctx, p.name, p.description, icon); err != nil {
		return fmt.Errorf("failed to create/update provider status: %w", err)
	}

	ipxeHandler, err := ipxe.NewHandler(p.imageFactoryPXEURL, p.ipxeServerPort, p.logger.With(zap.String("component", "ipxe_handler")))
	if err != nil {
		return fmt.Errorf("failed to create iPXE handler: %w", err)
	}

	configHandler, err := config.NewHandler(ctx, omniClient, p.logger.With(zap.String("component", "config_handler")))
	if err != nil {
		return fmt.Errorf("failed to create config handler: %w", err)
	}

	srvr := server.New(p.apiHost, p.apiPort, configHandler, ipxeHandler, p.logger.With(zap.String("component", "server")))
	agentController := agent.NewController(srvr, p.logger.With(zap.String("component", "controller")))
	dhcpProxy := dhcp.NewProxy(p.apiHost, p.apiPort, p.logger.With(zap.String("component", "dhcp_proxy")))
	tftpServer := tftp.NewServer(p.logger.With(zap.String("component", "tftp_server")))

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(p.runComponent("server", func() error {
		return srvr.Run(ctx)
	}))

	eg.Go(p.runComponent("reverse tunnel", func() error {
		return omniClient.RunReverseTunnel(ctx, agentController, p.logger.With(zap.String("component", "reverse_tunnel")))
	}))

	eg.Go(p.runComponent("DHCP proxy", func() error {
		return dhcpProxy.Run(ctx)
	}))

	eg.Go(p.runComponent("TFTP server", func() error {
		return tftpServer.Run(ctx)
	}))

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("failed to run provider: %w", err)
	}

	return nil
}

func (p *Provider) runComponent(name string, f func() error) func() error {
	return func() error {
		p.logger.Info("start component ", zap.String("name", name))

		err := f()
		if err != nil {
			p.logger.Error("failed to run component", zap.String("name", name), zap.Error(err))

			return err
		}

		p.logger.Info("component stopped", zap.String("name", name))

		return nil
	}
}
