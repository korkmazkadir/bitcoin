#!/bin/bash

counter=0
for filename in ./hosts-*; do
    echo $filename
    filePath="../../scripts$counter/hosts"
    rm $filePath
    mv $filename $filePath
    counter=$((counter+1))
done