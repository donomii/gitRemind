package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/donomii/goof"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/rivo/tview"
	"github.com/xlab/closer"
	"golang.org/x/image/font/gofont/goregular"
)

var autoSync bool
var ui bool
var gui bool
var verbose bool
var repos [][]string
var lastSelect string
var app *tview.Application
var workerChan chan string
var doneChan chan bool

var winWidth = 900
var winHeight = 900

type Option uint8

type State struct {
	bgColor nk.Color
	prop    int32
	opt     Option
}

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
	flag.BoolVar(&gui, "gui", false, "Experimental graphical user interface")
	flag.BoolVar(&verbose, "verbose", false, "Print details while working")
	flag.Parse()

	doScan()

	if ui {
		doui()
	}
	if gui {
		startNuke()
	}
	fmt.Println("Done!")
}

// Start nuklear
func startNuke() {
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		closer.Fatalln(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	win, err := glfw.CreateWindow(winWidth, winHeight, "Menu", nil, nil)
	if err != nil {
		closer.Fatalln(err)
	}
	win.MakeContextCurrent()

	width, height := win.GetSize()
	log.Printf("glfw: created window %dx%d", width, height)

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}
	gl.Viewport(0, 0, int32(width-1), int32(height-1))

	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	/*data, err := ioutil.ReadFile("FreeSans.ttf")
	if err != nil {
		panic("Could not find file")
	}*/

	sansFont := nk.NkFontAtlasAddFromBytes(atlas, goregular.TTF, 16, nil)
	// sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctx, sansFont.Handle())
	}

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	fpsTicker := time.NewTicker(time.Second / 30)
	for {
		select {
		case <-exitC:
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case <-fpsTicker.C:
			if win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			state := &State{
				bgColor: nk.NkRgba(28, 48, 62, 255),
			}
			gfxMain(win, ctx, state)
		}
	}

	//End Nuklear
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {

	maxVertexBuffer := 512 * 1024
	maxElementBuffer := 128 * 1024

	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	update := nk.NkBegin(ctx, "Menu", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)
	nk.NkWindowSetPosition(ctx, "Menu", nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, "Menu", nk.NkVec2(float32(winWidth), float32(winHeight)))

	if update > 0 {

		QuickFileEditor(ctx)

	}
	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func ButtonBox(ctx *nk.Context) {

}

func QuickFileEditor(ctx *nk.Context) {

}
