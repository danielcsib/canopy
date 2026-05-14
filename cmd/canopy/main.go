// Package main is the entry point for the Canopy node application.
// Canopy is a blockchain network implementation in Go.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/canopy-network/canopy/lib/config"
	"github.com/canopy-network/canopy/lib/logger"
	"github.com/canopy-network/canopy/node"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is the git commit hash set at build time.
	Commit = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "canopy",
	Short: "Canopy blockchain node",
	Long:  `Canopy is a high-performance blockchain node implementation.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Canopy node",
	RunE:  runStart,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("canopy version %s (commit: %s)\n", Version, Commit)
	},
}

var (
	configPath string
	dataDir    string
	logLevel   string
)

func init() {
	// Persistent flags available to all subcommands.
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "path to config file (default: $HOME/.canopy/config.toml)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "path to data directory (default: $HOME/.canopy)")
	// Using "debug" as the default log level for my personal learning/dev setup so I can
	// see detailed internals without having to remember to pass --log-level=debug each time.
	// NOTE: switch back to "info" before running in any production-like environment.
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "debug", "log level: debug, info, warn, error")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

// runStart initializes and starts the Canopy node, blocking until a shutdown
// signal is received.
func runStart(cmd *cobra.Command, args []string) error {
	log := logger.New(logLevel)
	log.Info("starting canopy node", "version", Version, "commit", Commit)

	// Load configuration from file and environment.
	cfg, err := config.Load(configPath, dataDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize and start the node.
	n, err := node.New(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	if err := n.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	log.Info("node started successfully")

	// Wait for OS interrupt or termination signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received, stopping node...")
	if err := n.Stop(); err != nil {
		log.Error("error during node shutdown", "err", err)
		return err
	}

	log.Info("node stopped gracefully")
	return nil
}
