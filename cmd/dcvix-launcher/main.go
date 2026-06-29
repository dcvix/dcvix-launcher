//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/charmbracelet/log"

	"github.com/dcvix/dcvix-launcher/internal/config"
	"github.com/dcvix/dcvix-launcher/internal/gui"
	"github.com/dcvix/dcvix-launcher/internal/logger"
	"github.com/dcvix/dcvix-launcher/internal/version"
)

func main() {
	var configPath string

	showVersion := flag.Bool("version", false, "Show version information")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.StringVar(&configPath, "config", "", "Path to the config")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\n", version.String())
		os.Exit(0)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		cfg.Log.Level = "debug"
	}

	// Setup logger
	if err := logger.SetupLogger(cfg.Log.Level); err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}

	a := app.NewWithID("net.cortassa.dcvix-launcher")
	gui.NewMainWindow(a, cfg.Launcher)
	a.Run()
}
