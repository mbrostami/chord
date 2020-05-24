package chord

import (
	"sync"
)

// PredecessorList successor list
type PredecessorList struct {
	Nodes map[int]*RemoteNode
	mutex sync.RWMutex
	r     int
}

// NewPredecessorList make new successor list
func NewPredecessorList() *PredecessorList {
	return &PredecessorList{
		Nodes: make(map[int]*RemoteNode),
		r:     RSIZE,
	}
}

// UpdatePredecessorList updates predecessor list
func (pl *PredecessorList) UpdatePredecessorList(successor *RemoteNode, predecessor *RemoteNode, localNode *Node, predecessorList *PredecessorList) {
	if predecessorList == nil || predecessor == nil {
		return
	}
	pl.mutex.Lock()
	defer pl.mutex.Unlock()
	pl.Nodes = make(map[int]*RemoteNode) // reset nodes
	pl.Nodes[0] = predecessor            // replace first item with predecessor itself
	index := 1
	for i := 0; i < len(predecessorList.Nodes); i++ {
		if len(pl.Nodes) >= pl.r { // prevent overloading successorlist (max(r)=(log N)) ref E.3
			break
		}
		chorNode := predecessorList.Nodes[i]
		// ignore same nodes
		if chorNode.Identifier == localNode.Identifier {
			continue
		}
		// in small networks where the number of nodes are smaller than pl.r number
		// it's possible to have current node successor in predecessor's predecessorlist
		// to prevent this loop, we should check predecessorlist and be sure to remove those
		// in order to prevent a loop (including failed nodes)
		if successor != nil {
			if chorNode.Identifier == successor.Identifier {
				break // skip next records including current one
			}
		}
		pl.Nodes[index] = chorNode
		index++
	}

}
