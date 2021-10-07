package consensus

import (
	"crypto/ed25519"
	"crypto/rand"
	"math"
	"math/big"
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
