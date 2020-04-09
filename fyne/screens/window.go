package screens

import (
	"fmt"

	"fyne.io/fyne"
	//"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func confirmCallback(response bool) {
	fmt.Println("Responded with", response)
}

// DialogScreen loads a panel that lists the dialog windows that can be tested.
func DialogScreen(win fyne.Window, repos [][]string) fyne.CanvasObject {

	largeText := widget.NewMultiLineEntry()
	form := &widget.Form{
		OnCancel: func() {
			fmt.Println("Cancelled")
		},
		OnSubmit: func() {

			fmt.Println("Message:", largeText.Text)
		},
	}
	form.Append("Message", largeText)

	diffs := widget.NewGroup("Diff", form)
	diffPanel := widget.NewVBox(diffs)
	buttons := []fyne.CanvasObject{}
	for _, r := range repos {
		name := r[0]
		text := r[1]
		b := widget.NewButton(name, func() {
			largeText.SetText(text)
		})
		buttons = append(buttons, b)
	}

	//buttons = append(buttons, form)
	dialogs := widget.NewGroup("Dialogs", buttons...)

	windows := widget.NewVBox(dialogs)

	return fyne.NewContainerWithLayout(layout.NewAdaptiveGridLayout(2), windows, diffPanel)
}
