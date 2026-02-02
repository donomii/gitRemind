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

// GitRemind holds the state of the git scan
type GitRemind struct {
	Repos      map[string][]string
	reposLock  sync.RWMutex
	workerChan chan string
	wg         sync.WaitGroup
}

// NewGitRemind creates a new GitRemind instance
func NewGitRemind() *GitRemind {
	return &GitRemind{
		Repos: make(map[string][]string),
	}
}

// GetRepos returns a safe copy of the repositories map
func (g *GitRemind) GetRepos() map[string][]string {
	g.reposLock.RLock()
	defer g.reposLock.RUnlock()

	copy := make(map[string][]string)
	for k, v := range g.Repos {
		copy[k] = v
	}
	return copy
}

// GetRepo returns details for a single repo safely
func (g *GitRemind) GetRepo(path string) ([]string, bool) {
	g.reposLock.RLock()
	defer g.reposLock.RUnlock()
	val, ok := g.Repos[path]
	return val, ok
}

// RemoveRepo removes a repo from the list
func (g *GitRemind) RemoveRepo(repo string) {
	g.reposLock.Lock()
	defer g.reposLock.Unlock()
	delete(g.Repos, repo)
}

// SetRepo updates the details for a repository
func (g *GitRemind) SetRepo(path string, details []string) {
	g.reposLock.Lock()
	defer g.reposLock.Unlock()
	g.Repos[path] = details
}

// Scan scans a directory for git repositories
func (g *GitRemind) Scan(scanDir string, verbose bool, autoSync bool) {
	log.Println("Starting scan:", scanDir)
	// Larger buffer to prevent blocking
	g.workerChan = make(chan string, 1000)

	// Start multiple workers to process the file system faster
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go g.worker(verbose, autoSync)
	}

	g.scanRepos(scanDir)
	g.wg.Wait()
	close(g.workerChan) // Close channel when done
	log.Println("Scan complete!")
}

func (g *GitRemind) scanRepos(scanDir string) {
	var gitRegex = regexp.MustCompile(`\.git`)
	walkHandler := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Optimization: Only check directories
		if !info.IsDir() {
			return nil
		}

		if !gitRegex.MatchString(path) {
			g.wg.Add(1)
			g.workerChan <- path
		} else {
			// Ensure we don't fall through without handling
			return nil
		}
		return nil
	}
	filepath.Walk(scanDir, walkHandler)
}

func grep(str string) string {
	var out string
	strs := strings.Split(str, "\n")
	for _, v := range strs {
		if len(v) > 0 && (v[0] == '+' || v[0] == '-') {
			out = out + v + "\n"
		} else {
			// Skip lines that don't start with + or -
			continue
		}
	}
	return out
}

func (g *GitRemind) worker(verbose bool, autoSync bool) {
	for path := range g.workerChan {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in worker for path %s: %v", path, r)
					g.wg.Done()
				}
			}()
			g.ProcessRepo(path, verbose, autoSync)
		}()
	}
}

func (g *GitRemind) ProcessRepo(path string, verbose, autoSync bool) {
	defer g.wg.Done()

	gitpath := filepath.Join(path, ".git")
	if !goof.IsDir(gitpath) {
		return
	}

	if verbose {
		log.Println("Found repo:", path)
	}

	// Fetch
	cmd := exec.Command("git", "fetch")
	cmd.Dir = path
	goof.QuickCommand(cmd)

	// Status (Porcelain is better for parsing)
	shortResult := getGitPorcelain(path)
	statusResult := getGitStatus(path)

	// Identify issues
	reasons, longReasons := analyzeStatus(statusResult, shortResult)

	if len(reasons) > 0 {
		// Diff
		diffResult := getGitDiff(path)

		fmt.Printf("%v: %v\n", path, strings.Join(longReasons, ", "))

		details := []string{path, shortResult, grep(diffResult), strings.Join(reasons, ", "), strings.Join(longReasons, ", "), statusResult}
		g.SetRepo(path, details)

		if verbose {
			fmt.Println(statusResult)
			fmt.Printf("\n\n\n\n\n")
		}
	} else {
		// No issues, remove from list if it exists
		g.RemoveRepo(path)
	}

	if autoSync {
		performAutoSync(path)
	}
}

func getGitStatus(path string) string {
	cmd := exec.Command("git", "status")
	cmd.Dir = path
	res, _ := goof.QuickCommand(cmd)
	return res
}

func getGitPorcelain(path string) string {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	res, _ := goof.QuickCommand(cmd)
	return res
}

func getGitDiff(path string) string {
	cmd := exec.Command("git", "diff", "--ignore-blank-lines")
	cmd.Dir = path
	res, _ := goof.QuickCommand(cmd)
	return res
}

func analyzeStatus(statusResult, porcelainResult string) ([]string, []string) {
	var aheadRegex = regexp.MustCompile(`Your branch is ahead of`)
	var behindRegex = regexp.MustCompile(`Your branch is behind`)
	var bothRegex = regexp.MustCompile(`different commits each, respectively.`)

	reasons := []string{}
	longReasons := []string{}

	if aheadRegex.MatchString(statusResult) {
		reasons = append(reasons, "push")
		longReasons = append(longReasons, "local commits not pushed")
	}

	if behindRegex.MatchString(statusResult) {
		reasons = append(reasons, "pull")
		longReasons = append(longReasons, "remote branch changed")
	}

	if bothRegex.MatchString(statusResult) {
		reasons = append(reasons, "diverge")
		longReasons = append(longReasons, "remote branch and local branch changed")
	}

	// Use porcelain for file changes (more reliable)
	if len(porcelainResult) > 0 {
		reasons = append(reasons, "commit")
		longReasons = append(longReasons, "uncommitted changes")

		if strings.Contains(porcelainResult, "??") {
			reasons = append(reasons, "untracked")
			longReasons = append(longReasons, "untracked files present")
		}
	}

	return reasons, longReasons
}

func performAutoSync(path string) {
	fmt.Println("Syncing " + path)
	cmd := exec.Command("git", "push")
	cmd.Dir = path
	goof.QuickCommand(cmd)

	cmd = exec.Command("git", "pull")
	cmd.Dir = path
	goof.QuickCommand(cmd)

	cmd = exec.Command("git", "push")
	cmd.Dir = path
	goof.QuickCommand(cmd)
}

func CommitWithMessagePush(targetDir, message string) {
	cwd, _ := os.Getwd()
	fmt.Println("Current directory", cwd)
	fmt.Println("Target directory", targetDir)

	message = strings.Replace(message, "\r", "", -1) //Remove windows line endings

	messageLines := strings.Split(message, "\n")

	var compactMessages = []string{}
	for _, v := range messageLines {
		if len(v) > 0 && v[0] != '#' {
			compactMessages = append(compactMessages, v)
		}
	}

	if len(compactMessages) > 0 {
		mess := strings.Join(compactMessages, "\n")
		// goof.QCI doesn't support setting Dir easily as it takes []string.
		// We should construct cmd manually.

		fmt.Printf("%v\n", []string{"git", "commit", "-a", "-m", mess})

		cmd := exec.Command("git", "commit", "-a", "-m", mess)
		cmd.Dir = targetDir
		goof.QuickCommand(cmd)

		cmd = exec.Command("git", "push")
		cmd.Dir = targetDir
		goof.QuickCommand(cmd)

		fmt.Println("Commit message:", mess)
	} else {
		fmt.Println("Not committing, due to empty message")
	}
}

// ScanRepo adds a single repo to the scan queue
func (g *GitRemind) ScanRepo(path string) {
	g.wg.Add(1)
	g.workerChan <- path
}
