package chord

import (
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
)

//go:generate moq -out remote_node_sender_interface_test.go . RemoteNodeSenderInterface

// RemoteNodeSenderInterface interface for client adapter
type RemoteNodeSenderInterface interface {
	// FindSuccessor find the closest node to the given identifier
	// ref D
	FindSuccessor(remote *RemoteNode, identifier [helpers.HashSize]byte) (*Node, error)

	// GetStablizerData successor's (successor list and predecessor)
	// to prevent duplicate rpc call, we get both together
	// ref E.3
	GetStablizerData(remote *RemoteNode, local *Node) (*Node, *SuccessorList, error)

	// Notify update predecessor
	// is being called periodically by predecessor or new node
	// ref E.1
	Notify(remote *RemoteNode, local *Node) error

	// Ping check if remote port is open - using to check predecessor state
	// FIXME should be cached
	// ref E.1
	Ping(remote *RemoteNode) bool

	GlobalMaintenance(remote *RemoteNode, data []byte) ([]byte, error)

	// Store store data in remote node
	Store(remote *RemoteNode, data []byte) bool

	// Store store data in remote node
	ForwardSync(remote *RemoteNode, plHash [helpers.HashSize]byte, data []byte, tree *merkle.MerkleTree) (*merkle.MerkleTree, error)

	// GetPredecessorList
	GetPredecessorList(remote *RemoteNode, local *Node) (*PredecessorList, error)
}
