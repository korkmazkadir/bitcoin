#!/bin/bash

ansible-playbook -i hosts ./playbooks/install-dependencies.yml

ansible-playbook -i hosts ./playbooks/upload-artifacts.yml

ansible-playbook -i hosts ./playbooks/deploy-experiment.yml


registery=`grep -r -A 1  "\[registry\]" hosts | grep -v "\[registry\]"`

echo "Connecting Registery =======> $registery"

ssh root@"$registery" "tail -100f ~/rapidchain/output/13.log"