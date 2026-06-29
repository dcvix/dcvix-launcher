//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package icon

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Icon.png
var resourceIconPngData []byte

var ResourceIconPng = &fyne.StaticResource{
	StaticName:    "Icon.png",
	StaticContent: resourceIconPngData,
}
