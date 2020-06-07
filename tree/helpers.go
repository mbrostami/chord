package tree

import (
	"math"
	"math/big"
	"time"

	"github.com/mbrostami/chord/helpers"
)

// HashSize size (bytes) of the hash function output
const HashSize = helpers.HashSize

// Hash calculate the hash of given data
func Hash(data []byte) [helpers.HashSize]byte {
	return helpers.Hash(string(data))
}

// CalculateBlockIndex calculate log2(source time - creation time) to use as block number
func CalculateBlockIndex(source time.Time, ctime time.Time) uint {
	duration := source.Sub(ctime)
	log2 := math.Round(math.Log2(duration.Seconds()))
	// if log2 duration is < 0, then we can consider that durations is less than 1 seconds so we use block number 0
	if log2 < 0 {
		return 0
	}
	// round log2 output to make an integer block number
	return uint(math.Round(math.Log2(duration.Seconds())))
}

// BytesLessThan check if a < b
func BytesLessThan(sourceHash [helpers.HashSize]byte, destHash [helpers.HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:helpers.HashSize])
	b.SetBytes(destHash[:helpers.HashSize])
	if a.Cmp(b) == -1 {
		return true
	}
	return false
}

// BytesEqual check if a = b
func BytesEqual(sourceHash [helpers.HashSize]byte, destHash [helpers.HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:helpers.HashSize])
	b.SetBytes(destHash[:helpers.HashSize])
	if a.Cmp(b) == 0 {
		return true
	}
	return false
}
