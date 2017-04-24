# gitRemind
Searches your drive for git repositories, and tells you if they need to be pushed.

I have a lot of git repositories, and I often worked on them when I'm not connected to the internet, e.g. on a plane.  Then I forget to upload my changes, and end up with merge conflicts or missing work.  So I wrote gitRemind to check all me repositories and remind me to sync them.

# Building

    go get github.com/donomii/gitRemind

# Use

    ./gitRemind

gitRemind will recursively search every directory under the currect one.  If it find a git repository, it does a git status, and if there are changed files or you are ahead of the master branch, it will tell you.


