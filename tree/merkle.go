package tree

import (
	"sort"
	"time"
)

type Merkle struct {
	blocks       map[uint]*Block
	blockIndexes []uint
	sourceTime   time.Time
	Root         [HashSize]byte
	firstRow     *Row
	lastRow      *Row
	nodes        []*MerkleNode
	blockHashes  [][HashSize]byte
}

func MakeMerkle(rows []*Row) *Merkle {
	sourceTime := time.Now()
	return MakeMerkleWithTime(rows, sourceTime)
}

func MakeMerkleWithTime(rows []*Row, sourceTime time.Time) *Merkle {
	tree := &Merkle{
		sourceTime: sourceTime,
	}
	blocks := make(map[uint]*Block)
	for i := 0; i < len(rows); i++ {
		// find smallest hash
		if tree.firstRow == nil {
			tree.firstRow = rows[i]
		} else if BytesLessThan(rows[i].Hash, tree.firstRow.Hash) {
			tree.firstRow = rows[i]
		}

		// find largest hash
		if tree.lastRow == nil {
			tree.lastRow = rows[i]
		} else if BytesLessThan(tree.lastRow.Hash, rows[i].Hash) {
			tree.lastRow = rows[i]
		}

		// find block number based on creation time of the row record
		blockIndex := CalculateBlockIndex(sourceTime, rows[i].CreationTime)
		if blocks[blockIndex] == nil {
			blocks[blockIndex] = MakeBlock(sourceTime, blockIndex)
			tree.blockIndexes = append(tree.blockIndexes, blockIndex)
		}
		// add record to the block and recalculate the block hash
		blocks[blockIndex].Append(rows[i])
	}
	// sort block numbers, cause we need to keep order of the blocks while making merkle tree
	sort.Slice(tree.blockIndexes, func(i, j int) bool { return tree.blockIndexes[i] < tree.blockIndexes[j] })

	tree.blocks = blocks
	tree.makeLeafs()
	tree.makeBranches(tree.nodes, 1)
	// last node in the tree is considered as root node/hash
	tree.Root = tree.nodes[len(tree.nodes)-1].hash

	return tree
}

// GetNodes return nodes in tree
func (m *Merkle) GetNodes() []*MerkleNode {
	return m.nodes
}

// GetBlocks return blocks in tree
func (m *Merkle) GetBlocks() map[uint]*Block {
	return m.blocks
}

func (m *Merkle) makeLeafs() {
	level := 0
	// use sorted block indexes to make leaf nodes in merkle tree
	for _, blockIndex := range m.blockIndexes {
		block := m.blocks[blockIndex]
		leafNode := &MerkleNode{
			Level: level,
			hash:  *block.Hash,
		}
		m.nodes = append(m.nodes, leafNode)
		m.blockHashes = append(m.blockHashes, leafNode.hash)
	}

	// number of leaf nodes must be even, if it's odd, then duplicate the last node
	if len(m.nodes)%2 == 1 {
		leafNode := &MerkleNode{
			Level: level,
			hash:  m.nodes[len(m.nodes)-1].hash, // last item should be duplicated
		}
		m.nodes = append(m.nodes, leafNode)
	}
}

// makeBranches make upper level branches in merkle tree using created leaf nodes
func (m *Merkle) makeBranches(nodelist []*MerkleNode, level int) {
	var updatedNodeList []*MerkleNode
	for i := 0; i < len(nodelist); i += 2 {
		leftIndex := i
		rightIndex := i + 1
		// last node, so duplicate that
		if rightIndex == len(nodelist) {
			rightIndex = i
		}
		contatinatedHash := append(nodelist[leftIndex].hash[:], nodelist[rightIndex].hash[:]...)
		hash := Hash(contatinatedHash)
		branch := &MerkleNode{
			Level: level,
			hash:  hash,
			Left:  nodelist[leftIndex],
			Right: nodelist[rightIndex],
		}
		updatedNodeList = append(updatedNodeList, branch)
		m.nodes = append(m.nodes, branch)
	}
	// if it's not the last node then calculate next level, until 1 node remains which is root node (root hash)
	if len(updatedNodeList) > 1 {
		m.makeBranches(updatedNodeList, level+1)
	}
}
