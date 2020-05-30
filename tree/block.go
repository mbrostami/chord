package tree

import (
	"time"
)

type Block struct {
	index      uint
	rows       []*Row
	sourceTime time.Time
	Hash       *[HashSize]byte
}

// MakeBlock create new block
func MakeBlock(sourceTime time.Time, index uint) *Block {
	return &Block{
		index:      index,
		sourceTime: sourceTime,
	}
}

// GetSize returns the block size
func (b *Block) GetSize() int {
	return len(b.rows)
}

// GetIndex returns the block index
func (b *Block) GetIndex() uint {
	return b.index
}

// Append push new row to the end of the list
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

// ValidateIndex check if row is applicable for this block
func (b *Block) ValidateIndex(row *Row) bool {
	return CalculateBlockIndex(b.sourceTime, row.CreationTime) == b.index
}
