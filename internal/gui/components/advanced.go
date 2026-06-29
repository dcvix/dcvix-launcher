//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/dcvix/dcvix-launcher/internal/config"
	"github.com/dcvix/dcvix-launcher/internal/service"
)

type AdvancedOptions struct {
	dialog                     dialog.Dialog
	qualitySelect              *widget.Select
	fullscreenCheck            *widget.Check
	useAllMonitorsCheck        *widget.Check
	enableDatagramDisplayCheck *widget.Check
	scalingModeCheck           *widget.Check
	certValidationCombo        *widget.Select
	customOptionsEntry         *widget.Entry
	dcvBinaryEntry             *widget.Entry
}

func NewAdvancedOptions(prefs fyne.Preferences, cfg config.Launcher) *AdvancedOptions {
	a := &AdvancedOptions{}

	a.qualitySelect = widget.NewSelect([]string{"maximum", "high", "medium", "low", "minimum"}, nil)
	a.qualitySelect.SetSelected(prefs.StringWithFallback("quality", "medium"))

	a.fullscreenCheck = widget.NewCheck("Start full screen", nil)
	a.fullscreenCheck.SetChecked(prefs.BoolWithFallback("fullScreen", false))

	a.useAllMonitorsCheck = widget.NewCheck("Use all monitors", nil)
	a.useAllMonitorsCheck.SetChecked(prefs.BoolWithFallback("useAllMonitors", false))

	a.enableDatagramDisplayCheck = widget.NewCheck("Enable datagram display", nil)
	a.enableDatagramDisplayCheck.SetChecked(prefs.BoolWithFallback("enableDatagramDisplay", false))

	a.scalingModeCheck = widget.NewCheck("Enable scaling mode (Windows only)", nil)
	a.scalingModeCheck.SetChecked(prefs.BoolWithFallback("scalingMode", false))

	a.certValidationCombo = widget.NewSelect([]string{"strict", "accept-untrusted", "ask-user"}, nil)
	a.certValidationCombo.SetSelected(prefs.StringWithFallback("certValidationPolicy", "accept-untrusted"))

	a.customOptionsEntry = widget.NewEntry()
	a.customOptionsEntry.SetText(prefs.StringWithFallback("customOptions", ""))

	a.dcvBinaryEntry = widget.NewEntry()
	a.dcvBinaryEntry.SetText(prefs.StringWithFallback("dcvBinary", cfg.Command))

	return a
}

func (a *AdvancedOptions) Show(parent fyne.Window) {
	if a.dialog == nil {
		content := container.NewVBox(
			widget.NewLabel("Quality:"),
			a.qualitySelect,
			a.fullscreenCheck,
			a.useAllMonitorsCheck,
			a.enableDatagramDisplayCheck,
			a.scalingModeCheck,
			widget.NewLabel("Certificate validation policy:"),
			a.certValidationCombo,
			widget.NewLabel("Custom options:"),
			a.customOptionsEntry,
			widget.NewLabel("DCV binary:"),
			a.dcvBinaryEntry,
		)
		scroll := container.NewVScroll(content)
		scroll.SetMinSize(fyne.NewSize(450, 550))
		a.dialog = dialog.NewCustom("Advanced Options", "Close", scroll, parent)
	}
	a.dialog.Show()
}

func (a *AdvancedOptions) SavePrefs(prefs fyne.Preferences) {
	prefs.SetString("quality", a.qualitySelect.Selected)
	prefs.SetBool("fullScreen", a.fullscreenCheck.Checked)
	prefs.SetBool("useAllMonitors", a.useAllMonitorsCheck.Checked)
	prefs.SetBool("enableDatagramDisplay", a.enableDatagramDisplayCheck.Checked)
	prefs.SetBool("scalingMode", a.scalingModeCheck.Checked)
	prefs.SetString("certValidationPolicy", a.certValidationCombo.Selected)
	prefs.SetString("customOptions", a.customOptionsEntry.Text)
	prefs.SetString("dcvBinary", a.dcvBinaryEntry.Text)
}

func (a *AdvancedOptions) GetData() service.DcvOptions {
	return service.DcvOptions{
		Quality:               a.qualitySelect.Selected,
		FullScreen:            a.fullscreenCheck.Checked,
		UseAllMonitors:        a.useAllMonitorsCheck.Checked,
		EnableDatagramDisplay: a.enableDatagramDisplayCheck.Checked,
		ScalingMode:           a.scalingModeCheck.Checked,
		CertValidationPolicy:  a.certValidationCombo.Selected,
		CustomOptions:         a.customOptionsEntry.Text,
		DcvBinary:             a.dcvBinaryEntry.Text,
	}
}
