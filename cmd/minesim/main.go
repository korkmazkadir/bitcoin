package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/korkmazkadir/bitcoin/consensus"
)

func main() {

	var numberOfValues = flag.Int("count", 10000, "The number of produced values. The default is 10.000")
	var cc = flag.Int("cc", 1, "CC value. Default is 1")
	var s = flag.String("seed", "", "seed value value. If not provided uses local time")

	flag.Parse()

	var seed string
	if *s != "" {
		seed = *s
	} else {
		seed = time.Now().String()
	}

	log.Println(seed)

	size := 1
	prob := float64(*cc) / (float64(600 + *cc))
	nbinom := consensus.NewNBinom(seed, size, prob)

	count := 0
	for i := 0; i < *numberOfValues; i++ {
		r := nbinom.Random()
		if r < 0 {
			log.Println(r)
			count++
		}
		fmt.Println(r)
	}

	log.Printf("Smaller tha 0 count is %d\n", count)

}
