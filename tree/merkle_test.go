package tree

import (
	"testing"
	"time"
)

func TestMakeBlocksWithExistingTime(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.Add(-20 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1, Hash(value1))
	row2 := MakeRow(ctime2, value2, Hash(value2))
	row3 := MakeRow(ctime3, value3, Hash(value3))

	blockID1 := CalculateBlockIndex(now, ctime1)
	blockID2 := CalculateBlockIndex(now, ctime2)
	blockID3 := CalculateBlockIndex(now, ctime3)

	hash1 := Hash(value1)
	hash2 := Hash(value2)
	hash3 := Hash(value3)

	blockHash1 := Hash(hash1[:])
	blockHash2 := Hash(hash2[:])
	blockHash3 := Hash(hash3[:])

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	for index, block := range tree.Blocks {
		if index == blockID1 {
			if blockHash1 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID1, blockHash1, *block.Hash)
			}
		} else if index == blockID2 {
			if blockHash2 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID2, blockHash2, *block.Hash)
			}
		} else if index == blockID3 {
			if blockHash3 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID3, blockHash3, *block.Hash)
			}
		}
	}
}

func TestMakeBlocksDataInSameBlockWithExistingTime(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1, Hash(value1))
	row2 := MakeRow(ctime1, value2, Hash(value2))
	row3 := MakeRow(ctime3, value3, Hash(value3))

	blockID12 := CalculateBlockIndex(now, ctime1)
	blockID3 := CalculateBlockIndex(now, ctime3)

	hash1 := Hash(value1) // 70cb3b7
	hash2 := Hash(value2) // 426b7f1
	hash3 := Hash(value3) // 7bb369b
	// first block hash is 0 hashsize bytes
	blockHash12 := Hash(hash1[:])
	blockHash12 = Hash(append(blockHash12[:], hash2[:]...))
	blockHash3 := Hash(hash3[:])

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)
	if !BytesEqual(tree.firstRow.Hash, hash2) {
		t.Errorf("First hash expected to be %x, got %x", hash2, tree.firstRow.Hash)
	}
	if !BytesEqual(tree.lastRow.Hash, hash3) {
		t.Errorf("Last hash expected to be %x, got %x", hash3, tree.lastRow.Hash)
	}
	if len(tree.Blocks) != 2 {
		t.Errorf("Expected to have 2 blocks got %d", len(tree.Blocks))
	}
	for index, block := range tree.Blocks {
		if index == blockID12 {
			if blockHash12 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID12, blockHash12, *block.Hash)
			}
		} else if index == blockID3 {
			if blockHash3 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID3, blockHash3, *block.Hash)
			}
		}
	}
}

func TestMakeBlocksDataInSameBlock(t *testing.T) {
	var rows []*Row

	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	now := time.Now()
	ctime1 := now.Add(-10 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1, Hash(value1))
	row2 := MakeRow(ctime1, value2, Hash(value2))
	row3 := MakeRow(ctime3, value3, Hash(value3))

	hash1 := Hash(value1)
	hash2 := Hash(value2)
	hash3 := Hash(value3)

	blockHash12 := Hash(hash1[:])
	blockHash12 = Hash(append(blockHash12[:], hash2[:]...))
	blockHash3 := Hash(hash3[:])

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkle(rows)

	blockID12 := CalculateBlockIndex(now, tree.sourceTime)
	blockID3 := CalculateBlockIndex(now, tree.sourceTime)

	if len(tree.Blocks) != 2 {
		t.Errorf("Expected to have 2 blocks got %d", len(tree.Blocks))
	}
	for index, block := range tree.Blocks {
		if index == blockID12 {
			if blockHash12 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID12, blockHash12, *block.Hash)
			}
		} else if index == blockID3 {
			if blockHash3 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID3, blockHash3, *block.Hash)
			}
		}
	}
}

func TestMakeBlocksOldCreationTime(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.AddDate(0, -12, 0) // log based 2 (365*24*3600) => 25
	ctime3 := now.AddDate(0, -11, 0) // log based 2 (335*24*3600) => 25

	row1 := MakeRow(ctime1, value1, Hash(value1))
	row2 := MakeRow(ctime2, value2, Hash(value2))
	row3 := MakeRow(ctime3, value3, Hash(value3))

	blockID1 := CalculateBlockIndex(now, ctime1)
	blockID2 := CalculateBlockIndex(now, ctime2)
	blockID3 := CalculateBlockIndex(now, ctime3)

	hash1 := Hash(value1)
	hash2 := Hash(value2)
	hash3 := Hash(value3)

	blockHash1 := Hash(hash1[:])
	// row2 and row3 are in the same block
	blockHash23 := Hash(hash2[:])
	blockHash23 = Hash(append(blockHash23[:], hash3[:]...))

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	for index, block := range tree.Blocks {
		if index == blockID2 && index == blockID3 {
			if blockHash23 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID2, blockHash23, *block.Hash)
			}
		} else if index == blockID1 {
			if blockHash1 != *block.Hash {
				t.Errorf("Hash values are not same %d expected %x, got %x", blockID1, blockHash1, *block.Hash)
			}
		}
	}
}

func TestMakeLeafs(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.AddDate(0, -12, 0) // log based 2 (365*24*3600) => 25
	ctime3 := now.AddDate(0, -11, 0) // log based 2 (335*24*3600) => 25

	row1 := MakeRow(ctime1, value1, Hash(value1))
	row2 := MakeRow(ctime2, value2, Hash(value2))
	row3 := MakeRow(ctime3, value3, Hash(value3))

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	hash1 := Hash(value1)
	hash2 := Hash(value2)
	hash3 := Hash(value3)

	blockHash1 := Hash(hash1[:])
	// row2 and row3 are in the same block
	blockHash23 := Hash(hash2[:])
	blockHash23 = Hash(append(blockHash23[:], hash3[:]...))

	tree := MakeMerkleWithTime(rows, now)

	// 2 leafs + root
	if len(tree.Nodes) != 3 {
		t.Errorf("Number of nodes expected %d, got %d", 3, len(tree.Nodes))
	}
	for index, node := range tree.Nodes {
		if node.Level == 0 {
			if index == 0 {
				if node.Hash != blockHash1 {
					t.Errorf("Leaf nodes are not valid, expected %x, got %x", blockHash1, node.Hash)
				}
			} else if index == 1 {
				if node.Hash != blockHash23 {
					t.Errorf("Leaf nodes are not valid, expected %x, got %x", blockHash23, node.Hash)
				}
			}
		} else if node.Level == 1 { // root
			rootHash := Hash(append(blockHash1[:], blockHash23[:]...))
			if node.Hash != rootHash {
				t.Errorf("Root hash is not valid expected %x, got %x", rootHash, node.Hash)
			}
			if node.Hash != tree.Root {
				t.Errorf("Root hash is not the same as last level node hash %x != %x", tree.Root, node.Hash)
			}
		}
	}
}
