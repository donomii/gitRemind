package main

import (
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


func worker (c chan string) {
    var ahead_regex = regexp.MustCompile(`Your branch is ahead of`)
    var not_staged_regex = regexp.MustCompile(`Changes not staged for commit:`)
    var modified_regex = regexp.MustCompile(`modified:`)
    var untracked_regex = regexp.MustCompile(`Untracked files:`)

        cwd, _ := os.Getwd()
        for path := range c {
            os.Chdir(cwd)
            if f, _ := os.Stat(fmt.Sprintf("%v/%v", path, ".git")); f != nil && f.IsDir() {
                os.Chdir(path)
                cmd := exec.Command("git", "status")
                result := quickCommand(cmd)
                if ahead_regex.MatchString(result) || modified_regex.MatchString(result) || not_staged_regex.MatchString(result) || untracked_regex.MatchString(result) {
                    fmt.Println(path)
                    //fmt.Println(result)
                    //fmt.Printf("\n\n\n\n\n")
                }
            }
        }
}

func main () {
    c := make(chan string)
    go worker (c)

    walkHandler := func (path string, info os.FileInfo, err error) error {
        c <- path
        return nil
    }
    fmt.Println("These repositories need some attention:")
    filepath.Walk(".", walkHandler)
    fmt.Println("Done!")
}
