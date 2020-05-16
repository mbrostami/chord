package chord

import "crypto/sha1"

const HashSize int = sha1.Size

func Hash(key string) [HashSize]byte {
	return sha1.Sum([]byte(key))
}
