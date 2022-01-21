package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/korkmazkadir/bitcoin/registery"
)

const (
	marker         = "RECEIVED"
	seperator      = "\t"
	bitcoinpp      = "Bitcoin++"
	bitcoinRedDiff = "BitcoinReducedDiff"
)

type ReceivedBlock struct {
	Issuer          int
	Height          int
	TimeString      string
	MicroblockIndex int
	Hash            string
	PrevHash        string
}

func main() {

	numberOfExperiments := 24

	statFile := getGlobalStatFile()

	for i := 1; i <= numberOfExperiments; i++ {

		configFilePath := fmt.Sprintf("/home/kadir/Desktop/Bitcoin_Fork_Data/experiment-%d/config.json", i)
		logFilePath := fmt.Sprintf("/home/kadir/Desktop/Bitcoin_Fork_Data/experiment-%d/1.log", i)

		nodeConfig := getConfig(configFilePath)
		blocks := parseLogFile(logFilePath)

		countMap := createMap(blocks)
		forkCountMap := calculateForkCounts(countMap)
		log.Println(forkCountMap)

		appendToLogs(nodeConfig, statFile, forkCountMap)

	}

	if err := statFile.Close(); err != nil {
		panic(err)
	}

}

func appendToLogs(config registery.NodeConfig, globalStatFile *os.File, forkCountMap map[int]int) {

	log.Printf("Processing BlockSize: %d, LeaderCount: %d, MiningTime: %f\n", config.BlockSize, config.LeaderCount, config.MiningTime)

	cc := config.LeaderCount
	expType := bitcoinpp
	if config.LeaderCount == 1 {
		expType = bitcoinRedDiff
		cc = int(float64(600) / config.MiningTime)
	}

	prefix := fmt.Sprintf("%d\t%d\t%s\t%d\t", config.BlockSize, cc, expType, config.EndRound)

	forkLine := ""
	for i := 1; i < 6; i++ {
		forkLine = fmt.Sprintf("%s%d\t", forkLine, forkCountMap[i])
	}
	forkLine = strings.TrimSuffix(forkLine, "\t")

	log.Println(forkLine)

	globalStatLine := fmt.Sprintf("%s%s", prefix, forkLine)
	_, err := fmt.Fprintln(globalStatFile, globalStatLine)
	if err != nil {
		panic(err)
	}
}

func getGlobalStatFile() *os.File {

	file, err := os.OpenFile("/home/kadir/Desktop/Bitcoin_Fork_Data/forks.stats", os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}

	return file
}

func getConfig(path string) registery.NodeConfig {

	config := registery.NodeConfig{}

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func calculateForkCounts(countMap map[int]int) map[int]int {

	forkCountsMap := make(map[int]int)

	// initializes count maps
	for i := 1; i < 6; i++ {
		forkCountsMap[i] = 0
	}

	currentRound := 0
	forked := false
	forkLength := 0

	count, ok := countMap[currentRound]

	forkStartRound := 0

	for ok {

		if count == 1 && forked {
			log.Printf("Fork found. Start Round:\t%d End Round:\t%d Lenfgth:\t%d\n", forkStartRound, currentRound, forkLength)

			forkCountsMap[forkLength] = forkCountsMap[forkLength] + 1
			forked = false
			forkLength = 0
		}

		if count > 1 && !forked {
			forked = true
			forkStartRound = currentRound
		}

		if forked {
			forkLength++
		}

		currentRound++
		count, ok = countMap[currentRound]

		if currentRound == 50 {
			//break
		}

	}

	return forkCountsMap
}

func createMap(blocks []ReceivedBlock) map[int]int {

	processedMap := make(map[string]struct{})
	countMap := make(map[int]int)

	for i := 0; i < len(blocks); i++ {
		b := blocks[i]

		_, ok := processedMap[b.PrevHash]
		if ok {
			continue
		}

		prevRound := b.Height - 1
		countMap[prevRound] = countMap[prevRound] + 1
		processedMap[b.PrevHash] = struct{}{}
	}

	return countMap
}

func parseLogFile(filePath string) []ReceivedBlock {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var receivedBlocks []ReceivedBlock

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, marker) {
			receivedBlockJson := strings.Split(line, seperator)[1]
			rb := ReceivedBlock{}
			err = json.Unmarshal([]byte(receivedBlockJson), &rb)
			if err != nil {
				panic(err)
			}
			receivedBlocks = append(receivedBlocks, rb)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return receivedBlocks
}
