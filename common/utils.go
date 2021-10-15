package common

import (
	"crypto/sha256"
)

func MacroblockHash(blocks []Block) []byte {

	if len(blocks) == 1 {
		return blocks[0].Hash()
	}

	var hashSlice []byte
	for i := range blocks {
		hashSlice = append(hashSlice, blocks[i].Hash()...)
	}

	h := sha256.New()
	_, err := h.Write(hashSlice)
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}
