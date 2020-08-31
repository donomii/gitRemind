package screens

import (
	"fmt"
	"strings"

	"fyne.io/fyne"
	//"fyne.io/fyne/dialog"
	"os"

	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/donomii/gitremind/textedit"
	"github.com/donomii/goof"
)

var commitMessage string
var targetDir string

func confirmCallback(response bool) {
	fmt.Println("Responded with", response)
}

// DialogScreen loads a panel that lists the dialog windows that can be tested.
func DialogScreen(win fyne.Window, a fyne.App, repos [][]string) fyne.CanvasObject {

	top := makeCell()
	bottom := makeCell()
	left := makeCell()
	right := makeCell()

	largeText := widget.NewLabel("")

	//form := &widget.Form{}
	//form.Append("Message", largeText)
	diffs := widget.NewGroupWithScroller("Status", largeText)
	middle := diffs

	borderLayout := layout.NewBorderLayout(top, bottom, left, right)

	pull_Button := widget.NewButton("Pull", func() {
		Pull(targetDir)
	})

	commit_Push_Button := widget.NewButton("Commit - Push", func() {
		editor := textedit.Show(a, targetDir)
		editor.SetText(commitMessage)
	})

	gitControls := widget.NewGroup("Git", pull_Button, commit_Push_Button)
	diffCon := fyne.NewContainerWithLayout(borderLayout,
		top, bottom, left, right, middle)
	//diffPanel := fyne.NewContainerWithLayout(diffLayout, borderLayout, gitControls)
	buttons := []fyne.CanvasObject{}
	for _, r := range repos {
		name := r[0]
		path := r[0]
		detailDisplay := "Problems\n-----\n" + r[4] + "\n\nFiles\n-----\n" + r[1] + "\nDiff\n----\n" + r[2]
		detailDisplay = strings.Replace(detailDisplay, "\t", "   ", -1)
		commitMessage = "\n#" + strings.Replace(r[5], "\n", "\n#", -1)
		b := widget.NewButton(name, func() {
			largeText.SetText(detailDisplay)
			targetDir = path
		})
		buttons = append(buttons, b)
	}

	//buttons = append(buttons, form)
	dialogs := widget.NewGroup("Repositories", buttons...)

	windows := widget.NewVBox(dialogs, gitControls)

	return fyne.NewContainerWithLayout(layout.NewAdaptiveGridLayout(2), windows, diffCon)
}

func CommitPush(targetDir string) {
	cwd, _ := os.Getwd()

	os.Chdir(targetDir)
	fmt.Printf("%v\n", []string{"git", "commit", "-a"})
	goof.QCI([]string{"git", "commit", "-a"})
	goof.QCI([]string{"git", "push"})
	os.Chdir(cwd)
}

func Pull(targetDir string) {
	cwd, _ := os.Getwd()

	os.Chdir(targetDir)
	goof.QCI([]string{"git", "pull"})
	os.Chdir(cwd)
}
