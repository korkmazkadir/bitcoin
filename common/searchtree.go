package common

import (
	"bytes"
	"crypto/sha256"
)

type SearchTree struct {
	blockIndex int
	children   []*SearchTree
}

func filterIndexBlocks(roundBlocks []BlockMetadata, index int) (int, []int) {
	var indexes []int
	for i := 0; i < len(roundBlocks); i++ {
		if roundBlocks[i].Index() == index {
			indexes = append(indexes, i)
		}
	}
	return len(indexes), indexes
}

func constructSearchTreeRecursively(root []*SearchTree, index int, roundBlocks []BlockMetadata) {

	currentIndex := index + 1
	count, blockIndexes := filterIndexBlocks(roundBlocks, currentIndex)

	if count == 0 {
		return
	}

	//  creates nodes
	var nodes []*SearchTree
	for i := 0; i < count; i++ {
		n := &SearchTree{blockIndex: blockIndexes[i]}
		nodes = append(nodes, n)
	}

	// append each node as child
	for i := 0; i < len(root); i++ {
		//root[i].children = append(root[i].children, nodes...)
		root[i].children = nodes
	}

	constructSearchTreeRecursively(nodes, currentIndex, roundBlocks)

}

func NewSearchTree(roundBlocks []BlockMetadata) *SearchTree {
	root := &SearchTree{}
	constructSearchTreeRecursively([]*SearchTree{root}, -1, roundBlocks)
	return root
}

func (s *SearchTree) IsHashAvailable(roundBlocks []BlockMetadata, hashToSearch []byte) ([]int, bool) {
	return depthFirstSearchHash(s, roundBlocks, hashToSearch, nil)
}

func depthFirstSearchHash(root *SearchTree, roundBlocks []BlockMetadata, hashToSearch []byte, stack []int) ([]int, bool) {

	if len(root.children) == 0 {
		hash := constructSingleHashSelective(roundBlocks, stack)

		if bytes.Equal(hash, hashToSearch) {
			return stack, true
		}

		return []int{}, false
	}

	for i := 0; i < len(root.children); i++ {
		r, found := depthFirstSearchHash(root.children[i], roundBlocks, hashToSearch, append(stack, root.children[i].blockIndex))
		if found {
			return r, found
		}
	}

	return []int{}, false
}

func constructSingleHashSelective(blocks []BlockMetadata, blockIndexes []int) []byte {

	if len(blocks) == 1 {
		return blocks[0].Hash()
	}

	h := sha256.New()
	for i := 0; i < len(blockIndexes); i++ {
		h.Write(blocks[blockIndexes[i]].Hash())
	}
	return h.Sum(nil)
}
