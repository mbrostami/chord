package tree

import (
	"sort"
	"time"
)

type Merkle struct {
	Blocks       map[uint]*Block `json:"-"`
	blockIndexes []uint
	sourceTime   time.Time
	Root         [HashSize]byte `json:"root"`
	firstRow     *Row
	lastRow      *Row
	Nodes        []*MerkleNode `json:"nodes"`
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

	tree.Blocks = blocks
	tree.makeLeafs()
	tree.makeBranches(tree.Nodes, 1)
	// last node in the tree is considered as root node/hash
	tree.Root = tree.Nodes[len(tree.Nodes)-1].Hash

	return tree
}

// GetNodes return nodes in tree
func (m *Merkle) GetNodes() []*MerkleNode {
	return m.Nodes
}

// GetBlocks return blocks in tree
func (m *Merkle) GetBlocks() map[uint]*Block {
	return m.Blocks
}

func (m *Merkle) makeLeafs() {
	level := 0
	// use sorted block indexes to make leaf nodes in merkle tree
	for _, blockIndex := range m.blockIndexes {
		block := m.Blocks[blockIndex]
		leafNode := &MerkleNode{
			Level: level,
			Hash:  *block.Hash,
		}
		m.Nodes = append(m.Nodes, leafNode)
		m.blockHashes = append(m.blockHashes, leafNode.Hash)
	}

	// number of leaf nodes must be even, if it's odd, then duplicate the last node
	if len(m.Nodes)%2 == 1 {
		leafNode := &MerkleNode{
			Level: level,
			Hash:  m.Nodes[len(m.Nodes)-1].Hash, // last item should be duplicated
		}
		m.Nodes = append(m.Nodes, leafNode)
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
		contatinatedHash := append(nodelist[leftIndex].Hash[:], nodelist[rightIndex].Hash[:]...)
		hash := Hash(contatinatedHash)
		branch := &MerkleNode{
			Level: level,
			Hash:  hash,
			left:  nodelist[leftIndex],
			right: nodelist[rightIndex],
		}
		updatedNodeList = append(updatedNodeList, branch)
		m.Nodes = append(m.Nodes, branch)
	}
	// if it's not the last node then calculate next level, until 1 node remains which is root node (root hash)
	if len(updatedNodeList) > 1 {
		m.makeBranches(updatedNodeList, level+1)
	}
}
