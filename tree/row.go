package tree

import (
	"time"
)

type Row struct {
	CreationTime time.Time
	Content      []byte
	Hash         [HashSize]byte
	Index        int
}

// MakeRow timestamp
func MakeRow(creationTime time.Time, content []byte) *Row {
	return &Row{
		CreationTime: creationTime,
		Content:      content,
		Hash:         Hash(content),
	}
}
