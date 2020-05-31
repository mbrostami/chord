package tree

import (
	"time"
)

// Row contains raw data to be stored and creation time
type Row struct {
	CreationTime time.Time
	Content      []byte
	Hash         [HashSize]byte
}

// MakeRow make new record based on creation time
func MakeRow(creationTime time.Time, content []byte) *Row {
	return &Row{
		CreationTime: creationTime,
		Content:      content,
		Hash:         Hash(content),
	}
}
