//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/dcvix/dcvix-launcher/internal/gui/components"
)

type guiNotifier struct {
	mainWindow  fyne.Window
	statusLabel *widget.Label
	serverList  *components.ServerList
	setEnabled  func(bool)
}

func (n *guiNotifier) SetButtonsEnabled(enabled bool) {
	fyne.Do(func() { n.setEnabled(enabled) })
}

func (n *guiNotifier) SetStatus(s string) {
	fyne.Do(func() { n.statusLabel.SetText(s) })
}

func (n *guiNotifier) ShowError(err error) {
	fyne.Do(func() { dialog.ShowError(err, n.mainWindow) })
}

func (n *guiNotifier) ClearServers() {
	if n.serverList != nil {
		n.serverList.ClearServers()
	}
}

func (n *guiNotifier) AddServer(s string) {
	if n.serverList != nil {
		n.serverList.AddServer(s)
	}
}

func (n *guiNotifier) ShowCustomEntry() {
	if n.serverList != nil {
		n.serverList.ShowCustomEntry()
	}
}
