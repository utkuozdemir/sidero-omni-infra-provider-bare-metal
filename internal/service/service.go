// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package service implements the bare metal infra provider GRPC service server.
package service

import (
	"context"

	goipmi "github.com/pensando/goipmi"
	"go.uber.org/zap"

	"github.com/siderolabs/omni-infra-provider-bare-metal/api/provider"
	"github.com/siderolabs/omni-infra-provider-bare-metal/api/specs"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/resources"
)

const ipmiUsername = "talos-agent"

// OmniClient is the interface to manage persisted resources.Machine resources.
type OmniClient interface {
	GetMachine(ctx context.Context, id string) (*resources.Machine, error)
	SaveMachine(ctx context.Context, id string, spec *specs.MachineSpec) (*resources.Machine, error)
	RemoveMachine(ctx context.Context, id string) error
}

// AgentController is the interface to send commands to the Talos metal agent.
type AgentController interface {
	SetIPMICredentials(ctx context.Context, id, username string) (string, error)
	GetIPMIInfo(ctx context.Context, id string) (string, int, error)
}

// ProviderServiceServer is the bare metal infra provider service server.
type ProviderServiceServer struct {
	providerpb.UnimplementedProviderServiceServer

	logger          *zap.Logger
	agentController AgentController
	omniClient      OmniClient
}

// NewProviderServiceServer creates a new ProviderServiceServer.
func NewProviderServiceServer(agentController AgentController, omniClient OmniClient, logger *zap.Logger) *ProviderServiceServer {
	return &ProviderServiceServer{
		agentController: agentController,
		omniClient:      omniClient,
		logger:          logger,
	}
}

// ProvisionMachine provisions a machine.
func (p *ProviderServiceServer) ProvisionMachine(ctx context.Context, request *providerpb.ProvisionMachineRequest) (*providerpb.ProvisionMachineResponse, error) {
	p.logger.Info("provision machine", zap.String("machine_id", request.Id))

	password, err := p.agentController.SetIPMICredentials(ctx, request.Id, ipmiUsername)
	if err != nil {
		return nil, err
	}

	ipmiIP, ipmiPort, err := p.agentController.GetIPMIInfo(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	machineSpec := &specs.MachineSpec{
		IpmiIp:       ipmiIP,
		IpmiPort:     uint32(ipmiPort),
		IpmiPassword: password,
	}

	machine, err := p.omniClient.SaveMachine(ctx, request.Id, machineSpec)
	if err != nil {
		return nil, err
	}

	if err = withClient(machine, func(client *goipmi.Client) error {
		return client.Control(goipmi.ControlPowerUp)
	}); err != nil {
		return nil, err
	}

	return &providerpb.ProvisionMachineResponse{}, nil
}

// DeprovisionMachine deprovisions a machine.
func (p *ProviderServiceServer) DeprovisionMachine(ctx context.Context, request *providerpb.DeprovisionMachineRequest) (*providerpb.DeprovisionMachineResponse, error) {
	p.logger.Info("deprovision machine", zap.String("machine_id", request.Id))

	machine, err := p.omniClient.GetMachine(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	// remove the machine resource, then reboot, so we boot back into the agent mode
	if err = p.omniClient.RemoveMachine(ctx, request.Id); err != nil {
		return nil, err
	}

	if err = withClient(machine, func(client *goipmi.Client) error {
		return client.Control(goipmi.ControlPowerCycle)
	}); err != nil {
		return nil, err
	}

	return &providerpb.DeprovisionMachineResponse{}, nil
}

// PowerOnMachine powers on a machine.
func (p *ProviderServiceServer) PowerOnMachine(ctx context.Context, request *providerpb.PowerOnMachineRequest) (*providerpb.PowerOnMachineResponse, error) {
	p.logger.Info("power on machine", zap.String("machine_id", request.Id))

	machine, err := p.omniClient.GetMachine(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	if err = withClient(machine, func(client *goipmi.Client) error {
		return client.Control(goipmi.ControlPowerUp)
	}); err != nil {
		return nil, err
	}

	return &providerpb.PowerOnMachineResponse{}, nil
}

func withClient(machine *resources.Machine, f func(client *goipmi.Client) error) error {
	conn := &goipmi.Connection{
		Hostname:  machine.TypedSpec().Value.IpmiIp,
		Port:      int(machine.TypedSpec().Value.IpmiPort),
		Username:  ipmiUsername,
		Password:  machine.TypedSpec().Value.IpmiPassword,
		Interface: "lanplus",
	}

	client, err := goipmi.NewClient(conn)
	if err != nil {
		return err
	}

	defer client.Close() //nolint:errcheck

	return f(client)
}
