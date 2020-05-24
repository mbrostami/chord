package merkle

import "github.com/mbrostami/chord/helpers"

//TestContent implements the Content interface provided by merkletree and represents the content stored in the tree.
type TestContent struct {
	X []byte
}

//CalculateHash hashes the values of a TestContent
func (t TestContent) CalculateHash() ([]byte, error) {
	hash := helpers.Hash(string(t.X))
	return hash[:], nil
}

//Equals tests for equality of two Contents
func (t TestContent) Equals(other Content) (bool, error) {
	var source [helpers.HashSize]byte
	copy(t.X[:helpers.HashSize], source[:helpers.HashSize])
	var dest [helpers.HashSize]byte
	copy(other.(TestContent).X[:helpers.HashSize], dest[:helpers.HashSize])
	return helpers.Equal(source, dest), nil
}
