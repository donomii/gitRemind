package main

import (
    "github.com/mattn/go-shellwords"
	"strings"
	//"text/scanner"
    "github.com/rivo/tview"
    "flag"
    "fmt"
    "crypto/md5"
    "encoding/hex"
    "io"
    "os"
    "os/exec"
    "bytes"
)


var autoSync bool
var ui bool
var repos [][]string
var lastSelect  string
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

func quickCommand (cmd *exec.Cmd) string{
    in := strings.NewReader("")
    cmd.Stdin = in
    var out bytes.Buffer
    cmd.Stdout = &out
    var err bytes.Buffer
    cmd.Stderr = &err
    cmd.Run()
    //fmt.Printf("Command result: %v\n", res)
    ret := fmt.Sprintf("%s", out)
    //fmt.Println(ret)
    return ret
}


func doQC (strs []string) string {
	cmd := exec.Command(strs[0], strs[1:]...)
    return quickCommand(cmd)
}

func quickCommandInteractive (cmd *exec.Cmd) {
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()
}



func doQCI (strs []string) {
	cmd := exec.Command(strs[0], strs[1:]...)
                quickCommandInteractive(cmd)
}

func doCommand(cmd string, args []string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "IO> %v\n", string(out))
		fmt.Fprintf(os.Stderr, "E> %v\n", err)
		//os.Exit(1)
	}
	if string(out) != "" {
		fmt.Fprintf(os.Stderr, "O> %v\n\n", string(out))
	}
    return string(out)
}


func grep (str string) string {
	var out string
	strs := strings.Split(str, "\n")
	for _, v := range strs {
		if strings.Index( v, "+") == 0 || strings.Index( v, "-") == 0  {
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

func doui(header string, cN *Node, cT []string) (currentNode *Node, currentThing []string){
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
        textView.SetText( "lalalala")
		
		textView2 := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
        textView2.SetText( "lalalala")
		
		footer := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
        footer.SetText( "lalalala")
		
		newPrimitive := func(text string) tview.Primitive {
			return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
		}
		
	list := tview.NewList()
        for i, vv := range currentNode.SubNodes {
            //node := vv.SubNodes[i]
            name := vv.Name
            v := vv
            list.AddItem(name, name, toChar(i), func(){
                currentThing = append(currentThing, name)
                currentNode = v
                app.Stop()
            })
        }
		list.AddItem("Quit", "Press to exit", 'q', func() {
            fmt.Println(strings.Join(currentThing, " ")+"\n")
			os.Exit(0)
		})


        	//menu := newPrimitive("Menu")
            	//sideBar := newPrimitive("Side Bar")

    grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive(header), 0, 0, 1, 3, 0, 0, false).
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
    return currentNode, currentThing
}

func main () {
var currentNode *Node
var currentThing []string
    flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
    flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
    flag.Parse()
    
	
    currentNode = createNodes()
    currentThing = []string{}
    for {
    currentNode, currentThing = doui(strings.Join(currentThing, " "), currentNode, currentThing)
}
}



type Node struct {
    Name string
    SubNodes []*Node
}

func findNode(n *Node, name string) *Node {
    for _, v:= range n.SubNodes {
        if v.Name == name {
            return v
        }
    }
    return nil

}


func createNodes() *Node {
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
    fmt.Println()
fmt.Printf("%+v\n", startNode)
dumpTree(&startNode, 0)
    return &startNode

}

func dumpTree(n *Node, indent int) {
    fmt.Printf("%*s%s\n", indent, "", n.Name)
    for _, v:= range n.SubNodes {
        dumpTree(v, indent+1)
    }

}

