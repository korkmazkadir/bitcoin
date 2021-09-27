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
