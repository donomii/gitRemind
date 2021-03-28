// SPDX-License-Identifier: Unlicense OR MIT

package main

// A Gio program that demonstrates Gio widgets. See https://gioui.org for more information.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/donomii/goof"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/donomii/gitremind"

	"golang.org/x/exp/shiny/materialdesign/icons"
)

var screenshot = flag.String("screenshot", "", "save a screenshot to a file and exit")
var disable = flag.Bool("disable", false, "disable all widgets")

func main() {
	flag.Parse()
	go gitremind.DoScan("../../", false, false)
	ic, err := widget.NewIcon(icons.ContentAdd)
	if err != nil {
		log.Fatal(err)
	}
	icon = ic
	progressIncrementer = make(chan float32)

	go func() {
		for {
			time.Sleep(time.Second)
			progressIncrementer <- 0.1
		}
	}()

	go func() {
		w := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(700)))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	editor.SetText(longText)
	var ops op.Ops
	for {

		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if *disable {
					gtx = gtx.Disabled()
				}
				if checkbox.Changed() {
					if checkbox.Value {
						transformTime = e.Now
					} else {
						transformTime = time.Time{}
					}
				}
				if mode == "directories" {
					kitchen(gtx, th)
				} else {
					commitWindow(gtx, th)
				}
				e.Frame(gtx.Ops)
			}
		case p := <-progressIncrementer:
			progress += p
			if progress > 1 {
				progress = 0
			}
			w.Invalidate()
		}
	}
}

var (
	editor     = new(widget.Editor)
	lineEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	mode              string = "directories"
	buttons                  = map[string]*widget.Clickable{}
	greenButton              = new(widget.Clickable)
	iconTextButton           = new(widget.Clickable)
	iconButton               = new(widget.Clickable)
	flatBtn                  = new(widget.Clickable)
	disableBtn               = new(widget.Clickable)
	commitBtn                = new(widget.Clickable)
	commitCancelBtn          = new(widget.Clickable)
	commitPushBtn            = new(widget.Clickable)
	syncBtn                  = new(widget.Clickable)
	radioButtonsGroup        = new(widget.Enum)
	list                     = &layout.List{
		Axis: layout.Vertical,
	}
	panes = &layout.List{
		Axis: layout.Horizontal,
	}
	gitBar = &layout.List{
		Axis: layout.Horizontal,
	}
	commitBar = &layout.List{
		Axis: layout.Horizontal,
	}
	progress            = float32(0)
	progressIncrementer chan float32
	green               = true

	icon          *widget.Icon
	checkbox      = new(widget.Bool)
	swtch         = new(widget.Bool)
	transformTime time.Time
	float         = new(widget.Float)
	committingDir = ""
)

type (
	D = layout.Dimensions
	C = layout.Context
)

func kitchen(gtx layout.Context, th *material.Theme) layout.Dimensions {
	in := layout.UniformInset(unit.Dp(8))
	widgets := []layout.Widget{material.H3(th, "Repositories").Layout}
	widgets2 := []layout.Widget{material.H3(th, "Details").Layout}
	//	files := goof.LslR(".")

	keys := make([]string, len(gitremind.Repos))

	i := 0
	for k := range gitremind.Repos {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		file := k
		s := gitremind.Repos[file]

		if _, ok := buttons[file]; !ok {
			button := new(widget.Clickable)
			buttons[file] = button
		}
		widgets = append(widgets,
			func(gtx C) D {
				return in.Layout(gtx, func(gtx C) D {
					for buttons[file].Clicked() {

						longText = "Problems with " + s[0] + ":" + s[3] + s[2] + "\n" + s[1] + "\n\n"
						editor.SetText(longText)
						committingDir = file
					}

					pth := goof.SplitPath(file)
					dims := material.Button(th, buttons[file], pth[len(pth)-1]+" ("+s[3]+")").Layout(gtx)
					pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
					return dims
				})

			})

	}

	widgets2 = append(widgets2)

	widgets2 = append(widgets2,
		func(gtx C) D {

			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(500))
			return material.Editor(th, editor, "Hint").Layout(gtx)
		})

	gitBarList := []layout.Widget{
		func(gtx C) D {
			return in.Layout(gtx, func(gtx C) D {
				for syncBtn.Clicked() {
					cwd, _ := os.Getwd()
					os.Chdir(committingDir)
					cmd := exec.Command("git", "pull")
					goof.QuickCommand(cmd)
					os.Chdir(cwd)
					gitremind.RemoveRepo(committingDir)
					gitremind.ScanRepo(committingDir)

				}

				dims := material.Button(th, syncBtn, "Sync").Layout(gtx)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				return dims
			})
		},
		func(gtx C) D {
			return in.Layout(gtx, func(gtx C) D {
				for commitPushBtn.Clicked() {
					mode = "commit"
					editor.SetText("")

				}

				dims := material.Button(th, commitPushBtn, "Commit - Push").Layout(gtx)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				return dims
			})
		},
	}
	widgets2 = append(widgets2, func(gtx C) D {
		return gitBar.Layout(gtx, len(gitBarList), func(gtx C, i int) D {
			return layout.Center.Layout(gtx, gitBarList[i])
		})
	})
	paneList := []layout.Widget{
		func(gtx C) D {
			return list.Layout(gtx, len(widgets), func(gtx C, i int) D {
				return layout.Center.Layout(gtx, widgets[i])
			})
		},
		func(gtx C) D {
			return list.Layout(gtx, len(widgets2), func(gtx C, i int) D {
				return layout.Center.Layout(gtx, widgets2[i])
			})
		},
	}
	return panes.Layout(gtx, len(paneList), func(gtx C, i int) D {
		return layout.UniformInset(unit.Dp(1)).Layout(gtx, paneList[i])
	})

}

func commitWindow(gtx layout.Context, th *material.Theme) layout.Dimensions {
	in := layout.UniformInset(unit.Dp(8))
	widgets2 := []layout.Widget{material.H3(th, "Details").Layout}
	//	files := goof.LslR(".")

	widgets2 = append(widgets2,
		func(gtx C) D {

			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(500))
			return material.Editor(th, editor, longText).Layout(gtx)
		})

	commitBarList := []layout.Widget{
		func(gtx C) D {
			return in.Layout(gtx, func(gtx C) D {
				for commitBtn.Clicked() {

					fmt.Println("Clicked commit")
					fmt.Println(editor.Text())
					gitremind.CommitWithMessagePush(committingDir, editor.Text())
					delete(buttons, committingDir)
					gitremind.RemoveRepo(committingDir)
					mode = "directories"
					gitremind.ScanRepo(committingDir)
				}

				dims := material.Button(th, commitBtn, "Commit").Layout(gtx)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				return dims
			})
		},
		func(gtx C) D {
			return in.Layout(gtx, func(gtx C) D {
				for commitCancelBtn.Clicked() {
					mode = "directories"
				}

				dims := material.Button(th, commitCancelBtn, "Cancel").Layout(gtx)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				return dims
			})
		},
	}
	widgets2 = append(widgets2, func(gtx C) D {
		return commitBar.Layout(gtx, len(commitBarList), func(gtx C, i int) D {
			return layout.Center.Layout(gtx, commitBarList[i])
		})
	})
	return list.Layout(gtx, len(widgets2), func(gtx C, i int) D {
		return layout.Center.Layout(gtx, widgets2[i])
	})

}

var longText = `1. I learned from my grandfather, Verus, to use good manners, and to
put restraint on anger. 2. In the famous memory of my father I had a
pattern of modesty and manliness. 3. Of my mother I learned to be
pious and generous; to keep myself not only from evil deeds, but even
from evil thoughts; and to live with a simplicity which is far from
customary among the rich. 4. I owe it to my great-grandfather that I
did not attend public lectures and discussions, but had good and able
teachers at home; and I owe him also the knowledge that for things of
this nature a man should count no expense too great.

5. My tutor taught me not to favour either green or blue at the
chariot races, nor, in the contests of gladiators, to be a supporter
either of light or heavy armed. He taught me also to endure labour;
not to need many things; to serve myself without troubling others; not
to intermeddle in the affairs of others, and not easily to listen to
slanders against them.

6. Of Diognetus I had the lesson not to busy myself about vain things;
not to credit the great professions of such as pretend to work
wonders, or of sorcerers about their charms, and their expelling of
Demons and the like; not to keep quails (for fighting or divination),
nor to run after such things; to suffer freedom of speech in others,
and to apply myself heartily to philosophy. Him also I must thank for
my hearing first Bacchius, then Tandasis and Marcianus; that I wrote
dialogues in my youth, and took a liking to the philosopher's pallet
and skins, and to the other things which, by the Grecian discipline,
belong to that profession.

7. To Rusticus I owe my first apprehensions that my nature needed
reform and cure; and that I did not fall into the ambition of the
common Sophists, either by composing speculative writings or by
declaiming harangues of exhortation in public; further, that I never
strove to be admired by ostentation of great patience in an ascetic
life, or by display of activity and application; that I gave over the
study of rhetoric, poetry, and the graces of language; and that I did
not pace my house in my senatorial robes, or practise any similar
affectation. I observed also the simplicity of style in his letters,
particularly in that which he wrote to my mother from Sinuessa. I
learned from him to be easily appeased, and to be readily reconciled
with those who had displeased me or given cause of offence, so soon as
they inclined to make their peace; to read with care; not to rest
satisfied with a slight and superficial knowledge; nor quickly to
assent to great talkers. I have him to thank that I met with the
discourses of Epictetus, which he furnished me from his own library.

8. From Apollonius I learned true liberty, and tenacity of purpose; to
regard nothing else, even in the smallest degree, but reason always;
and always to remain unaltered in the agonies of pain, in the losses
of children, or in long diseases. He afforded me a living example of
how the same man can, upon occasion, be most yielding and most
inflexible. He was patient in exposition; and, as might well be seen,
esteemed his fine skill and ability in teaching others the principles
of philosophy as the least of his endowments. It was from him that I
learned how to receive from friends what are thought favours without
seeming humbled by the giver or insensible to the gift.`
