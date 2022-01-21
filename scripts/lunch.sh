#!/bin/bash

ansible-playbook -i hosts ./playbooks/install-dependencies.yml

ansible-playbook -i hosts ./playbooks/upload-artifacts.yml

ansible-playbook -i hosts ./playbooks/deploy-experiment.yml