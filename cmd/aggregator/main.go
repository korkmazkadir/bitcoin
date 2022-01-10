package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/korkmazkadir/bitcoin/registery"
)

const configFile = "config.json"
const statFile = "stats.log"

func main() {

	var wg sync.WaitGroup
	globalStatFile := getGlobalStatFile()

	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() {
			return nil
		}

		config, err := getConfig(path)
		if err != nil {
			return nil
		}

		statFile, err := getStatFile(path)
		if err != nil {
			return nil
		}

		wg.Add(1)
		go func() {
			appendToLogs(config, statFile, globalStatFile)
			wg.Done()
		}()

		return nil
	})

	wg.Wait()

	if err != nil {
		panic(err)
	}

	if err := globalStatFile.Close(); err != nil {
		panic(err)
	}

}

func appendToLogs(config registery.NodeConfig, stats *os.File, globalStatFile *os.File) {

	log.Printf("Processing BlockSize: %d, LeaderCount: %d, MiningTime: %f\n", config.BlockSize, config.LeaderCount, config.MiningTime)

	scanner := bufio.NewScanner(stats)
	prefix := fmt.Sprintf("%d\t%d\t%f\t", config.BlockSize, config.LeaderCount, config.MiningTime)
	for scanner.Scan() {

		statLine := scanner.Text()
		globalStatLine := fmt.Sprintf("%s%s", prefix, statLine)
		_, err := fmt.Fprintln(globalStatFile, globalStatLine)
		if err != nil {
			panic(err)
		}
	}

}

func getGlobalStatFile() *os.File {

	file, err := os.OpenFile("experiment.stats", os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}

	return file
}

func getStatFile(path string) (*os.File, error) {

	file, err := os.Open(fmt.Sprintf("%s/%s", path, statFile))
	return file, err
}

func getConfig(path string) (registery.NodeConfig, error) {

	config := registery.NodeConfig{}

	file, err := os.Open(fmt.Sprintf("%s/%s", path, configFile))
	if err != nil {
		return config, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
