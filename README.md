[![Build Status](https://travis-ci.org/donomii/gitRemind.svg?branch=master)](https://travis-ci.org/donomii/gitRemind) [![GoDoc](https://godoc.org/github.com/donomii/gitRemind?status.svg)](https://godoc.org/github.com/donomii/gitRemind)

# gitRemind
Searches your drive for git repositories, and tells you if they need to be committed or pushed.

I have a lot of git repositories, and I often work on them when I'm not connected to the internet, e.g. on a plane.  Then I forget to upload my changes, and end up with merge conflicts or missing work.  So I wrote gitRemind to check all my repositories and remind me to sync them.

GitRemind now features an experimental text mode UI, to automate committing and pushing your changes.

# Building

    go get github.com/donomii/gitRemind

# Use

    ./gitRemind
    
    ./gitRemind --ui

gitRemind will recursively search every directory under the current one.  If it finds a git repository, it does a git status, and if there are changed files or you are ahead of the master branch, it will tell you.

# UI

Select the repository you want to examine, then press enter twice to commit and push it.
