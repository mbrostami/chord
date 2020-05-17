package helpers

import "crypto/sha1"

const HashSize int = sha1.Size

func Hash(key string) [HashSize]byte {
	return sha1.Sum([]byte(key))
}

func ConvertToHashSized(hash []byte) [HashSize]byte {
	var sourceHashResized [HashSize]byte
	copy(sourceHashResized[:HashSize], hash[:HashSize])
	return sourceHashResized
}
