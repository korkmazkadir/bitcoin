#!/bin/bash

wdir=$(pwd)
for folder in scripts[[:digit:]]; do
    echo $folder
    workingDir="$wdir/$folder"
    echo $workingDir
    gnome-terminal --tab --working-directory=$workingDir -- /bin/bash -c "./lunch.sh;bash"
done