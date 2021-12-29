package common

import (
	"crypto/sha256"
)

func MacroblockHash(blocks []BlockMetadata) []byte {

	if len(blocks) == 1 {
		return blocks[0].Hash()
	}

	h := sha256.New()
	for i := 0; i < len(blocks); i++ {
		h.Write(blocks[i].Hash())
	}
	return h.Sum(nil)
}
