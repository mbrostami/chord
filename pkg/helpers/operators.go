package helpers

import (
	"crypto/sha256"
	"math/big"
)

// OpGT greater than
func OpGT(sourceHash [sha256.Size]byte, targetHash [sha256.Size]byte) bool {
	sourceInt := new(big.Int)
	targetInt := new(big.Int)
	sourceInt.SetBytes(sourceHash[:sha256.Size])
	targetInt.SetBytes(targetHash[:sha256.Size])
	return sourceInt.Cmp(targetInt) == 1
}

// OpGTE greater than or equal
func OpGTE(sourceHash [sha256.Size]byte, targetHash [sha256.Size]byte) bool {
	sourceInt := new(big.Int)
	targetInt := new(big.Int)
	sourceInt.SetBytes(sourceHash[:sha256.Size])
	targetInt.SetBytes(targetHash[:sha256.Size])
	result := sourceInt.Cmp(targetInt)
	return result == 1 || result == 0
}

// OpLTE less than or equal
func OpLTE(sourceHash [sha256.Size]byte, targetHash [sha256.Size]byte) bool {
	sourceInt := new(big.Int)
	targetInt := new(big.Int)
	sourceInt.SetBytes(sourceHash[:sha256.Size])
	targetInt.SetBytes(targetHash[:sha256.Size])
	result := sourceInt.Cmp(targetInt)
	return result == -1 || result == 0
}

// OpLT less than
func OpLT(sourceHash [sha256.Size]byte, targetHash [sha256.Size]byte) bool {
	sourceInt := new(big.Int)
	targetInt := new(big.Int)
	sourceInt.SetBytes(sourceHash[:sha256.Size])
	targetInt.SetBytes(targetHash[:sha256.Size])
	return sourceInt.Cmp(targetInt) == -1
}
