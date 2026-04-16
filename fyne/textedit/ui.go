package textedit

import (
	"fmt"

	"github.com/donomii/gitremind"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type textEdit struct {
	cursorRow, cursorCol *widget.Label
	entry                *widget.Entry
	window               fyne.Window
	targetDir            string
}

func (e *textEdit) updateStatus() {
	e.cursorRow.SetText(fmt.Sprintf("%d", e.entry.CursorRow+1))
	e.cursorCol.SetText(fmt.Sprintf("%d", e.entry.CursorColumn+1))
}

func biggerButton(label string, tapped func()) (*widget.Button, fyne.CanvasObject) {
	b := widget.NewButton(label, tapped)
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(0, 60)) // Increase to 60 height
	return b, container.NewStack(rect, container.NewPadded(b))
}

func (e *textEdit) cut() {
	e.entry.TypedShortcut(&fyne.ShortcutCut{Clipboard: e.window.Clipboard()})
}

func (e *textEdit) copy() {
	e.entry.TypedShortcut(&fyne.ShortcutCopy{Clipboard: e.window.Clipboard()})
}

func (e *textEdit) paste() {
	e.entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: e.window.Clipboard()})
}

func (e *textEdit) SetText(t string) {
	e.entry.SetText(t)
}
func (e *textEdit) SetTargetDir(t string) {
	e.targetDir = t
}

func (e *textEdit) buildToolbar() *widget.Toolbar {
	return widget.NewToolbar(widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
		e.entry.SetText("")
	}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {
			e.cut()
		}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			e.copy()
		}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			e.paste()
		}))
}

// Show loads a new text editor for the user to type a commit message into
func Show(app fyne.App, targetDir string, onSuccess func()) *textEdit {
	window := app.NewWindow("Commit Message")
	//window.SetIcon(icon.TextEditorBitmap)

	entry := widget.NewMultiLineEntry()
	cursorRow := widget.NewLabel("1")
	cursorCol := widget.NewLabel("1")

	editor := &textEdit{
		cursorRow: cursorRow,
		cursorCol: cursorCol,
		entry:     entry,
		window:    window,
		targetDir: targetDir,
	}

	toolbar := editor.buildToolbar()

	// Create commit button separately to reference it for disabling
	var commitBtn *widget.Button
	var commitWrapper fyne.CanvasObject
	commitBtn, commitWrapper = biggerButton("Commit", func() {
		commitBtn.Disable()
		commitBtn.SetText("Working...")

		go func() {
			gitremind.CommitWithMessagePush(editor.targetDir, editor.entry.Text)
			if onSuccess != nil {
				onSuccess()
			}
			window.Close()
		}()
	})

	_, cancelWrapper := biggerButton("Cancel", func() { window.Close() })

	buttonBox := container.NewHBox(layout.NewSpacer(),
		commitWrapper,
		cancelWrapper)

	content := container.NewBorder(toolbar, buttonBox, nil, nil,
		widget.NewMultiLineEntry())
	// Note: The original code used widget.NewScrollContainer(entry) but MultiLineEntry wraps itself now
	// However, keeping consistent structure:

	scroll := container.NewScroll(entry)
	content = container.NewBorder(toolbar, buttonBox, nil, nil, scroll)

	editor.entry.OnCursorChanged = func() {
		editor.updateStatus()
	}

	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New", func() {
				editor.entry.SetText("")
			}),
		),
		fyne.NewMenu("Edit",
			fyne.NewMenuItem("Cut", editor.cut),
			fyne.NewMenuItem("Copy", editor.copy),
			fyne.NewMenuItem("Paste", editor.paste),
		),
	))

	window.SetContent(content)
	window.Resize(fyne.NewSize(480, 320))
	window.Show()

	// Focus the entry
	window.Canvas().Focus(entry)

	return editor
}
