package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func ShowError(err error, parent fyne.Window) {
	dialog.ShowError(err, parent)
}
