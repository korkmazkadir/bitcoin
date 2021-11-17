#!/bin/bash

function deploy {

    experimentConfig="$1"
    
    # removes existing config file
    rm ./artifacts/config.json

    # moves new config file
    cp "$experimentConfig" ./artifacts/config.json

    printf %"$COLUMNS"s | tr " " "+"
    echo "uploading config..."
    printf %"$COLUMNS"s | tr " " "+"
    # upload config
    ansible-playbook -i hosts playbooks/upload-config.yml

    printf %"$COLUMNS"s | tr " " "+"
    echo "deploying experiment..."
    printf %"$COLUMNS"s | tr " " "+"
    # deploy experiment
    ansible-playbook -i hosts playbooks/deploy-experiment.yml

    printf %"$COLUMNS"s | tr " " "+"
    echo "waiting for the experiment..."
    printf %"$COLUMNS"s | tr " " "+"
    # wait for the end of experiment
    ansible-playbook -i hosts playbooks/wait-endof-experiment.yml

    printf %"$COLUMNS"s | tr " " "+"
    echo "downloading stats for the experiment..."
    printf %"$COLUMNS"s | tr " " "+"
    # downloading stats
    ansible-playbook -i hosts playbooks/download-stats.yml

    rm "$experimentConfig"
}


# install dependencies
# ansible-playbook -i hosts playbooks/install-dependencies.yml

# upload artifacts
# ansible-playbook -i hosts playbooks/upload-artifacts.yml

# iterates over config files
for filename in ./experiments_to_conduct/*.json; do

    printf %"$COLUMNS"s | tr " " "|"
    echo " deploying $experimentConfig"

    deploy "$filename"

done


# download stats
ansible-playbook -i hosts playbooks/download-stats.yml 




