package main

import (
	"fmt"
	"log"
	"time"

	"github.com/korkmazkadir/bitcoin/consensus"
)

func main() {

	seed := time.Now().String()

	log.Println(seed)

	size := 1
	prob := float64(1) / (float64(600*1000) + 1)
	nbinom := consensus.NewNBinom(seed, size, prob)

	count := 0
	for i := 0; i < 1000; i++ {
		r := nbinom.Random()
		if r < 0 {
			log.Println(r)
			count++
		}
		fmt.Println(r)
	}

	log.Printf("Smaller tha 0 count is %d\n", count)

}
