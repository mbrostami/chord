package tree

// MerkleNode used as leafs/branches/root nodes in merkle tree
type MerkleNode struct {
	Level int
	hash  [HashSize]byte
	Left  *MerkleNode
	Right *MerkleNode
}

// GetHash returns the merkle node hash
func (mn *MerkleNode) GetHash() [HashSize]byte {
	return mn.hash
}
