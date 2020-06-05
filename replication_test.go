package chord

import (
	"testing"
	"time"

	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/tree"
)

func TestSingleRowMasterBlock(t *testing.T) {
	localNode := NewNode("127.0.0.1", 10002) // 7584b781f3a
	predecessorList := NewPredecessorList()
	remoteSender := MockRemoteNodeSenderInterface{}
	// as replica is 2 we only need 1 predecessor
	predecessorList.Nodes[0] = NewRemoteNode(NewNode("127.0.0.1", 10003), remoteSender) // 40551fa9
	replication := NewReplication(localNode, predecessorList, 2)

	var data []*tree.Row
	data = append(data, tree.MakeRow(time.Now(), []byte("a"))) // 86f7e437faa5 E (40551fa9, 7584b781f3a) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("b"))) // e9d71f5e E (40551fa9, 7584b781f3a) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("c"))) // 84a51684 E (40551fa9, 7584b781f3a) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("d"))) // 3c3638 E (40551fa9, 7584b781f3a) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("e"))) // 58e6b3a4 E (40551fa9, 7584b781f3a) => will be added to block
	replication.MakeTrees(data)
	if len(replication.masterBlocks) != 1 {
		t.Errorf("number of master blocks must be %d got %d", 1, len(replication.masterBlocks))
	}
	if len(replication.masterBlocks[0].rows) != 1 {
		t.Errorf("number of master blocks rows must be %d got %d", 1, len(replication.masterBlocks[0].rows))
	}

	// check if master block first row hash is the same as data hash
	rowHash := replication.masterBlocks[0].rows[0].Hash
	if !helpers.Equal(rowHash, helpers.Hash("e")) {
		t.Errorf("row hash must be %x got %x", rowHash, helpers.Hash("e"))
	}

	// check if master block first row value is the same as data value
	rowContent := replication.masterBlocks[0].rows[0].Content
	if string(rowContent) != "e" {
		t.Errorf("row content must be %s got %s", "e", string(rowContent))
	}

	// check number of merkle trees generated
	if len(replication.trees) != 1 {
		t.Errorf("number of trees must be %d got %d", 1, len(replication.trees))
	}

	// check number of merkle tree nodes generated
	if len(replication.trees[0].GetNodes()) != 3 {
		t.Errorf("number of tree nodes must be %d got %d", 3, len(replication.trees[0].GetNodes()))
	}

	var rootHash [helpers.HashSize]byte
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[0].Root[:]...)))
	if !helpers.Equal(rootHash, replication.RootHash) {
		t.Errorf("root hash must be %x got %x", rootHash, replication.RootHash)
	}
}

func TestTwoRowsMasterBlock(t *testing.T) {
	localNode := NewNode("127.0.0.1", 10004) //  bddf48b20eb1
	predecessorList := NewPredecessorList()
	remoteSender := MockRemoteNodeSenderInterface{}
	// as replica is 2 we only need 1 predecessor
	predecessorList.Nodes[0] = NewRemoteNode(NewNode("127.0.0.1", 10002), remoteSender) // 7584b781f3a

	replication := NewReplication(localNode, predecessorList, 2)

	var data []*tree.Row
	data = append(data, tree.MakeRow(time.Now(), []byte("a"))) // 86f7e437faa5 E (7584b781f3a, bddf48b20eb1) => will be added to block
	data = append(data, tree.MakeRow(time.Now(), []byte("b"))) // e9d71f5e E (7584b781f3a, bddf48b20eb1) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("c"))) // 84a51684 E (7584b781f3a, bddf48b20eb1) => will be added to block
	data = append(data, tree.MakeRow(time.Now(), []byte("d"))) // 3c3638 E (7584b781f3a, bddf48b20eb1) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("e"))) // 58e6b3a4 E (7584b781f3a, bddf48b20eb1) => skip
	replication.MakeTrees(data)

	if len(replication.masterBlocks) != 1 {
		t.Errorf("number of master blocks must be %d got %d", 1, len(replication.masterBlocks))
	}
	if len(replication.masterBlocks[0].rows) != 2 {
		t.Errorf("number of master blocks rows must be %d got %d", 2, len(replication.masterBlocks[0].rows))
	}

	// check if master block first row hash is the same as data hash
	rowHashA := replication.masterBlocks[0].rows[0].Hash
	rowHashC := replication.masterBlocks[0].rows[1].Hash
	if !helpers.Equal(rowHashA, helpers.Hash("a")) {
		t.Errorf("row hash must be %x got %x", rowHashA, helpers.Hash("a"))
	}
	if !helpers.Equal(rowHashC, helpers.Hash("c")) {
		t.Errorf("row hash must be %x got %x", rowHashC, helpers.Hash("c"))
	}

	// check number of merkle trees generated
	if len(replication.trees) != 1 {
		t.Errorf("number of trees must be %d got %d", 1, len(replication.trees))
	}

	// check number of merkle tree nodes generated
	if len(replication.trees[0].GetNodes()) != 3 {
		t.Errorf("number of tree nodes must be %d got %d", 3, len(replication.trees[0].GetNodes()))
	}

	var rootHash [helpers.HashSize]byte
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[0].Root[:]...)))
	if !helpers.Equal(rootHash, replication.RootHash) {
		t.Errorf("root hash must be %x got %x", rootHash, replication.RootHash)
	}
}

func TestMultipleRowsMasterBlock(t *testing.T) {
	localNode := NewNode("127.0.0.1", 10014) //  f88860c3ae
	predecessorList := NewPredecessorList()
	remoteSender := MockRemoteNodeSenderInterface{}
	// as replica is 2 we only need 1 predecessor
	predecessorList.Nodes[0] = NewRemoteNode(NewNode("127.0.0.1", 10015), remoteSender) // 4b107076bf0
	replication := NewReplication(localNode, predecessorList, 2)

	var data []*tree.Row
	data = append(data, tree.MakeRow(time.Now(), []byte("a"))) // 86f7e437faa5 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(time.Now(), []byte("b"))) // e9d71f5e E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(time.Now(), []byte("c"))) // 84a51684 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(time.Now(), []byte("d"))) // 3c3638 E (4b107076bf0, f88860c3ae) => skip
	data = append(data, tree.MakeRow(time.Now(), []byte("e"))) // 58e6b3a4 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(time.Now(), []byte("g"))) // 54fd17112 E (4b107076bf0, f88860c3ae) => will be added to block
	replication.MakeTrees(data)

	if len(replication.masterBlocks) != 1 {
		t.Errorf("number of master blocks must be %d got %d", 1, len(replication.masterBlocks))
	}
	if len(replication.masterBlocks[0].rows) != 5 {
		t.Errorf("number of master blocks rows must be %d got %d", 5, len(replication.masterBlocks[0].rows))
	}

	// check if master block first row hash is the same as data hash
	rowHashA := replication.masterBlocks[0].rows[0].Hash
	rowHashB := replication.masterBlocks[0].rows[1].Hash
	rowHashC := replication.masterBlocks[0].rows[2].Hash
	rowHashE := replication.masterBlocks[0].rows[3].Hash
	rowHashG := replication.masterBlocks[0].rows[4].Hash
	if !helpers.Equal(rowHashA, helpers.Hash("a")) {
		t.Errorf("row hash must be %x got %x", rowHashA, helpers.Hash("a"))
	}
	if !helpers.Equal(rowHashB, helpers.Hash("b")) {
		t.Errorf("row hash must be %x got %x", rowHashB, helpers.Hash("b"))
	}
	if !helpers.Equal(rowHashC, helpers.Hash("c")) {
		t.Errorf("row hash must be %x got %x", rowHashC, helpers.Hash("c"))
	}
	if !helpers.Equal(rowHashE, helpers.Hash("e")) {
		t.Errorf("row hash must be %x got %x", rowHashE, helpers.Hash("e"))
	}
	if !helpers.Equal(rowHashG, helpers.Hash("g")) {
		t.Errorf("row hash must be %x got %x", rowHashG, helpers.Hash("g"))
	}

	// check number of merkle trees generated
	if len(replication.trees) != 1 {
		t.Errorf("number of trees must be %d got %d", 1, len(replication.trees))
	}

	// check number of merkle tree nodes generated
	// although we use 5 rows to make tree, but all 5 rows will go to the same block
	// cause all have the same creation time so number of leaf nodes is (block hash + duplicated block hash + root hash) = 3
	if len(replication.trees[0].GetNodes()) != 3 {
		t.Errorf("number of tree nodes must be %d got %d", 3, len(replication.trees[0].GetNodes()))
	}

	var rootHash [helpers.HashSize]byte
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[0].Root[:]...)))
	if !helpers.Equal(rootHash, replication.RootHash) {
		t.Errorf("root hash must be %x got %x", rootHash, replication.RootHash)
	}
}

func TestMultipleRowsMasterBlockMultipleSubBlocks(t *testing.T) {
	localNode := NewNode("127.0.0.1", 10014) //  f88860c3ae
	predecessorList := NewPredecessorList()
	remoteSender := MockRemoteNodeSenderInterface{}
	// as replica is 2 we only need 1 predecessor
	predecessorList.Nodes[0] = NewRemoteNode(NewNode("127.0.0.1", 10015), remoteSender) // 4b107076bf0
	replication := NewReplication(localNode, predecessorList, 2)

	now := time.Now()
	var data []*tree.Row
	data = append(data, tree.MakeRow(now.Add(-10*time.Minute), []byte("a"))) // 86f7e437faa5 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-20*time.Minute), []byte("b"))) // e9d71f5e E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-30*time.Minute), []byte("c"))) // 84a51684 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-40*time.Minute), []byte("d"))) // 3c3638 E (4b107076bf0, f88860c3ae) => skip
	// next 2 records would be in same block
	data = append(data, tree.MakeRow(now.Add(-50*time.Minute), []byte("e"))) // 58e6b3a4 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-70*time.Minute), []byte("g"))) // 54fd17112 E (4b107076bf0, f88860c3ae) => will be added to block
	replication.MakeTrees(data)

	if len(replication.masterBlocks) != 1 {
		t.Errorf("number of master blocks must be %d got %d", 1, len(replication.masterBlocks))
	}
	if len(replication.masterBlocks[0].rows) != 5 {
		t.Errorf("number of master blocks rows must be %d got %d", 5, len(replication.masterBlocks[0].rows))
	}

	// check if master block first row hash is the same as data hash
	rowHashA := replication.masterBlocks[0].rows[0].Hash
	rowHashB := replication.masterBlocks[0].rows[1].Hash
	rowHashC := replication.masterBlocks[0].rows[2].Hash
	rowHashE := replication.masterBlocks[0].rows[3].Hash
	rowHashG := replication.masterBlocks[0].rows[4].Hash
	if !helpers.Equal(rowHashA, helpers.Hash("a")) {
		t.Errorf("row hash must be %x got %x", rowHashA, helpers.Hash("a"))
	}
	if !helpers.Equal(rowHashB, helpers.Hash("b")) {
		t.Errorf("row hash must be %x got %x", rowHashB, helpers.Hash("b"))
	}
	if !helpers.Equal(rowHashC, helpers.Hash("c")) {
		t.Errorf("row hash must be %x got %x", rowHashC, helpers.Hash("c"))
	}
	if !helpers.Equal(rowHashE, helpers.Hash("e")) {
		t.Errorf("row hash must be %x got %x", rowHashE, helpers.Hash("e"))
	}
	if !helpers.Equal(rowHashG, helpers.Hash("g")) {
		t.Errorf("row hash must be %x got %x", rowHashG, helpers.Hash("g"))
	}

	// check number of merkle trees generated
	if len(replication.trees) != 1 {
		t.Errorf("number of trees must be %d got %d", 1, len(replication.trees))
	}

	// check number of merkle tree nodes generated
	// although we use 5 rows to make tree, but last 2 rows will go to the same block so we would have 4 blocks
	blocks := replication.trees[0].GetBlocks()
	if len(blocks) != 4 {
		t.Errorf("number of blocks in tree must be %d got %d", 4, len(blocks))
	}
	// number of leaf nodes is (4 block hash + 2 branch hash + root hash) = 7
	if len(replication.trees[0].GetNodes()) != 7 {
		t.Errorf("number of tree nodes must be %d got %d", 7, len(replication.trees[0].GetNodes()))
	}

	var rootHash [helpers.HashSize]byte
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[0].Root[:]...)))
	if !helpers.Equal(rootHash, replication.RootHash) {
		t.Errorf("root hash must be %x got %x", rootHash, replication.RootHash)
	}
}

func TestMultipleRowsMasterBlockMultipleSubBlocksOdd(t *testing.T) {
	localNode := NewNode("127.0.0.1", 10014) //  f88860c3ae
	predecessorList := NewPredecessorList()
	remoteSender := MockRemoteNodeSenderInterface{}
	// as replica is 2 we only need 1 predecessor
	predecessorList.Nodes[0] = NewRemoteNode(NewNode("127.0.0.1", 10015), remoteSender) // 4b107076bf0
	replication := NewReplication(localNode, predecessorList, 2)

	now := time.Now()
	var data []*tree.Row
	data = append(data, tree.MakeRow(now.Add(-10*time.Minute), []byte("a")))  // 86f7e437faa5 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-20*time.Minute), []byte("b")))  // e9d71f5e E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-30*time.Minute), []byte("c")))  // 84a51684 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-40*time.Minute), []byte("d")))  // 3c3638 E (4b107076bf0, f88860c3ae) => skip
	data = append(data, tree.MakeRow(now.Add(-50*time.Minute), []byte("e")))  // 58e6b3a4 E (4b107076bf0, f88860c3ae) => will be added to block
	data = append(data, tree.MakeRow(now.Add(-120*time.Minute), []byte("g"))) // 54fd17112 E (4b107076bf0, f88860c3ae) => will be added to block
	replication.MakeTrees(data)

	if len(replication.masterBlocks) != 1 {
		t.Errorf("number of master blocks must be %d got %d", 1, len(replication.masterBlocks))
	}
	if len(replication.masterBlocks[0].rows) != 5 {
		t.Errorf("number of master blocks rows must be %d got %d", 5, len(replication.masterBlocks[0].rows))
	}

	// check if master block first row hash is the same as data hash
	rowHashA := replication.masterBlocks[0].rows[0].Hash
	rowHashB := replication.masterBlocks[0].rows[1].Hash
	rowHashC := replication.masterBlocks[0].rows[2].Hash
	rowHashE := replication.masterBlocks[0].rows[3].Hash
	rowHashG := replication.masterBlocks[0].rows[4].Hash
	if !helpers.Equal(rowHashA, helpers.Hash("a")) {
		t.Errorf("row hash must be %x got %x", rowHashA, helpers.Hash("a"))
	}
	if !helpers.Equal(rowHashB, helpers.Hash("b")) {
		t.Errorf("row hash must be %x got %x", rowHashB, helpers.Hash("b"))
	}
	if !helpers.Equal(rowHashC, helpers.Hash("c")) {
		t.Errorf("row hash must be %x got %x", rowHashC, helpers.Hash("c"))
	}
	if !helpers.Equal(rowHashE, helpers.Hash("e")) {
		t.Errorf("row hash must be %x got %x", rowHashE, helpers.Hash("e"))
	}
	if !helpers.Equal(rowHashG, helpers.Hash("g")) {
		t.Errorf("row hash must be %x got %x", rowHashG, helpers.Hash("g"))
	}

	// check number of merkle trees generated
	if len(replication.trees) != 1 {
		t.Errorf("number of trees must be %d got %d", 1, len(replication.trees))
	}

	// check number of merkle tree nodes generated
	// we use 5 rows to make tree, and each row will go to a different block so we would have 5 blocks
	blocks := replication.trees[0].GetBlocks()
	if len(blocks) != 5 {
		t.Errorf("number of blocks in tree must be %d got %d", 5, len(blocks))
	}
	// number of leaf nodes is (5 block hash + 1 last duplicate + 3 branch hash + 1 duplicate branch + 2 parent branches + root hash) = 12
	if len(replication.trees[0].GetNodes()) != 12 {
		t.Errorf("number of tree nodes must be %d got %d", 12, len(replication.trees[0].GetNodes()))
	}

	var rootHash [helpers.HashSize]byte
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[0].Root[:]...)))
	if !helpers.Equal(rootHash, replication.RootHash) {
		t.Errorf("root hash must be %x got %x", rootHash, replication.RootHash)
	}
}

func TestMultipleMasterBlockMultipleSubBlocksOdd(t *testing.T) {
	localNode := NewNode("127.0.0.1", 10014) //  f88860c3ae
	predecessorList := NewPredecessorList()
	remoteSender := MockRemoteNodeSenderInterface{}
	// as replica is 2 we only need 1 predecessor
	predecessorList.Nodes[0] = NewRemoteNode(NewNode("127.0.0.1", 10015), remoteSender) // 4b107076bf0
	predecessorList.Nodes[1] = NewRemoteNode(NewNode("127.0.0.1", 10018), remoteSender) // 0060657c273
	replication := NewReplication(localNode, predecessorList, 3)

	now := time.Now()
	var data []*tree.Row
	data = append(data, tree.MakeRow(now.Add(-10*time.Minute), []byte("a")))  // 86f7e437faa5 E (4b107076bf0, f88860c3ae) => will be added to block 1
	data = append(data, tree.MakeRow(now.Add(-20*time.Minute), []byte("b")))  // e9d71f5e E (4b107076bf0, f88860c3ae) => will be added to block 1
	data = append(data, tree.MakeRow(now.Add(-30*time.Minute), []byte("c")))  // 84a51684 E (4b107076bf0, f88860c3ae) => will be added to block 1
	data = append(data, tree.MakeRow(now.Add(-40*time.Minute), []byte("d")))  // 3c3638 E (4b107076bf0, f88860c3ae) => will be added to block 0
	data = append(data, tree.MakeRow(now.Add(-50*time.Minute), []byte("e")))  // 58e6b3a4 E (4b107076bf0, f88860c3ae) => will be added to block 1
	data = append(data, tree.MakeRow(now.Add(-120*time.Minute), []byte("g"))) // 54fd17112 E (4b107076bf0, f88860c3ae) => will be added to block 1
	replication.MakeTrees(data)

	if len(replication.masterBlocks) != 2 {
		t.Errorf("number of master blocks must be %d got %d", 2, len(replication.masterBlocks))
	}
	if len(replication.masterBlocks[0].rows) != 1 {
		t.Errorf("number of master blocks rows must be %d got %d", 1, len(replication.masterBlocks[0].rows))
	}
	if len(replication.masterBlocks[1].rows) != 5 {
		t.Errorf("number of master blocks rows must be %d got %d", 5, len(replication.masterBlocks[1].rows))
	}

	// check if master block first row hash is the same as data hash
	rowHashD := replication.masterBlocks[0].rows[0].Hash

	rowHashA := replication.masterBlocks[1].rows[0].Hash
	rowHashB := replication.masterBlocks[1].rows[1].Hash
	rowHashC := replication.masterBlocks[1].rows[2].Hash
	rowHashE := replication.masterBlocks[1].rows[3].Hash
	rowHashG := replication.masterBlocks[1].rows[4].Hash
	if !helpers.Equal(rowHashD, helpers.Hash("d")) {
		t.Errorf("row hash must be %x got %x", rowHashD, helpers.Hash("d"))
	}
	if !helpers.Equal(rowHashA, helpers.Hash("a")) {
		t.Errorf("row hash must be %x got %x", rowHashA, helpers.Hash("a"))
	}
	if !helpers.Equal(rowHashB, helpers.Hash("b")) {
		t.Errorf("row hash must be %x got %x", rowHashB, helpers.Hash("b"))
	}
	if !helpers.Equal(rowHashC, helpers.Hash("c")) {
		t.Errorf("row hash must be %x got %x", rowHashC, helpers.Hash("c"))
	}
	if !helpers.Equal(rowHashE, helpers.Hash("e")) {
		t.Errorf("row hash must be %x got %x", rowHashE, helpers.Hash("e"))
	}
	if !helpers.Equal(rowHashG, helpers.Hash("g")) {
		t.Errorf("row hash must be %x got %x", rowHashG, helpers.Hash("g"))
	}

	// check number of merkle trees generated
	if len(replication.trees) != 2 {
		t.Errorf("number of trees must be %d got %d", 2, len(replication.trees))
	}

	// check number of merkle tree nodes generated
	// we use 5 rows to make tree, and each row will go to a different block so we would have 5 blocks
	blocksA := replication.trees[0].GetBlocks()
	if len(blocksA) != 1 {
		t.Errorf("number of blocks in tree must be %d got %d", 1, len(blocksA))
	}
	blocksB := replication.trees[1].GetBlocks()
	if len(blocksB) != 5 {
		t.Errorf("number of blocks in tree must be %d got %d", 5, len(blocksB))
	}

	// number of leaf nodes is (1 block hash + 1 last duplicate  + root hash) = 3
	if len(replication.trees[0].GetNodes()) != 3 {
		t.Errorf("number of tree nodes must be %d got %d", 3, len(replication.trees[0].GetNodes()))
	}
	// number of leaf nodes is (5 block hash + 1 last duplicate + 3 branch hash + 1 duplicate branch + 2 parent branches + root hash) = 12
	if len(replication.trees[1].GetNodes()) != 12 {
		t.Errorf("number of tree nodes must be %d got %d", 12, len(replication.trees[1].GetNodes()))
	}

	var rootHash [helpers.HashSize]byte
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[0].Root[:]...)))
	rootHash = helpers.Hash(string(append(rootHash[:], replication.trees[1].Root[:]...)))
	if !helpers.Equal(rootHash, replication.RootHash) {
		t.Errorf("root hash must be %x got %x", rootHash, replication.RootHash)
	}
}
