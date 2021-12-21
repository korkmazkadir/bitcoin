package consensus

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"math"
	"math/big"

	"github.com/korkmazkadir/bitcoin/common"
)

func produceRandomNonce() int64 {
	nonce, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		panic(err)
	}
	return nonce.Int64()
}

func Sign(digest []byte, privateKey []byte) []byte {

	return ed25519.Sign(privateKey, digest)
}

// MicroBlockIndex calculates the index according to siblings and returns it
func MicroBlockIndex(nonce int64, siblings [][]byte, concurrencyLevel int) int {

	microblockIndex := int(nonce % int64(concurrencyLevel))
	if len(siblings) == 0 || len(siblings[microblockIndex]) == 0 {
		return microblockIndex
	}

	for {
		microblockIndex = (microblockIndex + 1) % concurrencyLevel
		if len(siblings[microblockIndex]) == 0 {
			return microblockIndex
		}
	}

}

func GetFullestMacroblock(cc int, blocks []common.Block) ([]common.Block, int, [][]byte) {

	macroblocks := make(map[string][]common.Block)
	previousBlockHashes := make(map[string][][]byte)
	counts := make(map[string]int)

	for _, block := range blocks {
		prevStr := ToString(block.PrevBlockHashes)
		val, ok := macroblocks[prevStr]
		if !ok {
			val = make([]common.Block, cc)
		}

		i := MicroBlockIndex(block.Nonce, block.Siblings, cc)

		//TODO: I want to do this check val[i] == common.Block{}
		// the compararison is added to add priority to blocks
		if len(val[i].Payload) > 0 && bytes.Compare(val[i].Hash(), block.Hash()) > 0 {
			continue
		}

		// we need to increase the count once
		if len(val[i].Payload) == 0 {
			counts[prevStr] += 1
		}

		val[i] = block
		macroblocks[prevStr] = val
		previousBlockHashes[prevStr] = block.PrevBlockHashes

	}

	var key string
	count := 0
	for k, v := range counts {
		if v > count {
			key = k
			count = v
		}
	}

	return macroblocks[key], count, previousBlockHashes[key]
}

func ToString(data [][]byte) string {

	h := sha256.New()
	for _, d := range data {
		_, err := h.Write(d)
		if err != nil {
			panic(err)
		}
	}

	return string(h.Sum(nil))
}

func Equal(a [][]byte, b [][]byte) bool {

	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}

	return true
}
