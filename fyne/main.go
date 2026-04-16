// Package main provides various examples of Fyne API capabilities
package main

import (
	"flag"
	"fmt"
	"net/url"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/donomii/gitremind"
	"github.com/donomii/gitremind/fyne/screens"
)

const preferenceCurrentTab = "currentTab"

var verbose bool
var autoSync bool
var gui bool
var scanDir string = "."

func biggerButton(label string, tapped func()) fyne.CanvasObject {
	b := widget.NewButton(label, tapped)
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(0, 60))
	return container.NewStack(rect, container.NewPadded(b))
}

func welcomeScreen(a fyne.App) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(theme.FyneLogo())
	logo.SetMinSize(fyne.NewSize(800, 600))

	link, err := url.Parse("https://fyne.io/")
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("Welcome to the Fyne toolkit demo app", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), logo, layout.NewSpacer()),
		widget.NewHyperlinkWithStyle("fyne.io", link, fyne.TextAlignCenter, fyne.TextStyle{}),
		layout.NewSpacer(),

		widget.NewCard("Theme", "",
			container.NewGridWithColumns(2,
				biggerButton("Dark", func() {
					a.Settings().SetTheme(theme.DarkTheme())
				}),
				biggerButton("Light", func() {
					a.Settings().SetTheme(theme.LightTheme())
				}),
			),
		),
	)
}

func main() {

	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&gui, "gui", false, "Experimental graphical user interface")
	flag.BoolVar(&verbose, "verbose", false, "Print details while working")
	flag.Parse()
	if len(flag.Args()) > 0 {
		scanDir = flag.Arg(0)
	}

	//repos := doScan()
	g := gitremind.NewGitRemind()
	g.Scan(scanDir, verbose, autoSync)

	if gui {
		doGui(g)
	}

}

func doGui(g *gitremind.GitRemind) {
	a := app.NewWithID("com.praeceptamachinae.com")
	a.SetIcon(theme.FyneLogo())

	w := a.NewWindow("Git Remind")
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() { fmt.Println("Menu New") }),
		// a quit item will be appended to our first menu
	), fyne.NewMenu("Edit",
		fyne.NewMenuItem("Cut", func() { fmt.Println("Menu Cut") }),
		fyne.NewMenuItem("Copy", func() { fmt.Println("Menu Copy") }),
		fyne.NewMenuItem("Paste", func() { fmt.Println("Menu Paste") }),
	)))
	w.SetMaster()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Repos", theme.ViewFullScreenIcon(), screens.DialogScreen(w, a, g)))
	tabs.SetTabLocation(container.TabLocationLeading)
	tabs.SelectTabIndex(a.Preferences().Int(preferenceCurrentTab))
	w.SetContent(tabs)
	w.Resize(fyne.NewSize(800, 600))

	w.ShowAndRun()
	a.Preferences().SetInt(preferenceCurrentTab, tabs.CurrentTabIndex())
}
