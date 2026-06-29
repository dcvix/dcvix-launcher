//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package service orchestrates the business logic, tying together configuration,
// client requests, GUI events, and system operations.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/shlex"

	"github.com/dcvix/dcvix-launcher/internal/client"
)

// UINotifier abstracts UI interactions from the service layer.
type UINotifier interface {
	SetButtonsEnabled(bool)
	SetStatus(string)
	ShowError(error)
	ClearServers()
	AddServer(string)
	ShowCustomEntry()
}

// DcvOptions holds DCV connection parameters.
type DcvOptions struct {
	Quality               string
	FullScreen            bool
	UseAllMonitors        bool
	EnableDatagramDisplay bool
	ScalingMode           bool
	CertValidationPolicy  string
	CustomOptions         string
	DcvBinary             string
}

// Login orchestrates the login flow
func Login(api *client.APIClient, userID, password, otp string, n UINotifier) {
	n.SetStatus("Logging in...")

	err := api.Login(userID, password, otp)
	if err != nil {
		n.ShowError(err)
		n.SetButtonsEnabled(true)
		n.SetStatus("")
		return
	}

	servers, err := api.ListServers()
	if err != nil {
		n.ShowError(err)
		n.SetButtonsEnabled(true)
		n.SetStatus("")
		return
	}

	sort.Strings(servers)

	n.ClearServers()
	for _, v := range servers {
		log.Debugf("server: %s", v)
		if v == "ALLOW_CUSTOM" {
			log.Debug("Showing custom entry for ALLOW_CUSTOM")
			n.ShowCustomEntry()
			continue
		}
		n.AddServer(v)
	}
	n.SetStatus("Logged in")
}

// Connect orchestrates the connection flow
func Connect(api *client.APIClient, server, userID, SessionType string, opts DcvOptions, n UINotifier) {
	var sessionID string
	var err error

	n.SetButtonsEnabled(false)
	n.SetStatus("Creating session...")

	var qualityValue string
	switch opts.Quality {
	case "maximum":
		qualityValue = "(90, 100)"
	case "high":
		qualityValue = "(80, 90)"
	case "medium":
		qualityValue = "(30, 80)"
	case "low":
		qualityValue = "(10, 30)"
	case "minimum":
		qualityValue = "(0, 10)"
	default:
		log.Warnf("Invalid quality setting: %s, using medium", opts.Quality)
		qualityValue = "(30, 80)"
	}

	log.Debugf("Setting quality config for server %s", server)
	err = api.SetConfig(server, []client.ConfigEntry{
		{Section: "display", Key: "quality", Value: qualityValue},
	})
	if err != nil {
		n.ShowError(err)
	}

	sessionID, err = api.CreateSession(server, userID, SessionType)
	if err != nil {
		n.ShowError(err)
		n.SetButtonsEnabled(true)
		return
	}
	log.Debugf("Got session id: %s", sessionID)

	err = api.GetConnectionToken(server, userID, sessionID)
	if err != nil {
		n.ShowError(err)
		n.SetButtonsEnabled(true)
		return
	}
	log.Debugf("Got connection Token: %s", api.ConnectionToken())

	n.SetStatus("Session created...")

	content := fmt.Sprintf(`
[version]
format=1.0

[connect]
host=%s
port=8443
sessionid=%s
user=%s
weburlpath=
authtoken=%s
certificatevalidationpolicy=%s

[options]
promptreconnect=false
fullscreen=%t
useallmonitors=%t
`, server, sessionID, userID, api.ConnectionToken(), opts.CertValidationPolicy, opts.FullScreen, opts.UseAllMonitors)

	dcvFile, err := os.CreateTemp("", "dcvix-*.dcv")
	if err != nil {
		n.ShowError(fmt.Errorf("failed to create temporary file: %w", err))
		n.SetButtonsEnabled(true)
		return
	}
	if _, err := dcvFile.WriteString(content); err != nil {
		dcvFile.Close()
		os.Remove(dcvFile.Name())
		n.ShowError(fmt.Errorf("failed to write connection file: %w", err))
		n.SetButtonsEnabled(true)
		return
	}
	dcvFile.Close()

	log.Debugf("Connection file created: %s", dcvFile.Name())

	// Build command args
	args := []string{dcvFile.Name()}
	if opts.EnableDatagramDisplay {
		args = append(args, "--enable-datagrams-display=true")
	}
	if opts.ScalingMode && runtime.GOOS == "windows" {
		args = append(args, "--scaling-mode", "scaling")
	}
	customOptions, err := shlex.Split(opts.CustomOptions)
	if err != nil {
		log.Errorf("failed to parse custom options: %s", err)
		n.ShowError(fmt.Errorf("invalid custom options: %w", err))
		n.SetButtonsEnabled(true)
		return
	}
	args = append(args, customOptions...)

	log.Debugf("Running command: %s %v", opts.DcvBinary, args)
	cmd := exec.Command(opts.DcvBinary, args...)

	if err := cmd.Start(); err != nil {
		os.Remove(dcvFile.Name())
		n.ShowError(err)
		n.SetButtonsEnabled(true)
		return
	}

	n.SetStatus("Connected to " + server)

	// delete connection file after 3 seconds (dcvviewer already read the file)
	go func() {
		time.Sleep(3 * time.Second)
		os.Remove(dcvFile.Name())
	}()

	// wait for dcv to close
	go func() {
		cmd.Wait()
		n.SetButtonsEnabled(true)
	}()
}
