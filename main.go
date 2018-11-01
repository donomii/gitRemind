package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/donomii/goof"
	"github.com/rivo/tview"
)

var autoSync bool
var ui bool
var verbose bool
var repos [][]string
var lastSelect string
var app *tview.Application
var workerChan chan string
var doneChan chan bool

func worker(c chan string) {
	var ahead_regex = regexp.MustCompile(`Your branch is ahead of`)
	var not_staged_regex = regexp.MustCompile(`Changes not staged for commit:`)
	var staged_not_committed_regex = regexp.MustCompile(`Changes to be committed`)
	var modified_regex = regexp.MustCompile(`modified:`)
	var untracked_regex = regexp.MustCompile(`Untracked files:`)

	repos = [][]string{}
	cwd, _ := os.Getwd()
	for path := range c {

		os.Chdir(cwd)
		gitpath := fmt.Sprintf("%v/%v", path, ".git")
		if goof.IsDir(gitpath) {
			if verbose {
				log.Println(gitpath)
			}
			os.Chdir(path)
			cmd := exec.Command("git", "status")
			result := goof.QuickCommand(cmd)
			cmd = exec.Command("git", "status", "--porcelain")
			shortresult := goof.QuickCommand(cmd)
			cmd = exec.Command("git", "diff", "--ignore-blank-lines")
			diffresult := goof.QuickCommand(cmd)
			reasons := []string{}
			longreasons := []string{}

			if ahead_regex.MatchString(result) {
				reasons = append(reasons, "push")
				longreasons = append(longreasons, "local commits not pushed")
			}
			if modified_regex.MatchString(result) || not_staged_regex.MatchString(result) || staged_not_committed_regex.MatchString(result) {
				reasons = append(reasons, "commit")
				longreasons = append(longreasons, "changes not committed")
			}
			if untracked_regex.MatchString(result) {
				reasons = append(reasons, "track")
				longreasons = append(longreasons, "untracked files present")
			}
			if len(reasons) > 0 {
				fmt.Printf("%v: %v\n", path, strings.Join(longreasons, ", "))
				repos = append(repos, []string{path, shortresult, grep(diffresult), strings.Join(reasons, ", "), strings.Join(longreasons, ", ")})
				if verbose {
					fmt.Println(result)
					fmt.Printf("\n\n\n\n\n")
				}
			}

			if autoSync {
				fmt.Println("Syncing " + path)
				cmd := exec.Command("git", "push")
				goof.QuickCommand(cmd)
				cmd = exec.Command("git", "pull")
				goof.QuickCommand(cmd)
			}
		}
		os.Chdir(cwd)
	}
	doneChan <- true
}

func grep(str string) string {
	var out string
	strs := strings.Split(str, "\n")
	for _, v := range strs {
		if strings.Index(v, "+") == 0 || strings.Index(v, "-") == 0 {
			out = out + v + "\n"
		}
	}
	return out
}

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
				cwd, _ := os.Getwd()

				os.Chdir(v[0])
				fmt.Printf("%v\n", []string{"git", "commit", "-a"})
				goof.QCI([]string{"git", "commit", "-a"})
				goof.QCI([]string{"git", "push"})
				os.Chdir(cwd)
				doScan()
				app.Run()
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

func scanRepos(c chan string) {
	var git_regex = regexp.MustCompile(`\.git`)
	walkHandler := func(path string, info os.FileInfo, err error) error {

		if !git_regex.MatchString(path) {

			c <- path

		}
		return nil
	}
	//fmt.Println("These repositories need some attention:")
	filepath.Walk(".", walkHandler)
	close(c)
}

func doScan() {
	workerChan = make(chan string, 10)
	doneChan = make(chan bool)
	go worker(workerChan)
	scanRepos(workerChan)
	<-doneChan
	close(doneChan)
	log.Println("Scan complete!")
}

func main() {
	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.BoolVar(&verbose, "verbose", false, "Print details while working")
	flag.Parse()

	doScan()

	if ui {
		doui()
	}
	fmt.Println("Done!")
}
