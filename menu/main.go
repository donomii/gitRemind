package main

import (
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-shellwords"

	//"text/scanner"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/rivo/tview"
)

var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var app *tview.Application
var workerChan chan string

func hash_file_md5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

func quickCommand(cmd *exec.Cmd) string {
	in := strings.NewReader("")
	cmd.Stdin = in
	var out bytes.Buffer
	cmd.Stdout = &out
	var err bytes.Buffer
	cmd.Stderr = &err
	cmd.Run()
	//fmt.Printf("Command result: %v\n", res)
	ret := out.String()
	//fmt.Println(ret)
	return ret
}

func doQC(strs []string) string {
	cmd := exec.Command(strs[0], strs[1:]...)
	return quickCommand(cmd)
}

func quickCommandInteractive(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func doQCI(strs []string) {
	cmd := exec.Command(strs[0], strs[1:]...)
	quickCommandInteractive(cmd)
}

func doCommand(cmd string, args []string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		//fmt.Fprintf(os.Stderr, "IO> %v\n", string(out))
		//fmt.Fprintf(os.Stderr, "E> %v\n", err)
		//os.Exit(1)
	}
	if string(out) != "" {
		//fmt.Fprintf(os.Stderr, "O> %v\n\n", string(out))
	}
	return string(out)
}

func grep(search, str string) string {
	var out string
	strs := strings.Split(str, "\n")
	for _, v := range strs {
		if strings.Index(v, search) == 0 {
			out = out + v + "\n"
		}
	}
	return out
}

func toCharStr(i int) string {
	return string('A' - 1 + i)
}

func toChar(i int) rune {
	return rune('a' + i)
}

func NodesToStringArray(ns []*Node) []string {
	var out []string
	for _, v := range ns {
		out = append(out, v.Name)

	}
	return out

}
func doui(cN *Node, cT []*Node, extraText string) (currentNode *Node, currentThing []*Node, result string) {
	currentNode = cN
	currentThing = cT

	//box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
	app = tview.NewApplication()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textView.SetText(extraText)

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	footer.SetText("lalalala")

	newPrimitive := func(text string) *tview.TextView {
		p := tview.NewTextView()
		p.SetTextAlign(tview.AlignCenter).
			SetText(text)
		p.SetChangedFunc(func() {
			app.Draw()
		})
		return p
	}
	header := newPrimitive("")
	header.SetText(strings.Join(NodesToStringArray(currentThing), " "))
	header.SetTextColor(tcell.ColorRed)

	list := tview.NewList()
	populateList := func(list *tview.List) { os.Exit(0) }
	extendList := func(list *tview.List) {
		list.AddItem("Run", "Run your text", 'R', func() {
			app.Stop()
			//app.Suspend(func() {
			result = doQC(NodesToStringArray(currentThing[1:]))

			//})
			textView.SetText(result)
			app.Run()
		})
		list.AddItem("Run Interactive", "Run your text", 'R', func() {
			app.Stop()
			//app.Suspend(func() {
			//result = doQC(NodesToStringArray(currentThing[1:]))
			doQCI(NodesToStringArray(currentThing[1:]))
			//})
			textView.SetText(result)
			app.Run()
		})
		list.AddItem("Back", "Go back", 'B', func() {
			//app.Stop()
			if len(currentThing) > 1 {
				currentNode = currentThing[len(currentThing)-2]
				currentThing = currentThing[:len(currentThing)-1]
				header.SetText(strings.Join(NodesToStringArray(currentThing), " "))
				list.Clear()
				populateList(list)
			}
		})

		list.AddItem("Quit", "Press to exit", 'Q', func() {
			fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
			app.Stop()
			os.Exit(0)
		})
		app.Draw()
	}

	populateList = func(list *tview.List) {
		list.Clear()
		result = ""
		if strings.HasPrefix(currentNode.Name, "!") {

			//It's a shell command

			cmd := currentNode.Name[1:]
			result = doCommand("/bin/sh", []string{"-c", cmd})
		}

		if strings.HasPrefix(currentNode.Name, "&") {

			//It's an internal command

			cmd := currentNode.Name[1:]
			if cmd == "lslR" {
				result = strings.Join(lslR("."), "\n")
			}
		}

		if result != "" {
			execNode := Node{"Exec", []*Node{}}
			addTextNodes(&execNode, result)
			currentNode = &execNode
		}
		for i, vv := range currentNode.SubNodes {
			//node := vv.SubNodes[i]
			name := vv.Name
			v := vv
			list.AddItem(name, name, toChar(i), func() {
				if !strings.HasPrefix(name, "!") && !strings.HasPrefix(name, "&") {
					currentThing = append(currentThing, v)
				}
				currentNode = v

				header.SetText("\n" + strings.Join(NodesToStringArray(currentThing[1:]), " "))
				list.Clear()
				populateList(list)
				//app.Stop()
			})
		}
		extendList(list)
	}

	populateList(list)

	//menu := newPrimitive("Menu")
	//sideBar := newPrimitive("Side Bar")

	grid := tview.NewGrid().
		SetRows(3, 0, 2).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(footer, 2, 0, 1, 2, 0, 0, false)

	/*
		        grid.AddItem(menu, 0, 0, 1, 3, 0, 0, false).
		        AddItem(list, 1, 0, 1, 3, 0, 0, true).
				AddItem(sideBar, 0, 0, 1, 3, 0, 0, false)
	*/

	grid.AddItem(list, 1, 0, 1, 1, 0, 40, true).
		AddItem(textView, 1, 1, 1, 1, 0, 40, false)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
	return currentNode, currentThing, result
}

func lslR(dir string) []string {
	out := []string{}
	walkHandler := func(path string, info os.FileInfo, err error) error {
		out = append(out, path)
		return nil
	}
	//fmt.Println("These repositories need some attention:")
	filepath.Walk(dir, walkHandler)
	return out
}

func git() string {
	return `!ls
git status
git push
git pull
git commit \&lslR
git commit .
git rebase
git merge
git stash
git stash apply
git diff
git reset
git reset --hard
git branch -a
git add \&lslR
`
}

func main() {
	var currentNode *Node
	var currentThing []*Node
	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.Parse()

	currentNode = &Node{"Start", []*Node{}}

	//    currentNode = addHistoryNodes()
	currentNode = addTextNodes(currentNode, git())
	//currentNode = addTextNodes(currentNode,grep("git", doCommand("fish", []string{"-c", "history"})))
	currentThing = []*Node{currentNode}
	result := ""
	for {
		currentNode, currentThing, result = doui(currentNode, currentThing, result)
	}
}

type Node struct {
	Name     string
	SubNodes []*Node
}

func (n *Node) String() string {
	return n.Name
}

func (n *Node) ToString() string {
	return n.Name
}

func findNode(n *Node, name string) *Node {
	if n == nil {
		return n
	}
	for _, v := range n.SubNodes {
		if v.Name == name {
			return v
		}
	}
	return nil

}

func addHistoryNodes() *Node {
	src := doCommand("fish", []string{"-c", "history"})
	lines := strings.Split(src, "\n")
	startNode := Node{"Start", []*Node{}}
	for _, l := range lines {
		currentNode := &startNode
		/*
				var s scanner.Scanner
				s.Init(strings.NewReader(l))
				s.Filename = "example"
				for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			        text := s.TokenText()
					fmt.Printf("%s: %s\n", s.Position, text)
			        if findNode(currentNode, text) == nil {
			            newNode := Node{text, []*Node{}}
			            currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
			            currentNode = &newNode
			        } else {
			            currentNode = findNode(currentNode, text)
			        }
		*/
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := Node{text, []*Node{}}
				currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
				currentNode = &newNode
			} else {
				currentNode = findNode(currentNode, text)
			}

		}
	}
	return &startNode
}

func addTextNodes(startNode *Node, src string) *Node {
	lines := strings.Split(src, "\n")
	for _, l := range lines {
		currentNode := startNode
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := Node{text, []*Node{}}
				currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
				currentNode = &newNode
			} else {
				currentNode = findNode(currentNode, text)
			}
		}
	}

	//fmt.Println()
	//fmt.Printf("%+v\n", startNode)
	//dumpTree(startNode, 0)
	return startNode

}

func dumpTree(n *Node, indent int) {
	fmt.Printf("%*s%s\n", indent, "", n.Name)
	for _, v := range n.SubNodes {
		dumpTree(v, indent+1)
	}

}
