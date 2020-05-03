package helpers

import (
	"crypto/sha256"
	"strconv"
)

// Hash create hash from ip:port
func Hash(ip string, port int) [sha256.Size]byte {
	return sha256.Sum256([]byte(ip + ":" + strconv.FormatInt(int64(port), 10)))
}
