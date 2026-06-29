//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package logger provides structured logging for the application.
package logger

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

// SetupLogger configures the global logger with the given log level.
// Valid levels are: debug, info, warn, error, fatal.
func SetupLogger(level string) error {
	// Parse log level
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Configure charmbracelet/log
	log.SetLevel(logLevel)
	if logLevel == log.DebugLevel {
		log.SetReportCaller(true)
	}
	log.SetReportTimestamp(true)
	log.SetTimeFormat(time.DateTime)

	return nil
}
