#!/bin/bash

macroblock_sizes=(1)
concurrency_constants=(16 8 4 2 1)

rm -rf experiments_to_conduct
mkdir experiments_to_conduct

file_index=1
for macroblock_size in "${macroblock_sizes[@]}"
do
    macroblock_size_real=500000
    for cc in "${concurrency_constants[@]}"
    do

        printf -v file_name "%04d_%dMB_CC%d.json" ${file_index} ${macroblock_size} ${cc}
        echo "${file_name}"

        jq --arg bs "$macroblock_size_real" --arg cc "$cc" '.BlockSize =($bs|tonumber) | .LeaderCount =($cc|tonumber)   ' template_config.json > "./experiments_to_conduct/${file_name}"

        ((file_index++))
    done

done
