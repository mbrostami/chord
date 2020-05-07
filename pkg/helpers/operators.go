package helpers

import (
	"crypto/sha256"
	"math/big"
)

// BetweenR check n ∈ (a, b]
func BetweenR(sourceHash [sha256.Size]byte, startHash [sha256.Size]byte, endHash [sha256.Size]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:sha256.Size])
	b.SetBytes(endHash[:sha256.Size])
	n.SetBytes(sourceHash[:sha256.Size])
	// a == b == n
	if a.Cmp(b) == 0 && b.Cmp(n) == 0 {
		return true
	}
	// a < b
	if a.Cmp(b) == -1 {
		// a < n <= b
		return a.Cmp(n) == -1 && (b.Cmp(n) == 1 || b.Cmp(n) == 0)
	}
	// a > b (circle ended)
	// a < n || n <= b
	return a.Cmp(n) == -1 || (b.Cmp(n) == 1 || b.Cmp(n) == 0)
}

// BetweenL check n ∈ [a, b)
func BetweenL(sourceHash [sha256.Size]byte, startHash [sha256.Size]byte, endHash [sha256.Size]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:sha256.Size])
	b.SetBytes(endHash[:sha256.Size])
	n.SetBytes(sourceHash[:sha256.Size])
	// a == b == n
	if a.Cmp(b) == 0 && b.Cmp(n) == 0 {
		return true
	}
	// a < b
	if a.Cmp(b) == -1 {
		// a <= n < b
		return (a.Cmp(n) == -1 || a.Cmp(n) == 0) && b.Cmp(n) == 1
	}
	// a > b (circle ended)
	// a <= n || n < b
	return (a.Cmp(n) == -1 || a.Cmp(n) == 0) || b.Cmp(n) == 1
}

// Between check n ∈ (a, b)
func Between(sourceHash [sha256.Size]byte, startHash [sha256.Size]byte, endHash [sha256.Size]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:sha256.Size])
	b.SetBytes(endHash[:sha256.Size])
	n.SetBytes(sourceHash[:sha256.Size])
	// a == b == n
	if a.Cmp(b) == 0 && b.Cmp(n) == 0 {
		return true
	}
	// a < b
	if a.Cmp(b) == -1 {
		// a < n < b
		return a.Cmp(n) == -1 && b.Cmp(n) == 1
	}
	// a > b (circle ended)
	// a < n || n < b
	return a.Cmp(n) == -1 || b.Cmp(n) == 1
}

// BetweenLR check n ∈ [a, b]
func BetweenLR(sourceHash [sha256.Size]byte, startHash [sha256.Size]byte, endHash [sha256.Size]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:sha256.Size])
	b.SetBytes(endHash[:sha256.Size])
	n.SetBytes(sourceHash[:sha256.Size])
	// a == b == n
	if a.Cmp(b) == 0 && b.Cmp(n) == 0 {
		return true
	}
	// a < b
	if a.Cmp(b) == -1 {
		// a <= n <= b
		return (a.Cmp(n) == -1 || a.Cmp(n) == 0) && (b.Cmp(n) == 1 || a.Cmp(n) == 0)
	}
	// a > b (circle ended)
	// a <= n || n <= b
	return (a.Cmp(n) == -1 || a.Cmp(n) == 0) || (b.Cmp(n) == 1 || a.Cmp(n) == 0)
}

// Equal check if a, b are equal
func Equal(sourceHash [sha256.Size]byte, targetHash [sha256.Size]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:sha256.Size])
	b.SetBytes(targetHash[:sha256.Size])
	// a == b == n
	return a.Cmp(b) == 0
}
