package screens

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func makeCell() fyne.CanvasObject {
	rect := canvas.NewRectangle(&color.RGBA{128, 128, 128, 255})
	rect.SetMinSize(fyne.NewSize(30, 30))
	return rect
}

func makeBorderLayout() *fyne.Container {
	top := makeCell()
	bottom := makeCell()
	left := makeCell()
	right := makeCell()
	middle := widget.NewLabelWithStyle("BorderLayout", fyne.TextAlignCenter, fyne.TextStyle{})

	return container.NewBorder(top, bottom, left, right, middle)
}

func makeBoxLayout() *fyne.Container {
	top := makeCell()
	bottom := makeCell()
	middle := widget.NewLabel("BoxLayout")
	center := makeCell()
	right := makeCell()

	col := container.NewVBox(top, middle, bottom)

	return container.NewHBox(col, center, right)
}

func makeFixedGridLayout() *fyne.Container {
	box1 := makeCell()
	box2 := widget.NewLabel("FixedGrid")
	box3 := makeCell()
	box4 := makeCell()

	return container.New(layout.NewGridWrapLayout(fyne.NewSize(75, 75)),
		box1, box2, box3, box4)
}

func makeGridLayout() *fyne.Container {
	box1 := makeCell()
	box2 := widget.NewLabel("Grid")
	box3 := makeCell()
	box4 := makeCell()

	return container.NewGridWithColumns(2,
		box1, box2, box3, box4)
}

// LayoutPanel loads a panel that shows the layouts available for a container
func LayoutPanel() fyne.CanvasObject {
	return container.NewAppTabs(
		container.NewTabItem("Border", makeBorderLayout()),
		container.NewTabItem("Box", makeBoxLayout()),
		container.NewTabItem("Fixed Grid", makeFixedGridLayout()),
		container.NewTabItem("Grid", makeGridLayout()),
	)
}
