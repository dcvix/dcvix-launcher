//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package gui renders the application user interface using the Fyne toolkit.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/charmbracelet/log"
	"github.com/dcvix/dcvix-launcher/internal/client"
	"github.com/dcvix/dcvix-launcher/internal/config"
	"github.com/dcvix/dcvix-launcher/internal/gui/components"
	"github.com/dcvix/dcvix-launcher/internal/icon"
	"github.com/dcvix/dcvix-launcher/internal/service"
	"github.com/dcvix/dcvix-launcher/internal/version"
)

type headerComponents struct {
	logoAndTitle *components.LogoAndTitle
	mainForm     *components.MainForm
	advancedOpts *components.AdvancedOptions
}

func buildHeader(prefs fyne.Preferences, cfg config.Launcher) headerComponents {
	return headerComponents{
		logoAndTitle: components.NewLogoAndTitle(),
		mainForm:     components.NewMainForm(prefs, cfg),
		advancedOpts: components.NewAdvancedOptions(prefs, cfg),
	}
}

func buildBody(
	prefs fyne.Preferences,
	cfg config.Launcher,
	h headerComponents,
	w fyne.Window,
	apiClient **client.APIClient,
	statusLabel **widget.Label,
) *components.ServerList {
	return components.NewServerList(prefs, cfg.AllowCustomServer, func(serverName, connType string, b *widget.Button) {
		if *apiClient == nil {
			dialog.ShowInformation("Not logged in", "Please log in first before connecting.", w)
			return
		}
		go service.Connect(
			*apiClient,
			serverName,
			h.mainForm.UserIDEntry.Text,
			connType,
			h.advancedOpts.GetData(),
			&guiNotifier{
				mainWindow:  w,
				statusLabel: *statusLabel,
				setEnabled: func(enabled bool) {
					if enabled {
						b.Enable()
					} else {
						b.Disable()
					}
				},
			},
		)
	})
}

func buildTray(a fyne.App, w fyne.Window) {
	desk, ok := a.(desktop.App)
	if !ok {
		return
	}
	m := fyne.NewMenu("dcvix Launcher",
		fyne.NewMenuItem("Show", func() { w.Show() }),
		fyne.NewMenuItem("Quit", func() { a.Quit() }),
	)
	desk.SetSystemTrayMenu(m)
	desk.SetSystemTrayIcon(icon.ResourceIconPng)
}

func buildMenu(a fyne.App, w fyne.Window, advancedOpts *components.AdvancedOptions) {
	advancedItem := fyne.NewMenuItem("Advanced Options", func() {
		advancedOpts.Show(w)
	})
	quitItem := fyne.NewMenuItem("Quit", func() { a.Quit() })
	aboutItem := fyne.NewMenuItem("About", func() {
		dialog.ShowInformation("About "+config.AppName,
			config.AppName+"\n"+version.String()+"\n\n"+config.AppDesc, w)
	})
	hideItem := fyne.NewMenuItem("Hide", func() { w.Hide() })
	fileMenu := fyne.NewMenu("File", advancedItem, fyne.NewMenuItemSeparator(), hideItem, quitItem)
	helpMenu := fyne.NewMenu("Help", aboutItem)
	mainMenu := fyne.NewMainMenu(fileMenu, helpMenu)
	w.SetMainMenu(mainMenu)
}

func buildLayout(logoAndTitle *components.LogoAndTitle, mainForm *components.MainForm, loginButton *widget.Button, serverList *components.ServerList, statusLabel *widget.Label) fyne.CanvasObject {
	topVbox := container.NewVBox(
		container.NewCenter(logoAndTitle.Component),
		container.NewPadded(mainForm.Component),
		container.NewCenter(loginButton),
		widget.NewLabel("Available Workstations:"),
	)
	bottomVbox := container.NewVBox(
		container.NewCenter(statusLabel),
	)
	return container.NewBorder(topVbox, bottomVbox, nil, nil, serverList.Component)
}

// NewMainWindow creates and returns the main application window with all UI components.
func NewMainWindow(a fyne.App, cfg config.Launcher) {
	w := a.NewWindow(config.AppName)
	w.SetIcon(icon.ResourceIconPng)
	prefs := a.Preferences()

	h := buildHeader(prefs, cfg)

	var apiClient *client.APIClient
	var loginButton *widget.Button
	var statusLabel *widget.Label

	serverList := buildBody(prefs, cfg, h, w, &apiClient, &statusLabel)

	savePrefs := func() {
		log.Debug("Saving preferences...")
		h.mainForm.SavePrefs(prefs)
		h.advancedOpts.SavePrefs(prefs)
		serverList.SavePrefs(prefs)
	}

	loginAction := func() {
		broker := h.mainForm.BrokerEntry.Text
		userID := h.mainForm.UserIDEntry.Text
		password := h.mainForm.PasswordEntry.Text
		otp := h.mainForm.OtpEntry.Text

		savePrefs()

		var err error
		if apiClient != nil {
			apiClient.Close()
		}
		apiClient, err = client.NewAPIClient(broker, cfg.AcceptUntrustedCert)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		loginButton.Disable()
		go service.Login(apiClient, userID, password, otp, &guiNotifier{
			mainWindow:  w,
			statusLabel: statusLabel,
			serverList:  serverList,
			setEnabled: func(enabled bool) {
				if enabled {
					loginButton.Enable()
				} else {
					loginButton.Disable()
				}
			},
		})
	}

	loginButton = widget.NewButton("Login", loginAction)
	h.mainForm.PasswordEntry.OnSubmitted = func(_ string) { loginAction() }
	statusLabel = widget.NewLabelWithStyle("Not connected", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true})

	content := buildLayout(h.logoAndTitle, h.mainForm, loginButton, serverList, statusLabel)

	buildTray(a, w)
	buildMenu(a, w, h.advancedOpts)

	w.SetCloseIntercept(func() { w.Hide() })
	a.Lifecycle().SetOnStopped(savePrefs)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 600))
	w.Show()
}
