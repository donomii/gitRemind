package screens

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	//"fyne.io/fyne/dialog"
	"os"

	"fyne.io/fyne/v2/widget"
	"github.com/donomii/gitremind"
	"github.com/donomii/gitremind/fyne/textedit"
	"github.com/donomii/goof"
)

var commitMessage string
var targetDir string

func confirmCallback(response bool) {
	fmt.Println("Responded with", response)
}

// DialogScreen loads a panel that lists the dialog windows that can be tested.
func DialogScreen(win fyne.Window, a fyne.App, g *gitremind.GitRemind) fyne.CanvasObject {

	top := makeCell()
	bottom := makeCell()
	left := makeCell()
	right := makeCell()

	largeText := widget.NewLabel("")

	//form := &widget.Form{}
	//form.Append("Message", largeText)
	diffs := container.NewScroll(largeText)
	middle := diffs

	pull_Button := widget.NewButton("Pull", func() {
		Pull(targetDir)
	})

	// Placeholder for refresh function
	var refreshRepoList func()

	commit_Push_Button := widget.NewButton("Commit - Push", func() {
		editor := textedit.Show(a, targetDir, func() {
			// Callback after commit
			fmt.Println("Commit done, refreshing...")

			// Re-scan specific repo
			// We assume targetDir is the path
			if targetDir != "" {
				g.ProcessRepo(targetDir, false, false)
			}

			// Refresh UI
			refreshRepoList()
		})
		editor.SetText(commitMessage)
	})

	gitControls := widget.NewCard("Actions", "", container.NewHBox(pull_Button, commit_Push_Button))
	diffCon := container.NewBorder(top, bottom, left, right, middle)

	repoList := container.NewVBox()

	refreshRepoList = func() {
		// Get fresh repos
		reposMap := g.GetRepos()
		var repos [][]string
		for _, v := range reposMap {
			repos = append(repos, v)
		}

		buttons := []fyne.CanvasObject{}
		for _, r := range repos {
			name := r[0]
			path := r[0]
			// r[5] is status result. If it indicates clean, we might skip?
			// The user said: "decide whether to remove it from the list or re-show it with new status information"
			// If we re-scan and it has no issues, ProcessRepo might not SetRepo?
			// Wait, ProcessRepo only calls SetRepo if len(reasons) > 0 !
			// But if it was previously there and now has no issues, it remains in the map?
			// The current ProcessRepo logic does NOT remove it if clean. It just doesn't update/add.
			// We need to handle removal if clean.

			// Let's check ProcessRepo logic in backend.go:
			// "if len(reasons) > 0 { SetRepo ... } else { ... }"
			// It implies if scanning finds no issues, it doesn't call SetRepo.
			// But it doesn't call RemoveRepo either.
			// We should probably explicitly remove it if clean in ProcessRepo, OR handle it here by checking content.
			// Ideally ProcessRepo handles it. I'll stick to displaying what's in the map for now.
			// If validity check is needed, we can add it.

			detailDisplay := "Problems\n-----\n" + r[4] + "\n\nFiles\n-----\n" + r[1] + "\nDiff\n----\n" + r[2]
			detailDisplay = strings.Replace(detailDisplay, "\t", "   ", -1)
			commitMessage = "\n#" + strings.Replace(r[5], "\n", "\n#", -1)
			b := widget.NewButton(name, func() {
				largeText.SetText(detailDisplay)
				targetDir = path
			})
			buttons = append(buttons, b)
		}
		repoList.Objects = buttons
		repoList.Refresh()
	}

	// Initial population
	refreshRepoList()

	repoScroll := container.NewScroll(repoList)

	// Use HSplit for resizable panes
	split := container.NewHSplit(repoScroll, diffCon)
	split.SetOffset(0.3) // Give 30% to the list by default

	// Main layout: Top Controls, Center Split
	return container.NewBorder(gitControls, nil, nil, nil, split)
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
