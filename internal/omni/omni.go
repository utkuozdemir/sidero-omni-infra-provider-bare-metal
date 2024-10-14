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
	"github.com/siderolabs/omni/client/pkg/jointoken"
	"github.com/siderolabs/omni/client/pkg/omni/resources/infra"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/client/pkg/omni/resources/siderolink"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	providerpb "github.com/siderolabs/omni-infra-provider-bare-metal/api/provider"
	"github.com/siderolabs/omni-infra-provider-bare-metal/api/specs"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/meta"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/resources"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/service"
)

// Client is a wrapper around the Omni client.
type Client struct {
	omniClient *client.Client
}

// BuildClient creates a new Omni client.
func BuildClient(endpoint string, insecureSkipTLSVerify bool) (*Client, error) {
	serviceAccountKey := os.Getenv("OMNI_SERVICE_ACCOUNT_KEY")

	cliOpts := []client.Option{
		client.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
	}

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

// GetSiderolinkAPIURL returns the SideroLink API URL.
func (c *Client) GetSiderolinkAPIURL(ctx context.Context) (string, error) {
	st := c.omniClient.Omni().State()

	connectionParams, err := safe.StateGetByID[*siderolink.ConnectionParams](ctx, st, siderolink.ConfigID)
	if err != nil {
		return "", fmt.Errorf("failed to get connection params: %w", err)
	}

	token, err := jointoken.NewWithExtraData(connectionParams.TypedSpec().Value.JoinToken, map[string]string{
		omni.LabelInfraProviderID:             meta.ProviderID,
		"omni.sidero.dev/requires-acceptance": "", // todo: pull this label up into Omni
	})
	if err != nil {
		return "", err
	}

	tokenString, err := token.Encode()
	if err != nil {
		return "", fmt.Errorf("failed to encode the siderolink token: %w", err)
	}

	apiURL, err := siderolink.APIURL(connectionParams, siderolink.WithJoinToken(tokenString))
	if err != nil {
		return "", fmt.Errorf("failed to build API URL: %w", err)
	}

	return apiURL, nil
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
func (c *Client) RunReverseTunnel(ctx context.Context, agentController service.AgentController, logger *zap.Logger) error {
	reverseTunnelServer := c.omniClient.Tunnel()
	providerServiceServer := service.NewProviderServiceServer(agentController, c, logger)

	providerpb.RegisterProviderServiceServer(reverseTunnelServer, providerServiceServer)

	// Open the reverse tunnel and serve requests.
	if _, err := reverseTunnelServer.Serve(ctx); err != nil {
		if status.Code(err) == codes.Canceled {
			return nil
		}

		return fmt.Errorf("failed to serve reverse tunnel: %w", err)
	}

	return nil
}

// GetMachine returns the machine with the given ID from the persistent state.
func (c *Client) GetMachine(ctx context.Context, id string) (*resources.Machine, error) {
	machine, err := safe.StateGetByID[*resources.Machine](ctx, c.omniClient.Omni().State(), id)
	if err != nil {
		return nil, err
	}

	return machine, nil
}

// SaveMachine saves the machine with the given ID and spec to the persistent state.
func (c *Client) SaveMachine(ctx context.Context, id string, spec *specs.MachineSpec) (*resources.Machine, error) {
	st := c.omniClient.Omni().State()
	machine := resources.NewMachine(id)
	machine.TypedSpec().Value = spec

	if err := st.Create(ctx, machine); err != nil {
		if !state.IsConflictError(err) {
			return nil, err
		}

		machine, err = safe.StateUpdateWithConflicts(ctx, st, machine.Metadata(), func(res *resources.Machine) error {
			res.TypedSpec().Value = spec

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return machine, nil
}

// RemoveMachine removes the machine from the persistent state with the given ID.
func (c *Client) RemoveMachine(ctx context.Context, id string) error {
	st := c.omniClient.Omni().State()

	if err := st.Destroy(ctx, resources.NewMachine(id).Metadata()); err != nil {
		if !state.IsNotFoundError(err) {
			return err
		}
	}

	return nil
}
