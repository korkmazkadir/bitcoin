package consensus

/*
#cgo LDFLAGS: -lRmath -lm
#define MATHLIB_STANDALONE 1
#include <Rmath.h>
#include <errno.h>
#include <stdlib.h>
#include <stdio.h>

extern double rnbinom_c(long int seed1, long int seed2, int size, double prob){
   set_seed(seed1,seed2);
   return rnbinom(size, prob);
}
*/
import "C"
import (
	"hash"
	"hash/fnv"
)

// https://github.com/SurajGupta/r-source/blob/master/src/nmath/rnbinom.c
// https://github.com/SurajGupta/r-source/blob/a28e609e72ed7c47f6ddfbb86c85279a0750f0b7/src/nmath/standalone/sunif.c
// install r-mathlib using following command
// sudo apt-get install -y r-mathlib

type NBinom struct {
	seedString string
	size       int
	prob       float64
	h          hash.Hash64
}

func NewNBinom(seedString string, size int, prob float64) *NBinom {
	nbinom := &NBinom{
		seedString: seedString + "glUa187N6j3BAqMhr8tw",
		size:       size,
		prob:       prob,
		h:          fnv.New64(),
	}

	nbinom.h.Write([]byte(nbinom.seedString))

	return nbinom
}

func (n *NBinom) Random() int {
	result := C.rnbinom_c(C.long(n.seed()), C.long(n.seed()), C.int(n.size), C.double(n.prob))

	return int(result)
}

func (n *NBinom) seed() uint64 {

	rndnumber := n.h.Sum64()
	n.h.Write(n.h.Sum(nil))
	return rndnumber
}
