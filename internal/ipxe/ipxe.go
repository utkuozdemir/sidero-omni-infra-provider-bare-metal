// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package ipxe provides iPXE functionality.
package ipxe

import (
	"net/http"

	"go.uber.org/zap"
)

// Handler represents an iPXE handler.
type Handler struct {
	logger *zap.Logger
}

// ServeHTTP serves the iPXE request.
//
// URL pattern: http://ip-of-this-provider:50042/ipxe?uuid=${uuid}&mac=${net${idx}/mac:hexhyp}&domain=${domain}&hostname=${hostname}&serial=${serial}&arch=${buildarch}
//
// Implements http.Handler interface.
func (s *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	uuid := req.URL.Query().Get("uuid")
	mac := req.URL.Query().Get("mac")
	domain := req.URL.Query().Get("domain")
	hostname := req.URL.Query().Get("hostname")
	serial := req.URL.Query().Get("serial")
	arch := req.URL.Query().Get("arch")

	s.logger.Info("handle iPXE request", zap.String("uuid", uuid), zap.String("mac", mac),
		zap.String("domain", domain), zap.String("hostname", hostname), zap.String("serial", serial), zap.String("arch", arch))

	// TODO implement me: boot into the agent mode + partial machine config (talos.config kernel arg) to join Omni SideroLink
	// Here we need 2 modes:
	// a. boot via chaining to the factory - build the schematic once and use it for all machines
	// b. boot by providing kernel and initramfs from this server for the agent development

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// NewHandler creates a new iPXE server.
func NewHandler(endpoint string, port int, logger *zap.Logger) (*Handler, error) {
	logger.Info("patch iPXE binaries")

	if err := patchBinaries(endpoint, port); err != nil {
		return nil, err
	}

	logger.Info("successfully patched iPXE binaries")

	return &Handler{
		logger: logger,
	}, nil
}
