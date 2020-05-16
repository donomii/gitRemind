[![Build Status](https://travis-ci.org/donomii/gitRemind.svg?branch=master)](https://travis-ci.org/donomii/gitRemind) [![GoDoc](https://godoc.org/github.com/donomii/gitRemind?status.svg)](https://godoc.org/github.com/donomii/gitRemind)

# gitRemind
Searches your drive for git repositories, and tells you if they need to be committed or pushed.

I have a lot of git repositories, and I at the end of the night I have 10 or more repositories to update.  Then I forget to upload my changes, and end up with merge conflicts or missing work.  So I wrote gitRemind to check all my repositories and remind me to sync them.

# Building

    go get github.com/donomii/gitRemind

# Use

    ./gitRemind
    
    ./gitRemind --gui

gitRemind will recursively search every directory under the current one.  If it finds a git repository, it does a git status, and if there are changed files or you are ahead of the master branch, it will tell you.
