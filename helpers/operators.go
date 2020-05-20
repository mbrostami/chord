package helpers

import (
	"math/big"
)

// BetweenR check n ∈ (a, b]
func BetweenR(sourceHash [HashSize]byte, startHash [HashSize]byte, endHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:HashSize])
	b.SetBytes(endHash[:HashSize])
	n.SetBytes(sourceHash[:HashSize])
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
func BetweenL(sourceHash [HashSize]byte, startHash [HashSize]byte, endHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:HashSize])
	b.SetBytes(endHash[:HashSize])
	n.SetBytes(sourceHash[:HashSize])
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
func Between(sourceHash [HashSize]byte, startHash [HashSize]byte, endHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:HashSize])
	b.SetBytes(endHash[:HashSize])
	n.SetBytes(sourceHash[:HashSize])
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
func BetweenLR(sourceHash [HashSize]byte, startHash [HashSize]byte, endHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	n := new(big.Int)
	a.SetBytes(startHash[:HashSize])
	b.SetBytes(endHash[:HashSize])
	n.SetBytes(sourceHash[:HashSize])
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