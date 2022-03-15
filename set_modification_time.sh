#!/bin/bash

set -e

UNTRACKED=$(git status --porcelain --untracked-files=no)
if [[ -n $UNTRACKED ]]; then
    echo "Uncommited changes detected." > /dev/stderr
    echo "This script resets the modification time of every file in the repo. It usually only makes sense to run it on a copy with no uncommited changes to files in the repo." > /dev/stderr
    echo "Output from git status: {$UNTRACKED}" > /dev/stderr

    if [ "$1" != "-f" ]; then
        echo "Aborting" > /dev/stderr
        exit 1
    fi
    echo "Carrying on, as -f was set." > /dev/stderr
fi

echo "Resetting modification time of all files in the repo to their last commit time..."
echo "This stops Make from unnecessarily re-generating auto-generated code that is already up-to-date in the repo."
os=$(uname -s)
git ls-tree -r --name-only HEAD | while read -r filename; do
  unixtime=$(git log -1 --format="%at" -- "$filename")
  if [ "$os" == "Darwin" ]; then
    touchtime=$(date -jr "$unixtime" +'%Y%m%d%H%M.%S')
  else
    touchtime=$(date -d "@$unixtime" +'%Y%m%d%H%M.%S')
  fi
  touch -t "$touchtime" "$filename"
done
