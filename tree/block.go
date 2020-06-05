package tree

import (
	"time"
)

// Block dynamic size block contains rows with the same round(log2(lifetime))
type Block struct {
	index      *uint
	rows       []*Row
	sourceTime time.Time
	Hash       *[HashSize]byte
}

// MakeBlock create new block
func MakeBlock(sourceTime time.Time, index uint) *Block {
	block := &Block{
		index:      &index,
		sourceTime: sourceTime,
	}
	return block
}

// GetSize returns the block size
func (b *Block) GetSize() int {
	return len(b.rows)
}

// GetIndex returns the block index
func (b *Block) GetIndex() uint {
	return *b.index
}

// Append pushes new row to the end of the list and update the block hash
func (b *Block) Append(row *Row) {
	b.rows = append(b.rows, row)
	// Make new hash (prev hash + new row hash)
	if b.Hash == nil {
		hash := Hash(row.Hash[:])
		b.Hash = &hash
	} else {
		hash := Hash(append(b.Hash[:], row.Hash[:]...))
		b.Hash = &hash
	}
}

// ValidateIndex check if row can be stored in this block
func (b *Block) ValidateIndex(row *Row) bool {
	return CalculateBlockIndex(b.sourceTime, row.CreationTime) == *b.index
}
