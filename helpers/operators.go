package helpers

import (
	"math/big"
)

// Equal check a == b
func Equal(sourceHash [HashSize]byte, destHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:HashSize])
	b.SetBytes(destHash[:HashSize])
	// a == b == n
	if a.Cmp(b) == 0 {
		return true
	}
	return false
}

// LessThan check if a < b
func LessThan(sourceHash [HashSize]byte, destHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:HashSize])
	b.SetBytes(destHash[:HashSize])
	// a > b
	if a.Cmp(b) == -1 {
		return true
	}
	return false
}

// GreaterThan check if a > b
func GreaterThan(sourceHash [HashSize]byte, destHash [HashSize]byte) bool {
	a := new(big.Int)
	b := new(big.Int)
	a.SetBytes(sourceHash[:HashSize])
	b.SetBytes(destHash[:HashSize])
	// a > b
	if a.Cmp(b) == 1 {
		return true
	}
	return false
}

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
	// a == b != n
	if a.Cmp(b) == 0 {
		return false
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
