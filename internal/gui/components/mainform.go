//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package components

import (
	"fmt"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/dcvix/dcvix-launcher/internal/config"
)

type MainForm struct {
	Component     *fyne.Container
	BrokerEntry   *widget.Entry
	UserIDEntry   *widget.Entry
	PasswordEntry *widget.Entry
	OtpEntry      *widget.Entry
}

// NewMainForm creates the login form with broker, username, password, and optional OTP fields.
func NewMainForm(prefs fyne.Preferences, cfg config.Launcher) *MainForm {
	f := &MainForm{}

	brokerLabel := widget.NewLabel("Broker Server:")
	f.BrokerEntry = widget.NewEntry()
	f.BrokerEntry.SetText(prefs.StringWithFallback("broker", cfg.Broker))
	if cfg.Broker != "" {
		brokerLabel.Hide()
		f.BrokerEntry.Hide()
	}
	f.UserIDEntry = widget.NewEntry()
	f.UserIDEntry.SetText(prefs.StringWithFallback("userID", ""))

	f.PasswordEntry = widget.NewPasswordEntry()

	otpLabel := widget.NewLabel("OTP:")
	f.OtpEntry = widget.NewEntry()
	f.OtpEntry.SetPlaceHolder("Enter OTP")
	otpRegex := regexp.MustCompile(`^\d{6}$`)
	f.OtpEntry.Validator = func(text string) error {
		if !otpRegex.MatchString(text) {
			return fmt.Errorf("invalid input, please enter six numbers")
		}
		return nil
	}
	if !cfg.OTP {
		otpLabel.Hide()
		f.OtpEntry.Hide()
	}

	f.Component = container.New(layout.NewFormLayout(),
		brokerLabel, f.BrokerEntry,
		widget.NewLabel("Username:"), f.UserIDEntry,
		widget.NewLabel("Password:"), f.PasswordEntry,
		otpLabel, f.OtpEntry,
	)

	return f
}

// SavePrefs saves the current form field values to the application preferences.
func (a *MainForm) SavePrefs(prefs fyne.Preferences) {
	prefs.SetString("broker", a.BrokerEntry.Text)
	prefs.SetString("userID", a.UserIDEntry.Text)
}
