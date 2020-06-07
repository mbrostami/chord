package chord

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/tree"
)

const REPLICAS = 2

type Replication struct {
	SourceTime   time.Time                      `json:"source_time"`
	Ranges       map[int][helpers.HashSize]byte `json:"ranges"`
	Trees        map[int]*tree.Merkle           `json:"trees"`
	RootHash     [helpers.HashSize]byte         `json:"master_root"`
	MasterBlocks map[int]*MasterBlock
	replicas     int
}

type MasterBlock struct {
	ID       int                    `json:"id"`
	RootHash [helpers.HashSize]byte `json:"root"`
	min      [helpers.HashSize]byte
	max      [helpers.HashSize]byte
	rows     []*tree.Row
}

func NewReplication(sourceTime time.Time, ranges map[int][helpers.HashSize]byte, replicas int) *Replication {
	if replicas < 2 {
		replicas = REPLICAS
	}
	return &Replication{
		Ranges:       ranges,
		MasterBlocks: make(map[int]*MasterBlock),
		Trees:        make(map[int]*tree.Merkle),
		replicas:     replicas,
		SourceTime:   sourceTime,
	}
}
func (r *Replication) MakeTreesWithData(data map[[helpers.HashSize]byte]*Record) error {
	// sort data
	sortedKeys := make([]string, 0, len(data))
	for k := range data {
		sortedKeys = append(sortedKeys, string(k[:]))
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		var key [helpers.HashSize]byte
		copy(key[:helpers.HashSize], []byte(k)[:helpers.HashSize])

		blockID, found, err := r.findMasterBlockNumber(key)
		if err != nil {
			fmt.Printf("error range %v\n", err)
		}
		if found == true {
			r.MasterBlocks[blockID].appendRow(tree.MakeRow(data[key].CreationTime, data[key].GetJson(), data[key].Identifier))
		}
	}
	for i := 0; i < len(r.MasterBlocks); i++ {
		tree := tree.MakeMerkleWithTime(r.MasterBlocks[i].rows, r.SourceTime)
		r.MasterBlocks[i].RootHash = tree.Root
		r.Trees[i] = tree
		// calculate master root hash
		r.RootHash = helpers.Hash(string(append(r.RootHash[:], tree.Root[:]...)))
	}
	return nil
}

func (r *Replication) MakeTrees(data []*tree.Row) error {
	for i := 0; i < len(data); i++ {
		key := data[i].Hash
		blockID, found, err := r.findMasterBlockNumber(key)
		if err != nil {
			fmt.Printf("error range %v\n", err)
		}
		if found == true {
			r.MasterBlocks[blockID].appendRow(data[i])
		}
	}
	for i := 0; i < len(r.MasterBlocks); i++ {
		tree := tree.MakeMerkleWithTime(r.MasterBlocks[i].rows, r.SourceTime)
		r.MasterBlocks[i].RootHash = tree.Root
		r.Trees[i] = tree
		// calculate master root hash
		r.RootHash = helpers.Hash(string(append(r.RootHash[:], tree.Root[:]...)))
	}
	return nil
}

// findMasterBlockNumber
// if replica = 2 then we need to find the keys which are between predecessor[0] and current node
// to replicate in successor
// if replica = 3 then (predecessor[1] < keys <= predecessor[0]) are in first master block
// and (predecessor[0] < keys <= current node) are in the next master block
func (r *Replication) findMasterBlockNumber(key [helpers.HashSize]byte) (int, bool, error) {
	masterBlockNumber := 0
	for i := len(r.Ranges) - 1; i > 0; i-- {
		minHash := r.Ranges[i]
		nextHash := r.Ranges[i-1]
		// fmt.Printf("predecessor %d -> %x\n", i, r.predecessorList.Nodes[i].Identifier)
		// fmt.Printf("check %x E (%x, %x) %v\n", key, minHash, nextHash, helpers.BetweenR(key, minHash, nextHash))
		if helpers.BetweenR(key, minHash, nextHash) {
			if r.MasterBlocks[masterBlockNumber] == nil { // if masterblock is not created yet
				masterBlock := &MasterBlock{
					min: minHash,
					max: nextHash,
					ID:  masterBlockNumber,
				}
				r.MasterBlocks[masterBlockNumber] = masterBlock
			}
			return masterBlockNumber, true, nil
		}
		masterBlockNumber++
	}
	return masterBlockNumber, false, nil
}

func (r *Replication) FindMissingData(diffs *Diffs) ([]*tree.Row, error) {
	var rowsToReplicate []*tree.Row
	if diffs.RootHash == r.RootHash {
		return rowsToReplicate, nil
	}
	for mblockID, mblock := range r.MasterBlocks {
		if diffs.MasterBlocks[mblockID] == nil {
			// fmt.Print("FindMissingData: root hash is empty!\n")
			rowsToReplicate = append(rowsToReplicate, mblock.rows...)
			continue
		}
		// if master blocks are same, go to the next one
		if diffs.MasterBlocks[mblockID].RootHash == mblock.RootHash {
			// fmt.Print("FindMissingData: master block hashes are same!\n")
			continue
		}
		// if diffs.masterBlock[mblockID] exists there must be diffs.Trees[mblockID]
		diffTreeNodes := diffs.Trees[mblockID].Nodes
		treeNodes := r.Trees[mblockID].Nodes

		// if number of nodes in local tree, is not the same as number of nodes in diff
		if len(treeNodes) != len(diffTreeNodes) {
			// FIXME must optimize this
			// check leaf nodes
			for i := len(treeNodes) - 1; i >= 0; i-- {
				if i >= len(diffTreeNodes) {
					if treeNodes[i].Level == 0 {
						blocks := r.Trees[mblockID].Blocks
						// find the block with same hash, and get rows
						for _, block := range blocks {
							if *block.Hash == treeNodes[i].Hash {
								// fmt.Print("FindMissingData: diff node is empty!\n")
								rowsToReplicate = append(rowsToReplicate, block.Rows...)
								break
							}
						}
					}
					continue
				}
				if treeNodes[i].Hash != diffTreeNodes[i].Hash {
					// only leaf nodes considered as diffBlockNodes
					if treeNodes[i].Level == 0 {
						blocks := r.Trees[mblockID].Blocks
						// find the block with same hash, and get rows
						for _, block := range blocks {
							if *block.Hash == treeNodes[i].Hash {
								// fmt.Print("FindMissingData: diff node hash is not same!\n")
								rowsToReplicate = append(rowsToReplicate, block.Rows...)
								break
							}
						}
					}
				}
			}
		} else {
			diffRootHash := diffTreeNodes[len(diffTreeNodes)-1].Hash
			rootHash := treeNodes[len(treeNodes)-1].Hash
			// if root hashes are not same, check children
			if diffRootHash != rootHash {
				for i := len(treeNodes) - 1; i >= 0; i-- {
					if treeNodes[i].Hash != diffTreeNodes[i].Hash {
						// only leaf nodes considered as diffBlockNodes
						if diffTreeNodes[i].Level == 0 {
							blocks := r.Trees[mblockID].Blocks
							// find the block with same hash, and get rows
							for _, block := range blocks {
								if *block.Hash == treeNodes[i].Hash {
									// fmt.Print("FindMissingData: same len: diff node hash is not same!\n")
									rowsToReplicate = append(rowsToReplicate, block.Rows...)
									break
								}
							}
						}
					}
				}
			}
		}
	}
	return rowsToReplicate, nil
}

func (r *Replication) FindDiffs(basicTransport BasicTranport) ([]byte, error) {
	if r.RootHash == basicTransport.RootHash {
		// means everything is synced
		return nil, nil
	}
	hasDiff := false
	diffs := make(map[int]*tree.Merkle)
	for mblockID, mblock := range basicTransport.MasterBlocks {
		if r.MasterBlocks[mblockID] == nil {
			hasDiff = true
			continue
		}
		if mblock.RootHash != r.MasterBlocks[mblockID].RootHash {
			diffs[mblockID] = r.Trees[mblockID]
			hasDiff = true
		}
	}
	if hasDiff == false {
		return nil, nil
	}
	// if there is at least one diff, then return basic serialize
	return Serialize(r, diffs)
}

type Diffs struct {
	SourceTime   time.Time                      `json:"time"`
	RootHash     [helpers.HashSize]byte         `json:"master_root"`
	MasterBlocks map[int]*MasterBlock           `json:"master_blocks"`
	Ranges       map[int][helpers.HashSize]byte `json:"ranges"`
	Trees        map[int]*tree.Merkle           `json:"trees"`
}

type BasicTranport struct {
	SourceTime   time.Time                      `json:"time"`
	RootHash     [helpers.HashSize]byte         `json:"master_root"`
	MasterBlocks map[int]*MasterBlock           `json:"master_blocks"`
	Ranges       map[int][helpers.HashSize]byte `json:"ranges"`
}

// Serialize make json
func Serialize(r *Replication, trees map[int]*tree.Merkle) ([]byte, error) {
	transport := Diffs{
		SourceTime:   r.SourceTime,
		RootHash:     r.RootHash,
		MasterBlocks: r.MasterBlocks,
		Ranges:       r.Ranges,
		Trees:        trees,
	}
	// fmt.Printf("Serialize: diffs %+v\n", transport)
	return json.Marshal(transport)
}

// Unserialize make struct out of json
func Unserialize(jsonData []byte) *Diffs {
	diffs := Diffs{}
	json.Unmarshal(jsonData, &diffs)
	// fmt.Printf("Unserialize: diffs %+v\n", diffs)
	return &diffs
}

// BasicSerialize make json
func BasicSerialize(r *Replication) ([]byte, error) {
	basicTransport := BasicTranport{
		SourceTime:   r.SourceTime,
		RootHash:     r.RootHash,
		MasterBlocks: r.MasterBlocks,
		Ranges:       r.Ranges,
	}
	return json.Marshal(basicTransport)
}

// BasicUnserialize make struct out of json
func BasicUnserialize(jsonData []byte) *BasicTranport {
	basicTransport := BasicTranport{}
	json.Unmarshal(jsonData, &basicTransport)
	// fmt.Printf("BasicUnserialize: data %+v\n", basicTransport)
	return &basicTransport
}

func (mb *MasterBlock) appendRow(row *tree.Row) {
	mb.rows = append(mb.rows, row)
}
