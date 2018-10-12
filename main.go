package main

import (
    "github.com/rivo/tview"
    "flag"
    "regexp"
    "strings"
    "fmt"
    "path/filepath"
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

func worker (c chan string) {
    var ahead_regex = regexp.MustCompile(`Your branch is ahead of`)
    var not_staged_regex = regexp.MustCompile(`Changes not staged for commit:`)
    var modified_regex = regexp.MustCompile(`modified:`)
    var untracked_regex = regexp.MustCompile(`Untracked files:`)
	repos = [][]string{}
        cwd, _ := os.Getwd()
        for path := range c {
            os.Chdir(cwd)
            if f, _ := os.Stat(fmt.Sprintf("%v/%v", path, ".git")); f != nil && f.IsDir() {
                os.Chdir(path)
                cmd := exec.Command("git", "status")
                result := quickCommand(cmd)
				cmd = exec.Command("git", "status", "--porcelain")
                shortresult := quickCommand(cmd)
				cmd = exec.Command("git", "diff", "--ignore-blank-lines")
                diffresult := quickCommand(cmd)
				reasons := []string{}
				longreasons := []string{}
                if ahead_regex.MatchString(result) {
					reasons = append(reasons, "!push")
					longreasons = append(longreasons, "local commits not pushed")
				}
				if modified_regex.MatchString(result) || not_staged_regex.MatchString(result) {
					reasons = append(reasons, "!commtd")
					longreasons = append(longreasons, "changes not committed")
				}				
				if untracked_regex.MatchString(result) {
					reasons = append(reasons, "!tracked")
					longreasons = append(longreasons, "untracked files present")
				}
				if len(reasons)>0 {
                    fmt.Printf("%v: %v\n", path, strings.Join(longreasons, ", "))
                    repos = append(repos, []string{path, shortresult, grep(diffresult), strings.Join(reasons, ", "), strings.Join(longreasons, ", ")})
                    //fmt.Println(result)
                    //fmt.Printf("\n\n\n\n\n")
                }
                if autoSync  {
                    fmt.Println("Syncing "+path)
                    cmd := exec.Command("git", "push")
                    quickCommand(cmd)
                    cmd = exec.Command("git", "pull")
                    quickCommand(cmd)
                }
            }
        }
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
        for i, v := range repos {
			
            ii := i
            list.AddItem(v[0], v[3], 'a', func(){
			if lastSelect == v[0] {
				app.Stop()
				doQCI([]string{"git", "commit", v[0]})
				doQCI([]string{"git", "push"})
				scanRepos(workerChan)
				app.Run()
			}
			textView.SetText(repos[ii][2])
			textView2.SetText(repos[ii][1])
			footer.SetText(repos[ii][4])
			lastSelect = v[0]
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
    walkHandler := func (path string, info os.FileInfo, err error) error {
        c <- path
        return nil
    }
    //fmt.Println("These repositories need some attention:")
    filepath.Walk(".", walkHandler)
}

func main () {
    flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
    flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
    flag.Parse()
    
	workerChan = make(chan string)
    go worker (workerChan)
	scanRepos(workerChan)
	
    if ui { doui() }
    fmt.Println("Done!")
}
