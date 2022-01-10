package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
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

	blocks := parseLogFile("/home/kadir/Desktop/1.log")
	log.Printf("number of blocks data is %d \n", len(blocks))

	countMap := createMap(blocks)
	log.Println(countMap)

	forkCountMap := make(map[int]int)
	calculateForkCounts(countMap, 0, forkCountMap)

	log.Println(forkCountMap)

}

func calculateForkCounts(countMap map[int]int, startRound int, forkCountMap map[int]int) {

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

const (
	marker    = "RECEIVED"
	seperator = "\t"
)

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
