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

	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/controller"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/dhcp"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/ipxe"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/omni"
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
}

// New creates a new Provider.
func New(name, description, omniAPIEndpoint, imageFactoryPXEURL string, ipxeServerPort int,
	apiHost string, apiPort int, logger *zap.Logger,
) *Provider {
	return &Provider{
		name:               name,
		description:        description,
		omniAPIEndpoint:    omniAPIEndpoint,
		imageFactoryPXEURL: imageFactoryPXEURL,
		ipxeServerPort:     ipxeServerPort,
		apiHost:            apiHost,
		apiPort:            apiPort,
		logger:             logger,
	}
}

// Run runs the provider.
func (p *Provider) Run(ctx context.Context) error {
	omniClient, err := omni.BuildClient(p.omniAPIEndpoint)
	if err != nil {
		return fmt.Errorf("failed to build omni client: %w", err)
	}

	defer omniClient.Close() //nolint:errcheck

	if err = omniClient.EnsureProviderStatus(ctx, p.name, p.description, icon); err != nil {
		return fmt.Errorf("failed to create/update provider status: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	ctrller := controller.New(p.apiHost, p.apiPort, p.logger.With(zap.String("component", "controller")))
	dhcpProxy := dhcp.NewProxy(p.apiHost, p.apiPort, p.logger.With(zap.String("component", "dhcp_proxy")))
	ipxeServer := ipxe.NewServer(p.imageFactoryPXEURL, p.ipxeServerPort, p.logger.With(zap.String("component", "ipxe_server")))
	tftpServer := tftp.NewServer(p.logger.With(zap.String("component", "tftp_server")))

	eg.Go(func() error {
		return omniClient.RunReverseTunnel(ctx, ctrller, p.logger.With(zap.String("component", "reverse_tunnel")))
	})

	eg.Go(func() error {
		return ctrller.Run(ctx)
	})

	eg.Go(func() error {
		return dhcpProxy.Run()
	})

	eg.Go(func() error {
		return ipxeServer.Run(ctx)
	})

	eg.Go(func() error {
		return tftpServer.Run(ctx)
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("failed to run provider: %w", err)
	}

	return nil
}
