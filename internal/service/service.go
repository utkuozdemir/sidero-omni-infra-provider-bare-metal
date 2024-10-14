// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package service implements the bare metal infra provider GRPC service server.
package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/siderolabs/omni-infra-provider-bare-metal/api/provider"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/controller"
)

// ProviderServiceServer is the bare metal infra provider service server.
type ProviderServiceServer struct {
	providerpb.UnimplementedProviderServiceServer

	logger     *zap.Logger
	controller *controller.Controller
}

// NewProviderServiceServer creates a new ProviderServiceServer.
func NewProviderServiceServer(controller *controller.Controller, logger *zap.Logger) *ProviderServiceServer {
	return &ProviderServiceServer{
		controller: controller,
		logger:     logger,
	}
}

// ProvisionMachine provisions a machine.
func (p *ProviderServiceServer) ProvisionMachine(_ context.Context, request *providerpb.ProvisionMachineRequest) (*providerpb.ProvisionMachineResponse, error) {
	// todo: implement
	p.logger.Info("provisioning machine", zap.String("machine_id", request.Id))

	return &providerpb.ProvisionMachineResponse{}, nil
}

// DeprovisionMachine deprovisions a machine.
func (p *ProviderServiceServer) DeprovisionMachine(_ context.Context, request *providerpb.DeprovisionMachineRequest) (*providerpb.DeprovisionMachineResponse, error) {
	// todo: implement
	p.logger.Info("deprovisioning machine", zap.String("machine_id", request.Id))

	return &providerpb.DeprovisionMachineResponse{}, nil
}

// PowerOnMachine powers on a machine.
func (p *ProviderServiceServer) PowerOnMachine(_ context.Context, request *providerpb.PowerOnMachineRequest) (*providerpb.PowerOnMachineResponse, error) {
	// todo: implement
	p.logger.Info("powering on machine", zap.String("machine_id", request.Id))

	return &providerpb.PowerOnMachineResponse{}, nil
}
