package tree

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSerializeUnserialize(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.Add(-20 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1)
	row2 := MakeRow(ctime2, value2)
	row3 := MakeRow(ctime3, value3)

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	jsonData, _ := Serialize(*tree)
	jsonNodes := Unserialize(jsonData)

	if len(jsonNodes) != 7 {
		t.Errorf("Number of nodes are expected %d got %d", 7, len(jsonNodes))
	}
	for i := 0; i < len(jsonNodes); i++ {
		if !BytesEqual(tree.Nodes[i].Hash, jsonNodes[i].Hash) {
			t.Errorf("Mismatch merkle tree hash with unserialized node hash expected %x got %x", tree.Nodes[i].Hash, jsonNodes[i].Hash)
		}
		if tree.Nodes[i].Level != jsonNodes[i].Level {
			t.Errorf("Mismatch merkle tree level with unserialized node level expected %x got %x", tree.Nodes[i].Level, jsonNodes[i].Level)
		}
	}
	rootHash := jsonNodes[len(jsonNodes)-1].Hash
	if !BytesEqual(tree.Root, rootHash) {
		t.Errorf("Mismatch merkle tree root hash with unserialized root hash expected %x got %x", tree.Root, rootHash)
	}
}

func TestDiffsSameJson(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.Add(-20 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1)
	row2 := MakeRow(ctime2, value2)
	row3 := MakeRow(ctime3, value3)

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	jsonData, _ := Serialize(*tree)

	diffs, err := GetMissingHashesInTree(jsonData, *tree)
	if err != nil {
		t.Errorf("expected to get no error, error received %v", err)
	}
	if diffs != nil {
		t.Errorf("expected to receive no diff, %s received", string(diffs))
	}
}

func TestDiffsMissingInDestination(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")
	value4 := []byte("test row 4")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.Add(-20 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)
	ctime4 := now.Add(-10 * time.Second)

	row1 := MakeRow(ctime1, value1)
	row2 := MakeRow(ctime2, value2)
	row3 := MakeRow(ctime3, value3)
	row4 := MakeRow(ctime4, value4)

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	rows = append(rows, row4)
	treeNew := MakeMerkleWithTime(rows, now)

	jsonData, _ := Serialize(*treeNew)

	diffBlockHashes, err := GetMissingHashesInTree(jsonData, *tree)
	if err != nil {
		t.Errorf("expected to get no error, error received %v", err)
	}
	if diffBlockHashes == nil {
		t.Error("expected to receive diff, but no diffs detected")
	}
	var jsonNodes []JSONNode
	json.Unmarshal(diffBlockHashes, &jsonNodes)
	if len(jsonNodes) != 1 {
		t.Errorf("expected to receive 1 diff, got %d", len(jsonNodes))
	}
	hash4 := Hash(value4)
	blockHash := Hash(hash4[:])
	if !BytesEqual(jsonNodes[0].Hash, blockHash) {
		t.Errorf("expected to receive diff as %x, got %x", jsonNodes[0].Hash, blockHash)
	}
}

func TestDiffsMissingInSource(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.Add(-20 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1)
	row2 := MakeRow(ctime2, value2)
	row3 := MakeRow(ctime3, value3)

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	rows = rows[:2] // remove row 3
	treeNew := MakeMerkleWithTime(rows, now)

	jsonData, _ := Serialize(*treeNew)

	diffBlockHashes, err := GetMissingHashesInTree(jsonData, *tree)
	if err != nil {
		t.Errorf("expected to get no error, error received %v", err)
	}
	if diffBlockHashes != nil {
		t.Errorf("expected to receive no diff, but got %+v", string(diffBlockHashes))
	}
}

func TestExtraDiffs(t *testing.T) {
	var rows []*Row

	now := time.Now()
	value1 := []byte("test row 1")
	value2 := []byte("test row 2")
	value3 := []byte("test row 3")

	ctime1 := now.Add(-10 * time.Minute)
	ctime2 := now.Add(-20 * time.Minute)
	ctime3 := now.Add(-20 * time.Second)

	row1 := MakeRow(ctime1, value1)
	row2 := MakeRow(ctime2, value2)
	row3 := MakeRow(ctime3, value3)

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)

	tree := MakeMerkleWithTime(rows, now)

	rows = rows[:2] // remove row 3
	treeNew := MakeMerkleWithTime(rows, now)

	jsonData, _ := Serialize(*treeNew)

	diffBlockHashes, err := GetExtraHashesInTree(jsonData, *tree)
	if err != nil {
		t.Errorf("expected to get no error, error received %v", err)
	}
	var jsonNodes []JSONNode
	json.Unmarshal(diffBlockHashes, &jsonNodes)

	if jsonNodes == nil {
		t.Error("expected to receive 1 diff, but got null")
	}
	if len(jsonNodes) != 1 {
		t.Errorf("expected to receive 1 diff, got %d", len(jsonNodes))
	}
	hash3 := Hash(value3)
	blockHash3 := Hash(hash3[:]) // b11374738
	if !BytesEqual(jsonNodes[0].Hash, blockHash3) {
		t.Errorf("expected to receive diff as %x, got %x", jsonNodes[0].Hash, blockHash3)
	}
}
