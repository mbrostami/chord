package chord

import (
	"sync"

	"github.com/mbrostami/chord/helpers"
)

// RSIZE is the number of records in successor list
// could be (log n) refer to Theorem IV.5
// Increasing r makes the system more robust
// ref E.3 - Theorem IV.5
const RSIZE int = helpers.HashSize // 20

// SuccessorList successor list
type SuccessorList struct {
	Nodes map[int]*RemoteNode
	mutex sync.RWMutex
	r     int
}

// NewSuccessorList make new successor list
func NewSuccessorList() *SuccessorList {
	return &SuccessorList{
		Nodes: make(map[int]*RemoteNode),
		r:     RSIZE,
	}
}

func (sl *SuccessorList) ClosestPrecedingNode(identifier [helpers.HashSize]byte, localNode *Node, fingerTableClosest *RemoteNode) *RemoteNode {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()
	// ref E.3 - modified version of closestPrecedingNode will also check the successor list
	// to find closest node to the identifier
	for i := 0; i < len(sl.Nodes); i++ {
		// successorList[i] ∈ (fingerTable[k], id)
		// if successor item is closer to identifier than the found fingertable item, then return successor item
		if fingerTableClosest != nil {
			if helpers.BetweenR(sl.Nodes[i].Identifier, fingerTableClosest.Identifier, identifier) {
				return sl.Nodes[i]
			}
		} else {
			// successorList[i] ∈ (n, id)
			if helpers.BetweenR(sl.Nodes[i].Identifier, localNode.Identifier, identifier) {
				return sl.Nodes[i]
			}
		}
	}
	return nil
}

// UpdateSuccessorList updates successor list - ref E.3
func (sl *SuccessorList) UpdateSuccessorList(successor *RemoteNode, localNode *Node, successorList *SuccessorList) {
	if successorList == nil || successor == nil {
		return
	}
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	sl.Nodes = make(map[int]*RemoteNode) // reset nodes
	sl.Nodes[0] = successor              // replace first item with successor itself
	index := 1
	for i := 0; i < len(successorList.Nodes); i++ {
		if len(sl.Nodes) >= sl.r { // prevent overloading successorlist (max(r)=(log N)) ref E.3
			break
		}
		chorNode := successorList.Nodes[i]
		// ignore same nodes
		if chorNode.Identifier == localNode.Identifier {
			continue
		}
		sl.Nodes[index] = chorNode
		index++
	}
}
