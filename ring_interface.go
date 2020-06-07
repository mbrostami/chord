package chord

import (
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
)

//go:generate moq -out ring_interface_test.go . RingInterface

// RingInterface interface for chord ring
type RingInterface interface {

	// Join joins a node to the network through remoteNode
	Join(remoteNode *RemoteNode) error

	// Verbose prints information about ring
	Verbose()

	// FixFingers fixes finger table periodically
	FixFingers()

	// CheckPredecessor check predecessor if it's not available periodically
	CheckPredecessor()

	// Stabilize checks successor if it's available, also updates successor list periodically
	Stabilize()

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
	GetStabilizerData(caller *Node) (predecessor *RemoteNode, successorList *SuccessorList)

	// GetPredecessorList predecessor's (predecessor list)
	GetPredecessorList(caller *Node) (predecessorList *PredecessorList)

	SyncData() error
	GlobalMaintenance(data []byte) ([]byte, error)

	// Store
	// is being called periodically by predecessor or new node
	// ref E.1
	Store(data []byte) bool

	ForwardSync(newData []byte, predecessorListHash [helpers.HashSize]byte, serializedData []*merkle.SerializedNode) ([]*merkle.SerializedNode, error)
}
