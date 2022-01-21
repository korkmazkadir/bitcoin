package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	IniTemplate = `
%s

[all:vars]
ansible_user=root
ansible_python_interpreter=/usr/bin/python3
ansible_ssh_common_args='-o StrictHostKeyChecking=no'

[registry]
%s
`

	LogTailCMD = `gnome-terminal --tab -- /bin/bash -c 'ssh root@%s "tail -100f /root/rapidchain/output/1.log";bash'`
)

func readHostFile(pathOfHostFile string) []string {
	file, err := os.Open(pathOfHostFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hosts []string
	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return hosts
}

func constructHostFileContent(hosts []string, experimentCount int, machineCount int) ([]string, []string) {

	var hostFiles []string
	var logTailCommands []string
	for i := 0; i < experimentCount; i++ {
		currentHosts, remaininghosts := hosts[0:machineCount], hosts[machineCount:]

		registeryHost := currentHosts[0]
		allHosts := strings.Join(currentHosts[:], "\n")

		// creates a host file
		hostFiles = append(hostFiles, fmt.Sprintf(IniTemplate, allHosts, registeryHost))

		logTailCommands = append(logTailCommands, fmt.Sprintf(LogTailCMD, registeryHost))

		hosts = remaininghosts
	}

	return hostFiles, logTailCommands
}

func createHostFiles(hostFiles []string) {

	for i, c := range hostFiles {
		log.Printf("creating host file %d", i)
		hostFilePah := fmt.Sprintf("./hosts-%d", i)
		err := os.WriteFile(hostFilePah, []byte(c), 0644)
		if err != nil {
			panic(err)
		}
	}

}

func createTailScript(tailCommands []string) {

	tailCommands = append([]string{"#!/bin/bash"}, tailCommands...)
	scriptContent := strings.Join(tailCommands[:], "\n")

	log.Println("creating tails.sh")
	err := os.WriteFile("./tails.sh", []byte(scriptContent), 0755)
	if err != nil {
		panic(err)
	}

}

func main() {

	experimentCount := flag.Int("ec", -1, "experiment count")
	machineCountPerExperiment := flag.Int("mc", -1, "machine count per experiment")
	pathOfHostsFile := flag.String("h", "", "path of the hosts file")

	flag.Parse()

	if *experimentCount == -1 {
		panic("you did not provide -ec (experiment count) parameter")
	}

	if *machineCountPerExperiment == -1 {
		panic("you did not provide -mc (machine count per experiment) parameter")
	}

	if *pathOfHostsFile == "" {
		panic("you did not provide -h (path of host file) parameter")
	}

	hosts := readHostFile(*pathOfHostsFile)
	neededNodeCount := *experimentCount * *machineCountPerExperiment
	if len(hosts) < neededNodeCount {
		panic(fmt.Errorf("there are not enough nodes(%d): there are only %d nodes available", neededNodeCount, len(hosts)))
	}

	hostFileContents, tailCommands := constructHostFileContent(hosts, *experimentCount, *machineCountPerExperiment)

	createHostFiles(hostFileContents)
	createTailScript(tailCommands)

}
