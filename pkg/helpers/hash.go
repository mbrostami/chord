package helpers

import (
	"crypto/sha256"
)

// Hash create hash from ip:port
func Hash(key string) [sha256.Size]byte {
	return sha256.Sum256([]byte(key))
}
