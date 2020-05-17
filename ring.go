package chord

import (
	"fmt"

	"github.com/mbrostami/chord/helpers"
)

type Ring struct {
	localNode     *Node
	remoteSender  RemoteNodeSenderInterface
	fingerTable   *FingerTable
	successorList *SuccessorList
	stabilizer    *Stabilizer
	Predecessor   *RemoteNode
	Successor     *RemoteNode
}

func NewRing(localNode *Node, remoteSender RemoteNodeSenderInterface) *RingInterface {
	var ring RingInterface
	successorList := NewSuccessorList()
	ring = &Ring{
		localNode:     localNode,
		remoteSender:  remoteSender,
		fingerTable:   &FingerTable{},
		stabilizer:    NewStabilizer(successorList),
		successorList: successorList,
	}
	return &ring
}

func (r *Ring) GetLocalNode() *Node {
	return r.localNode
}

func (r *Ring) FindSuccessor(identifier [helpers.HashSize]byte) *RemoteNode {
	// fmt.Printf("FindSuccessor: start looking for key %x \n", identifier)
	if r.Successor.Identifier == r.localNode.Identifier {
		return NewRemoteNode(r.localNode, r.remoteSender)
	}
	// id ∈ (n, successor]
	if helpers.BetweenR(identifier, r.localNode.Identifier, r.Successor.Identifier) {
		return r.Successor
	}
	closestRemoteNode := r.fingerTable.ClosestPrecedingNode(identifier, r.localNode)
	successorListClosestNode := r.successorList.ClosestPrecedingNode(identifier, r.localNode, closestRemoteNode)
	if successorListClosestNode != nil {
		closestRemoteNode = successorListClosestNode
	} else if closestRemoteNode == nil {
		closestRemoteNode = NewRemoteNode(r.localNode, r.remoteSender) // make a copy local node as remote node
	}
	if closestRemoteNode.Identifier == r.localNode.Identifier { // current node is the only node in figer table
		return closestRemoteNode // return local node
	}
	nextNodeSuccessor, err := closestRemoteNode.FindSuccessor(identifier)
	if err != nil { // unexpected error on successor
		fmt.Printf("Unexpected error from successor %v", err)
		return nil
	}
	return nextNodeSuccessor
}

// Stabilize keep successor and predecessor updated
// Runs periodically
// ref E.1 - E.3
func (r *Ring) Stabilize() {
	if len(r.successorList.Nodes) == 0 { // if no successor available
		r.Successor = NewRemoteNode(r.localNode, r.remoteSender)
		r.fingerTable.Set(1, r.Successor)
	} else {
		successor := r.stabilizer.Start(r.Successor, r.localNode)
		// If successor is changed while stabilizing
		if successor.Identifier != r.Successor.Identifier {
			r.Successor = successor
			r.fingerTable.Set(1, r.Successor)
			// immediatly update new successor about it's new predecessor
			r.Successor.Notify(r.localNode)
		}
	}
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (r *Ring) Notify(caller *Node) bool {
	// (c.predecessor is nil or node ∈ (c.predecessor, n))
	if r.Predecessor == nil {
		r.Predecessor = NewRemoteNode(caller, r.remoteSender)
		// First node in the network has itself as successor
		if r.Successor.Identifier == r.localNode.Identifier {
			r.Successor = r.Predecessor
			r.Successor.Notify(r.localNode)
		}
		return true
	}
	if helpers.Between(caller.Identifier, r.Predecessor.Identifier, r.localNode.Identifier) {
		r.Predecessor = NewRemoteNode(caller, r.remoteSender)
		return true
	}
	return false
}

func (r *Ring) CheckPredecessor(caller *RemoteNode) {
	if r.Predecessor != nil {
		if !r.Predecessor.Ping() {
			r.Predecessor = nil // set nil to be able to update predecessor by notify
		}
	}
}

// FixFingers refreshes finger table entities
// Runs periodically
// ref D - E.1 - finger[k] = (n + 2 ** k-1) Mod M
func (r *Ring) FixFingers() {
	if r.Successor == nil {
		return
	}
	index, identifier := r.fingerTable.CalculateIdentifier(r.localNode)
	remoteNode := r.FindSuccessor(identifier)
	r.fingerTable.Set(index, remoteNode)
	if index == 1 && remoteNode.Identifier != r.Successor.Identifier { // means it's first entry of fingerTable (first entry should be always the next successor of current node)
		r.Successor = remoteNode
		// immediatly update new successor about it's new predecessor
		r.Successor.Notify(r.localNode)
	}
}

// GetSuccessorList returns unsorted successor list
// ref E.3
func (r *Ring) GetSuccessorList() *SuccessorList {
	return r.successorList
}

// GetStabilizerData return predecessor and successor list
func (r *Ring) GetStabilizerData(caller *Node) (*RemoteNode, *SuccessorList) {
	remoteCaller := NewRemoteNode(caller, r.remoteSender)
	predecessor := r.getPredecessor(remoteCaller)
	return predecessor, r.GetSuccessorList()
}

func (r *Ring) getPredecessor(caller *RemoteNode) *RemoteNode {
	if r.Predecessor != nil {
		// extension on chord
		if helpers.Between(caller.Identifier, r.Predecessor.Identifier, r.localNode.Identifier) {
			r.Predecessor = caller // update predecessor if caller is closer than predecessor
		}
		return r.Predecessor
	}
	return NewRemoteNode(r.localNode, r.remoteSender)
}
