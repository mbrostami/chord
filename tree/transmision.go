package tree

import (
	"encoding/json"
	"errors"
)

// JSONNode level and hash
type JSONNode struct {
	Level int            `json:"level"`
	Hash  [HashSize]byte `json:"hash"`
}

// Serialize serialize data
func Serialize(tree Merkle) ([]byte, error) {
	var jsonNodes []JSONNode
	for i := 0; i < len(tree.Nodes); i++ {
		jsonNodes = append(jsonNodes, JSONNode{
			Level: tree.Nodes[i].Level,
			Hash:  tree.Nodes[i].Hash,
		})
	}
	return json.Marshal(jsonNodes)
}

// Unserialize json data
func Unserialize(jsonData []byte) []JSONNode {
	var jsonNodes []JSONNode
	json.Unmarshal(jsonData, &jsonNodes)
	return jsonNodes
}

// GetMissingHashesInTree compare json data and merkle tree, return diffs
// but only level 0 nodes (leafs)
func GetMissingHashesInTree(jsonData []byte, tree Merkle) ([]byte, error) {
	jsonNodes := Unserialize(jsonData)
	if len(jsonNodes) == 0 {
		return nil, errors.New("there is no data to compare")
	}
	if len(jsonNodes) != len(tree.Nodes) {
		return getMissingsInTreeLeafs(jsonNodes, tree)
	}
	// to keep using serialize, unserialize we need to use jsonnode instead of simple hash
	var diffBlockNodes []JSONNode
	hasDiff := false
	// define root level as last level
	lastLevel := jsonNodes[len(jsonNodes)-1].Level
	// start from the last one which is root hash
	for i := len(jsonNodes) - 1; i >= 0; i-- {
		// if parent levels are the same, all children hashes are the same too
		if lastLevel != jsonNodes[i].Level && hasDiff == false {
			return nil, nil
		}
		lastLevel = jsonNodes[i].Level
		if !BytesEqual(tree.Nodes[i].Hash, jsonNodes[i].Hash) || tree.Nodes[i].Level != jsonNodes[i].Level {
			hasDiff = true
			// only leaf nodes considered as diffBlockNodes
			if jsonNodes[i].Level == 0 {
				// TODO this can be prepended to the list if we want to keep order of diffs
				diffBlockNodes = append(diffBlockNodes, jsonNodes[i])
			}
		}
		// at the same time, remove diffs which are already exist in tree, but have different order
		for index, blockNode := range diffBlockNodes {
			if BytesEqual(tree.Nodes[i].Hash, blockNode.Hash) {
				// remove diff
				//fmt.Printf("removing diff %x\n", blockNode.Hash)
				diffBlockNodes = append(diffBlockNodes[:index], diffBlockNodes[index+1:]...)
				break
			}
		}
	}
	return json.Marshal(diffBlockNodes)
}

// GetExtraHashesInTree compare json data and merkle tree, return extra nodes in merkle tree
// but only level 0 nodes (leafs)
func GetExtraHashesInTree(jsonData []byte, tree Merkle) ([]byte, error) {
	jsonNodes := Unserialize(jsonData)
	if len(tree.blockHashes) == 0 {
		return nil, errors.New("there is no data in tree to compare")
	}
	return getExtrasInTreeLeafs(jsonNodes, tree)
}

// getExtrasInTreeLeafs compare json data and merkle tree when number of elements is not matched
// only compare level 0 nodes (leafs)
// only returns leafs which are not in the json nodes
func getExtrasInTreeLeafs(jsonNodes []JSONNode, tree Merkle) ([]byte, error) {
	if len(tree.blockHashes) == 0 {
		return nil, errors.New("there is no data in tree to compare")
	}
	// to keep using serialize, unserialize we need to use jsonnode instead of simple hash
	var diffBlockNodes []JSONNode
	for _, hash := range tree.blockHashes {
		exists := false
		for i := 0; i < len(jsonNodes); i++ {
			if BytesEqual(jsonNodes[i].Hash, hash) {
				exists = true
				break
			}
		}
		if !exists {
			diffBlockNodes = append(diffBlockNodes, JSONNode{
				Level: 0,
				Hash:  hash,
			})
		}
	}
	if diffBlockNodes == nil {
		return nil, nil
	}
	return json.Marshal(diffBlockNodes)
}

// getMissingsInTreeLeafs compare json data and merkle tree when number of elements is not matched
// only compare level 0 nodes (leafs)
// only return json nodes which are not in the tree's leafs
func getMissingsInTreeLeafs(jsonNodes []JSONNode, tree Merkle) ([]byte, error) {
	if len(jsonNodes) == 0 {
		return nil, errors.New("there is no data to compare")
	}
	// to keep using serialize, unserialize we need to use jsonnode instead of simple hash
	var diffBlockNodes []JSONNode
	// start from the last one which is root hash
	for i := 0; i < len(jsonNodes); i++ {
		if jsonNodes[i].Level > 0 {
			break
		}
		exists := false
		for _, hash := range tree.blockHashes {
			if BytesEqual(jsonNodes[i].Hash, hash) {
				exists = true
				break
			}
		}
		if !exists {
			diffBlockNodes = append(diffBlockNodes, jsonNodes[i])
		}
	}
	if diffBlockNodes == nil {
		return nil, nil
	}
	return json.Marshal(diffBlockNodes)
}
