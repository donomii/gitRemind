// gui.go
package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"
	"golang.org/x/image/font/gofont/goregular"
)

// Start nuklear
func startNuke() {

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
	update := nk.NkBegin(ctx, "GitRemind", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)
	nk.NkWindowSetPosition(ctx, "GitRemind", nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, "GitRemind", nk.NkVec2(float32(winWidth), float32(winHeight)))

	if update > 0 {

		ButtonBox(ctx)

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

	nk.NkLayoutRowDynamic(ctx, 400, 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			if len(repos) > 0 {
				for _, vv := range repos {
					//node := vv.SubNodes[i]
					name := vv[0]

					if nk.NkButtonLabel(ctx, name) > 0 {
						targetDir = name
						detailDisplay = "Conditions\n-----\n" + vv[4] + "\n\nFiles\n-----\n" + vv[1] + "\nDiff\n----\n" + vv[2]
					}
				}
			} else {

				if 0 < nk.NkButtonLabel(ctx, "No repos found, click to open a directory") {

				}
			}

			if 0 < nk.NkButtonLabel(ctx, "Change directory") {

			}

			if 0 < nk.NkButtonLabel(ctx, "Exit") {

				os.Exit(0)
			}
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 10, 1)
		{
			//Control the display
			nk.NkLayoutRowDynamic(ctx, 20, 3)
			{

				if 0 < nk.NkButtonLabel(ctx, "Pull") {
					Pull(targetDir)
				}

				if 0 < nk.NkButtonLabel(ctx, "Commit - Push") {
					CommitPush(targetDir)

				}

				if 0 < nk.NkButtonLabel(ctx, "Sync") {

				}

			}
			nk.NkLayoutRowDynamic(ctx, 10, 1)
			{
				results := strings.Split(detailDisplay, "\n")
				for _, v := range results {
					//nk.NkLabel(ctx, v, nk.WindowBorder)
					nk.NkLabel(ctx, v, nk.TextLeft)
				}
			}
		}
		nk.NkGroupEnd(ctx)
	}
}
