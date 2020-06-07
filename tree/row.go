package tree

import (
	"time"

	"github.com/mbrostami/chord/helpers"
)

// Row contains raw data to be stored and creation time
type Row struct {
	CreationTime time.Time
	Content      []byte
	Hash         [helpers.HashSize]byte
}

// MakeRow make new record based on creation time
func MakeRow(creationTime time.Time, content []byte, hash [helpers.HashSize]byte) *Row {
	return &Row{
		CreationTime: creationTime,
		Content:      content,
		Hash:         hash,
	}
}
