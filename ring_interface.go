package chord

import "github.com/mbrostami/chord/helpers"

//go:generate moq -out ring_interface_test.go . RingInterface

// RingInterface interface for chord ring
type RingInterface interface {

	// GetLocalNode returns local node
	GetLocalNode() *Node

	// FindSuccessor find the closest node to the given identifier
	// ref D
	FindSuccessor(identifier [helpers.HashSize]byte) *RemoteNode

	// Notify update predecessor
	// is being called periodically by predecessor or new node
	// ref E.1
	Notify(caller *Node) bool

	// GetPredecessor return predecessor
	// to prevent duplicate rpc call, we get both together
	// ref E.3
	GetPredecessor(caller *RemoteNode) *RemoteNode

	// GetStabilizerData successor's (successor list and predecessor)
	// FIXME should be cached
	// ref E.1
	GetStabilizerData(caller *Node) (predecessor *Node, successorList *SuccessorList)
}
