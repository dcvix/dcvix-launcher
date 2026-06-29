//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dcvix/dcvix-launcher/internal/config"
	"github.com/dcvix/dcvix-launcher/internal/icon"
)

type LogoAndTitle struct {
	Component *fyne.Container
}

// NewLogoAndTitle creates the application logo and title component.
func NewLogoAndTitle() *LogoAndTitle {
	t := &LogoAndTitle{}

	// App Logo
	logoImage := canvas.NewImageFromResource(icon.ResourceIconPng)
	logoImage.FillMode = canvas.ImageFillContain
	logoImage.SetMinSize(fyne.NewSize(100, 100))

	// Title widget
	titleLabel := widget.NewLabel(config.AppName)

	// Horizontal layout for logo and title
	t.Component = container.NewVBox(
		container.NewCenter(logoImage),
		container.NewCenter(titleLabel),
	)

	return t
}
