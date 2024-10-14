// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package omni provides Omni-related functionality.
package omni

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/omni/resources/infra"
	"go.uber.org/zap"

	providerpb "github.com/siderolabs/omni-infra-provider-bare-metal/api/provider"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/controller"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/meta"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/service"
)

// Client is a wrapper around the Omni client.
type Client struct {
	omniClient *client.Client
}

// BuildClient creates a new Omni client.
func BuildClient(endpoint string) (*Client, error) {
	serviceAccountKey := os.Getenv("OMNI_SERVICE_ACCOUNT_KEY")

	var cliOpts []client.Option

	if serviceAccountKey != "" {
		cliOpts = append(cliOpts, client.WithServiceAccount(serviceAccountKey))
	}

	omniClient, err := client.New(endpoint, cliOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Omni client: %w", err)
	}

	return &Client{omniClient: omniClient}, nil
}

// Close closes the Omni client.
func (c *Client) Close() error {
	return c.omniClient.Close()
}

// EnsureProviderStatus makes sure that the infra.ProviderStatus resource exists and is up to date for this provider.
func (c *Client) EnsureProviderStatus(ctx context.Context, name, description string, rawIcon []byte) error {
	providerStatus := infra.NewProviderStatus(meta.ProviderID)

	providerStatus.TypedSpec().Value.Name = name
	providerStatus.TypedSpec().Value.Description = description
	providerStatus.TypedSpec().Value.Icon = base64.RawStdEncoding.EncodeToString(rawIcon)

	// todo: pull this label up into Omni or maybe introduce a new field on the resource
	providerStatus.Metadata().Labels().Set("omni.sidero.dev/is-static-provider", "")

	st := c.omniClient.Omni().State()

	if err := st.Create(ctx, providerStatus); err != nil {
		if !state.IsConflictError(err) {
			return err
		}

		_, err = safe.StateUpdateWithConflicts(ctx, st, providerStatus.Metadata(), func(res *infra.ProviderStatus) error {
			res.TypedSpec().Value = providerStatus.TypedSpec().Value

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// RunReverseTunnel starts the reverse GRPC tunnel to Omni.
func (c *Client) RunReverseTunnel(ctx context.Context, controller *controller.Controller, logger *zap.Logger) error {
	reverseTunnelServer := c.omniClient.Tunnel()

	providerServiceServer := service.NewProviderServiceServer(controller, logger)

	providerpb.RegisterProviderServiceServer(reverseTunnelServer, providerServiceServer)

	// Open the reverse tunnel and serve requests.
	if _, err := reverseTunnelServer.Serve(ctx); err != nil {
		return fmt.Errorf("failed to serve reverse tunnel: %w", err)
	}

	return nil
}
