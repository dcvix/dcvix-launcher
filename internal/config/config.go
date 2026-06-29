//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package config reads the application configuration from an INI file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/log"
	"gopkg.in/ini.v1"
)

const AppName = "dcvix Launcher"
const AppID = "net.cortassa.dcvix-launcher"
const AppDesc = "DCV viewer launcher."

type Config struct {
	Launcher Launcher
	Log      Log
}

type Launcher struct {
	Broker              string
	UserID              string
	OTP                 bool
	Command             string
	AcceptUntrustedCert bool
	AllowCustomServer   bool
}

type Log struct {
	Level string
}

// Load reads the INI configuration file and returns the parsed Config.
// If configPath is empty, it searches for a config file in the user config
// directory, the current directory, and the executable directory.
// If none is found, a default config is written to the user config directory.
func Load(configPath string) (*Config, error) {
	// Start with embedded defaults as the single source of truth.
	cfg, err := defaults()
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded defaults: %w", err)
	}

	if configPath != "" {
		// Explicit path provided: error if not found or unreadable
		return loadFromPath(cfg, configPath)
	}

	// Search config:
	// user config dir
	//     ~/.config/net.cortassa.dcvix-launcher/dcvix-launcher.conf on Linux
	//     ~/Library/Application Support/net.cortassa.dcvix-launcher/dcvix-launcher.conf on macOS
	//     %AppData%/net.cortassa.dcvix-launcher/dcvix-launcher.conf on Windows
	// current dir
	// executable dir
	configPath = findConfig()
	if configPath != "" {
		return loadFromPath(cfg, configPath)
	}

	// No config found — use embedded defaults and write them to disk.
	log.Warn("Config file not found, using default config")
	if userConfigDir, err := os.UserConfigDir(); err == nil {
		userConfig := filepath.Join(userConfigDir, AppID, "dcvix-launcher.conf")
		if err := os.MkdirAll(filepath.Dir(userConfig), 0755); err == nil {
			if err := os.WriteFile(userConfig, defaultConfig, 0644); err == nil {
				log.Infof("Wrote default config to %s", userConfig)
			}
		}
	}
	return cfg, nil
}

// defaults parses the embedded default config into a Config struct.
func defaults() (*Config, error) {
	file, err := ini.Load(defaultConfig)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	loadIni(file, cfg)
	return cfg, nil
}

// loadFromPath loads a config file from disk and overlays its values on cfg.
func loadFromPath(cfg *Config, path string) (*Config, error) {
	log.Debugf("Loading %s", path)
	file, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	loadIni(file, cfg)
	return cfg, nil
}

// findConfig searches for dcvix-launcher.conf in standard locations.
func findConfig() string {
	if userConfigDir, err := os.UserConfigDir(); err == nil {
		userConfig := filepath.Join(userConfigDir, AppID, "dcvix-launcher.conf")
		if _, err := os.Stat(userConfig); err == nil {
			return userConfig
		}
	}
	if _, err := os.Stat("dcvix-launcher.conf"); err == nil {
		return "dcvix-launcher.conf"
	}
	ex, err := os.Executable()
	if err == nil {
		execConfig := filepath.Join(filepath.Dir(ex), "dcvix-launcher.conf")
		if _, err := os.Stat(execConfig); err == nil {
			return execConfig
		}
	}
	return ""
}

func loadIni(file *ini.File, cfg *Config) {
	launcherSection := file.Section("dcvix-launcher")
	cfg.Launcher.Broker = launcherSection.Key("broker").MustString(cfg.Launcher.Broker)
	cfg.Launcher.UserID = launcherSection.Key("user_id").MustString(cfg.Launcher.UserID)
	cfg.Launcher.OTP = launcherSection.Key("otp").MustBool(cfg.Launcher.OTP)
	cfg.Launcher.Command = launcherSection.Key("command").MustString(cfg.Launcher.Command)
	cfg.Launcher.AcceptUntrustedCert = launcherSection.Key("accept-untrusted-cert").MustBool(cfg.Launcher.AcceptUntrustedCert)
	cfg.Launcher.AllowCustomServer = launcherSection.Key("allow-custom-server").MustBool(cfg.Launcher.AllowCustomServer)

	logSection := file.Section("log")
	cfg.Log.Level = logSection.Key("level").MustString(cfg.Log.Level)
}

// DefaultDcvPath returns the default path to the DCV viewer binary for the current platform.
func DefaultDcvPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Program Files (x86)\NICE\DCV\Client\bin\dcvviewer.exe`
	}
	return "dcvviewer"
}
