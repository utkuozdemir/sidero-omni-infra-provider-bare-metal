// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package main implements the main entrypoint for the Omni bare metal infra provider.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/ip"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/meta"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/provider"
	"github.com/siderolabs/omni-infra-provider-bare-metal/internal/version"
)

const apiHostFlag = "api-host"

var rootCmdArgs struct {
	apiHost             string
	omniAPIEndpoint     string
	imageFactoryPXEURL  string
	providerName        string
	providerDescription string

	ipxeServerPort int
	apiPort        int

	insecureSkipTLSVerify bool
	debug                 bool
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     version.Name,
	Short:   "Run the Omni bare metal infra provider",
	Version: version.Tag,
	Args:    cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		cmd.SilenceUsage = true // if the args are parsed fine, no need to show usage
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		logger, err := initLogger()
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		defer logger.Sync() //nolint:errcheck

		return run(cmd.Context(), logger)
	},
}

func initLogger() (*zap.Logger, error) {
	var loggerConfig zap.Config

	if rootCmdArgs.debug {
		loggerConfig = zap.NewDevelopmentConfig()
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		loggerConfig.Level.SetLevel(zap.DebugLevel)
	} else {
		loggerConfig = zap.NewProductionConfig()
		loggerConfig.Level.SetLevel(zap.InfoLevel)
	}

	return loggerConfig.Build()
}

func run(ctx context.Context, logger *zap.Logger) error {
	apiHost := rootCmdArgs.apiHost

	if apiHost == "" {
		routableIPs, err := ip.RoutableIPs()
		if err != nil {
			return fmt.Errorf("failed to get routable IPs: %w", err)
		}

		if len(routableIPs) != 1 {
			return fmt.Errorf(`expected exactly one routable IP, got %d: %v. specify "--%s" flag explicitly`, len(routableIPs), routableIPs, apiHostFlag)
		}

		apiHost = routableIPs[0]
	}

	logger.Info("starting server", zap.String("api_host", apiHost), zap.Int("port", rootCmdArgs.apiPort))

	prov := provider.New(
		rootCmdArgs.providerName, rootCmdArgs.providerDescription, rootCmdArgs.omniAPIEndpoint, rootCmdArgs.imageFactoryPXEURL, rootCmdArgs.ipxeServerPort,
		apiHost, rootCmdArgs.apiPort, rootCmdArgs.insecureSkipTLSVerify, logger)

	if err := prov.Run(ctx); err != nil {
		return fmt.Errorf("failed to run provider: %w", err)
	}

	return nil
}

func main() {
	if err := runCmd(); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}

func runCmd() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.Flags().StringVar(&rootCmdArgs.apiHost, apiHostFlag, "",
		"The IP address to bind on and advertise. Required if the server has more than a single routable IP address. If not specified, the single routable IP address will be used.")
	rootCmd.Flags().IntVar(&rootCmdArgs.apiPort, "api-port", 50042, "The port to run the api server on.")
	rootCmd.Flags().StringVar(&rootCmdArgs.omniAPIEndpoint, "omni-api-endpoint", os.Getenv("OMNI_ENDPOINT"),
		"The endpoint of the Omni API, if not set, defaults to OMNI_ENDPOINT env var.")
	rootCmd.Flags().StringVar(&meta.ProviderID, "id", meta.ProviderID, "The id of the infra provider, it is used to match the resources with the infra provider label.")
	rootCmd.Flags().StringVar(&rootCmdArgs.imageFactoryPXEURL, "image-factory-pxe-url", "https://pxe.factory.talos.dev", "The URL of the image factory PXE server.")
	rootCmd.Flags().IntVar(&rootCmdArgs.ipxeServerPort, "ipxe-server-port", 50043, "The port the local (chaining) iPXE server should run on.")

	rootCmd.Flags().StringVar(&rootCmdArgs.providerName, "provider-name", "Bare Metal", "Provider name as it appears in Omni")
	rootCmd.Flags().StringVar(&rootCmdArgs.providerDescription, "provider-description", "Bare metal infrastructure provider", "Provider description as it appears in Omni")

	// todo: add labels to be set to the machine

	rootCmd.Flags().BoolVar(&rootCmdArgs.insecureSkipTLSVerify, "insecure-skip-tls-verify", false, "Skip TLS verification when connecting to the Omni API.")
	rootCmd.Flags().BoolVar(&rootCmdArgs.debug, "debug", false, "Enable debug mode & logs.")
}
