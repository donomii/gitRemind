package gitremind

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/donomii/goof"
)

var workerChan chan string
var wg sync.WaitGroup

var Repos map[string][]string

func DoScan(scanDir string, verbose bool, autoSync bool) {
	log.Println("Starting scan")
	workerChan = make(chan string, 10)
	//	doneChan = make(chan bool)
	go worker(workerChan, verbose, autoSync)
	scanRepos(scanDir, workerChan)
	wg.Wait()
	//<-doneChan
	//close(doneChan)

	log.Println("Scan complete!")
}

func scanRepos(scanDir string, c chan string) {
	var git_regex = regexp.MustCompile(`\.git`)
	walkHandler := func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path)

		if !git_regex.MatchString(path) {
			wg.Add(1)
			c <- path

		}
		return nil
	}
	//fmt.Println("These repositories need some attention:")
	filepath.Walk(scanDir, walkHandler)
	//close(c)
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

func worker(c chan string, verbose bool, autoSync bool) {
	var ahead_regex = regexp.MustCompile(`Your branch is ahead of`)
	var not_staged_regex = regexp.MustCompile(`Changes not staged for commit:`)
	var staged_not_committed_regex = regexp.MustCompile(`Changes to be committed`)
	var modified_regex = regexp.MustCompile(`modified:`)
	var untracked_regex = regexp.MustCompile(`Untracked files:`)
	var behind_regex = regexp.MustCompile(`Your branch is behind`)
	var both_regex = regexp.MustCompile(`different commits each, respectively.`)

	cwd, _ := os.Getwd()
	for path := range c {

		os.Chdir(cwd)
		gitpath := fmt.Sprintf("%v/%v", path, ".git")
		if goof.IsDir(gitpath) {
			if verbose {
				log.Println(gitpath)
			}
			os.Chdir(path)
			cmd := exec.Command("git", "fetch")
			goof.QuickCommand(cmd)
			cmd = exec.Command("git", "status")
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
			if behind_regex.MatchString(result) {
				reasons = append(reasons, "pull")
				longreasons = append(longreasons, "remote branch changed")
			}
			if both_regex.MatchString(result) {
				reasons = append(reasons, "diverge")
				longreasons = append(longreasons, "remote branch and local branch changed")
			}
			if modified_regex.MatchString(result) || not_staged_regex.MatchString(result) || staged_not_committed_regex.MatchString(result) {
				reasons = append(reasons, "commit")
				longreasons = append(longreasons, "changes not committed")
			}
			if untracked_regex.MatchString(result) {
				reasons = append(reasons, "untracked")
				longreasons = append(longreasons, "untracked files present")
			}
			if len(reasons) > 0 {
				fmt.Printf("%v: %v\n", path, strings.Join(longreasons, ", "))
				//fullPath := fmt.Sprintf("%v/%v", cwd, path)
				if Repos == nil {
					Repos = map[string][]string{}
				}
				Repos[path] = []string{path, shortresult, grep(diffresult), strings.Join(reasons, ", "), strings.Join(longreasons, ", "), result}
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
		wg.Done()
	}

}

func CommitWithMessagePush(targetDir, message string) {
	cwd, _ := os.Getwd()
	fmt.Println("Current directory", cwd)
	fmt.Println("Target directory", targetDir)

	message = strings.Replace(message, "\r", "", -1) //Remove windows line endings

	os.Chdir(targetDir)
	messageLines := strings.Split(message, "\n")
	//Searches a list of strings, return any that match search.  Case insensitive

	var compactMessages = []string{}
	for _, v := range messageLines {
		if len(v) > 0 && v[0] != '#' {
			compactMessages = append(compactMessages, v)
		}
	}

	if len(compactMessages) > 0 {
		mess := strings.Join(compactMessages, "\n")
		fmt.Printf("%v\n", []string{"git", "commit", "-a"})
		goof.QCI([]string{"git", "commit", "-a", "-m", mess})
		goof.QCI([]string{"git", "push"})
		fmt.Println("Commit message:", mess)
	} else {
		fmt.Println("Not committing, due to empty message")
	}
	os.Chdir(cwd)
}

func RemoveRepo(repo string) {
	delete(Repos, repo)
}

func ScanRepo(path string) {
	wg.Add(1)
	workerChan <- path
}
