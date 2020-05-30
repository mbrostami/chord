package tree

import (
	"math"
	"math/big"
	"time"

	"github.com/mbrostami/chord/helpers"
)

const HashSize = helpers.HashSize

func Hash(data []byte) [HashSize]byte {
	return helpers.Hash(string(data))
}

func CalculateBlockIndex(source time.Time, ctime time.Time) uint {
	duration := source.Sub(ctime)
	return uint(math.Round(math.Log2(duration.Seconds())))
}

// BytesLessThan check if a < b
func BytesLessThan(sourceHash [HashSize]byte, destHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:HashSize])
	b.SetBytes(destHash[:HashSize])
	if a.Cmp(b) == -1 {
		return true
	}
	return false
}

// BytesEqual check if a < b
func BytesEqual(sourceHash [HashSize]byte, destHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:HashSize])
	b.SetBytes(destHash[:HashSize])
	if a.Cmp(b) == 0 {
		return true
	}
	return false
}
