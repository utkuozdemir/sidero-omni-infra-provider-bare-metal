// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package ipxe provider iPXE functionality.
package ipxe

import (
	"context"

	"go.uber.org/zap"
)

// Server represents an iPXE server.
type Server struct {
	logger   *zap.Logger
	endpoint string
	port     int
}

// NewServer creates a new iPXE server.
func NewServer(endpoint string, port int, logger *zap.Logger) *Server {
	return &Server{
		endpoint: endpoint,
		port:     port,
		logger:   logger,
	}
}

// Run runs the iPXE server.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("patch iPXE binaries")

	if err := patchBinaries(s.endpoint, s.port); err != nil {
		return err
	}

	s.logger.Info("successfully patched iPXE binaries")

	// todo: here

	<-ctx.Done()

	return nil
}
