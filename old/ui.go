// ui.go
package main

import (

	"github.com/rivo/tview"
)
func doui() {

	//box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
	app = tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textView.SetText("lalalala")

	textView2 := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textView2.SetText("lalalala")

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	footer.SetText("lalalala")

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	list := tview.NewList()
	for i, vv := range repos {

		ii := i
		v := vv
		list.AddItem(v[0], v[3], 'a', func() {

			if lastSelect == v[0] {
				app.Stop()

				CommitPush(v[0])
				doScan()
				doui()
			}
			if len(repos) > ii { //FIXME???
				textView.SetText(repos[ii][2])
				textView2.SetText(repos[ii][1])
				footer.SetText(repos[ii][4])
				lastSelect = v[0]
			}
		})
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	//menu := newPrimitive("Menu")
	//sideBar := newPrimitive("Side Bar")

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
		AddItem(footer, 2, 0, 1, 3, 0, 0, false)

	/*
		        grid.AddItem(menu, 0, 0, 1, 3, 0, 0, false).
		        AddItem(list, 1, 0, 1, 3, 0, 0, true).
				AddItem(sideBar, 0, 0, 1, 3, 0, 0, false)
	*/

	grid.AddItem(list, 1, 0, 1, 1, 0, 100, true).
		AddItem(textView, 1, 1, 1, 1, 0, 100, false).
		AddItem(textView2, 1, 2, 1, 1, 0, 100, false)
	//left := flex.AddItem(tview.NewBox().SetBorder(true).SetTitle("Left (1/2 x width of Top)"), 0, 1, false)
	//row := tview.NewFlex().SetDirection(tview.FlexRow)
	//row = row.AddItem(list.SetBorder(true).SetTitle("Repos"), 0, 3, true)
	//row = row.AddItem(textView.SetBorder(true).SetTitle("Status"), 0, 3, false)
	//flex.AddItem(left.SetBorder(true), 10, 1, false)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
