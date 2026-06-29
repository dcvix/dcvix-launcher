//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ServerList struct {
	Component              fyne.CanvasObject
	serverListData         []string
	OnConnect              func(serverName string, connType string, b *widget.Button)
	CustomServerEntry      *widget.Entry
	CustomServerConnectBtn *widget.Button
	list                   *widget.List
	customBoxContainer     *fyne.Container
}

// NewServerList creates a scrollable list of available servers with connect buttons.
func NewServerList(prefs fyne.Preferences, allowCustomServer bool, onConnect func(string, string, *widget.Button)) *ServerList {
	s := &ServerList{
		OnConnect: onConnect,
	}

	s.list = widget.NewList(

		// func length
		func() int {
			return len(s.serverListData)
		},

		// func createItem
		func() fyne.CanvasObject {
			return container.NewPadded(
				container.NewBorder(
					nil, nil,
					// Left
					widget.NewLabel(""),
					// Right
					container.NewHBox(
						widget.NewButton("", nil),
						widget.NewButton("", nil),
					),
				),
			)
		},

		// func updateItem
		func(i widget.ListItemID, o fyne.CanvasObject) {
			pctn, _ := o.(*fyne.Container)
			ctn, _ := pctn.Objects[0].(*fyne.Container)

			l := ctn.Objects[0].(*widget.Label)
			l.SetText(s.serverListData[i])

			// Button box
			bbox := ctn.Objects[1].(*fyne.Container)

			// TODO: For now we only support Console sessions.
			b := bbox.Objects[0].(*widget.Button)
			b.OnTapped = func() {
				s.OnConnect(s.serverListData[i], "Virtual", b)
			}
			b.SetText("Virtual")
			b.Hide()

			b2 := bbox.Objects[1].(*widget.Button)
			b2.OnTapped = func() {
				s.OnConnect(s.serverListData[i], "Console", b2)
			}
			b2.SetText("Connect")
		},
	)

	s.CustomServerEntry = widget.NewEntry()
	s.CustomServerEntry.SetPlaceHolder("server-address:port")
	s.CustomServerEntry.Text = prefs.StringWithFallback("customServer", "")

	s.CustomServerConnectBtn = widget.NewButton("Connect", func() {
		addr := s.CustomServerEntry.Text
		if addr != "" {
			s.OnConnect(addr, "Console", s.CustomServerConnectBtn)
		}
	})

	customBox := container.NewBorder(
		nil, nil,
		widget.NewLabel("Custom Server:"),
		s.CustomServerConnectBtn,
		s.CustomServerEntry,
	)
	s.customBoxContainer = container.NewVBox(customBox)
	if !allowCustomServer {
		s.customBoxContainer.Hide()
	}

	s.Component = container.NewBorder(
		s.customBoxContainer, nil, nil, nil,
		s.list,
	)

	return s
}

// ClearServers removes all servers from the list and refreshes the UI.
func (sl *ServerList) ClearServers() {
	fyne.Do(func() {
		sl.serverListData = nil
		sl.list.Refresh()
	})
}

// AddServer appends a server name to the list and refreshes the UI.
func (sl *ServerList) AddServer(serverName string) {
	fyne.Do(func() {
		sl.serverListData = append(sl.serverListData, serverName)
		sl.list.Refresh()
	})
}

// ShowCustomEntry makes the custom server entry row visible.
func (sl *ServerList) ShowCustomEntry() {
	if sl.customBoxContainer != nil {
		fyne.Do(func() { sl.customBoxContainer.Show() })
	}
}

// SavePrefs saves the custom server address to preferences.
func (sl *ServerList) SavePrefs(prefs fyne.Preferences) {
	if sl.CustomServerEntry != nil {
		prefs.SetString("customServer", sl.CustomServerEntry.Text)
	}
}
