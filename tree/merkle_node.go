package tree

type MerkleNode struct {
	Level int
	hash  [HashSize]byte
	Left  *MerkleNode
	Right *MerkleNode
}

// GetHash returns the node hash
func (mn *MerkleNode) GetHash() [HashSize]byte {
	return mn.hash
}
