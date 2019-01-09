// gui.go
package main

import (
	//"C"
	"log"
	"os"
	"strings"
	"time"

	//"unsafe"

	_ "image/jpeg"
	_ "image/png"

	//"github.com/donomii/glim"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"
	"golang.org/x/image/font/gofont/goregular"
)

//type Image C.struct_nk_image

// Start nuklear
func startNuke() {
	log.Println("Starting nuke")
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

	log.Println("Loading Image")
	h, err := NewTextureFromFile("test.png", 480, 480)
	log.Println("Image loaded:", h.Handle, err)
	testim = nk.NkImageId(int32(h.Handle))
	/*
		withGlctx(func() {
			pic, w, h := glim.LoadImage("test.png")
			log.Println("Loaded image")
			testim = load_nk_image(pic, w, h)
			//var ti C.struct_nk_image = *(*C.struct_nk_image)(unsafe.Pointer(&testim))
			//var ti Image = Image(testim)
			//ti.w = 480
			log.Println("Uploaded image")
		})
	*/
	log.Println("Initialised gui")

	pane1 := func() {
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

	pane2 := func() {
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
	pane3 := func() {
		nk.NkButtonImage(ctx, testim)
	}
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
			ClassicEmail3Pane(win, ctx, state, pane1, pane2, pane3)
		}
	}

	//End Nuklear
}

func ClassicEmail3Pane(win *glfw.Window, ctx *nk.Context, state *State, pane1, pane2, pane3 func()) {
	//log.Println("Redraw")
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

		ButtonBox(ctx, pane1, pane2)
		nk.NkLayoutRowStatic(ctx, 480, 480, 1)
		{
			/*withGlctx(func() {
				pic, w, h := glim.LoadImage("test.png")
				log.Println("Loaded image")
				testim = load_nk_image(pic, w, h)
				log.Println("Uploaded image")
			})*/
			//log.Println("Loading Image")
			//h, _ := gfx.NewTextureFromFile("test.png", 480, 480)
			//log.Println("Image loaded:", h.Handle)
			pane3()

			//log.Println("Image displayed")
			//Control the display
		}
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

func ButtonBox(ctx *nk.Context, pane1, pane2 func()) {

	nk.NkLayoutRowDynamic(ctx, 400, 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			pane1()
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)

		nk.NkLayoutRowDynamic(ctx, 10, 1)
		{

			pane2()
		}
		nk.NkGroupEnd(ctx)
	}
}
