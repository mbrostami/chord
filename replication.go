package chord

import (
	"errors"
	"fmt"

	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/tree"
)

const REPLICAS = 2

type Replication struct {
	replicas        int
	predecessorList *PredecessorList
	localNode       *Node
	masterBlocks    map[int]*MasterBlock
	trees           []*tree.Merkle
	RootHash        [helpers.HashSize]byte
}
type MasterBlock struct {
	Min  [helpers.HashSize]byte
	Max  [helpers.HashSize]byte
	ID   int
	rows []*tree.Row
}

func NewReplication(localNode *Node, predecessorList *PredecessorList, replicas int) *Replication {
	if replicas < 2 {
		replicas = REPLICAS
	}
	return &Replication{
		predecessorList: predecessorList,
		masterBlocks:    make(map[int]*MasterBlock),
		localNode:       localNode,
		replicas:        replicas,
	}
}

func (r *Replication) MakeTrees(data []*tree.Row) error {
	for i := 0; i < len(data); i++ {
		key := data[i].Hash
		blockID, found, err := r.findMasterBlockNumber(key)
		if err != nil {
			fmt.Printf("error range %v\n", err)
		}
		if found == true {
			r.masterBlocks[blockID].appendRow(data[i])
		}
	}
	for i := 0; i < len(r.masterBlocks); i++ {
		tree := tree.MakeMerkle(r.masterBlocks[i].rows)
		r.trees = append(r.trees, tree)
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
	for i := r.replicas - 2; i >= 0; i-- {
		if len(r.predecessorList.Nodes) <= i {
			return masterBlockNumber, false, errors.New("not enough predecessors")
		}
		minHash := r.predecessorList.Nodes[i].Identifier
		var nextHash [helpers.HashSize]byte
		if i == 0 {
			nextHash = r.localNode.Identifier
		} else {
			nextHash = r.predecessorList.Nodes[i-1].Identifier
		}
		// fmt.Printf("predecessor %d -> %x\n", i, r.predecessorList.Nodes[i].Identifier)
		// fmt.Printf("check %x E (%x, %x) %v\n", key, minHash, nextHash, helpers.BetweenR(key, minHash, nextHash))
		if helpers.BetweenR(key, minHash, nextHash) {
			if r.masterBlocks[masterBlockNumber] == nil { // if masterblock is not created yet
				masterBlock := &MasterBlock{
					Min: minHash,
					Max: nextHash,
					ID:  masterBlockNumber,
				}
				r.masterBlocks[masterBlockNumber] = masterBlock
			}
			return masterBlockNumber, true, nil
		}
		masterBlockNumber++
	}
	return masterBlockNumber, false, nil
}

func (mb *MasterBlock) appendRow(row *tree.Row) {
	mb.rows = append(mb.rows, row)
}
