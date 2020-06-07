package tree

// MerkleNode used as leafs/branches/root nodes in merkle tree
type MerkleNode struct {
	Level int            `json:"level"`
	Hash  [HashSize]byte `json:"hash"`
	left  *MerkleNode
	right *MerkleNode
}

// GetHash returns the merkle node hash
func (mn *MerkleNode) GetHash() [HashSize]byte {
	return mn.Hash
}
