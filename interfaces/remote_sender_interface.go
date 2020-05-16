package interfaces

//go:generate moq -out node_remote_sender_interface_test.go . RemoteSenderInterface
import (
	"github.com/mbrostami/chord"
)

// RemoteSenderInterface interface for client adapter
type RemoteSenderInterface interface {
	// FindSuccessor find the closest node to the given identifier
	// ref D
	FindSuccessor(remote *chord.RemoteNode, identifier []byte) (*chord.Node, error)

	// GetStablizerData successor's (successor list and predecessor)
	// to prevent duplicate rpc call, we get both together
	// ref E.3
	GetStablizerData(remote *chord.RemoteNode, caller *chord.Node) (*chord.Node, *chord.SuccessorList, error)

	// Notify update predecessor
	// is being called periodically by predecessor or new node
	// ref E.1
	Notify(remote *chord.RemoteNode, caller *chord.Node) error

	// Ping check if remote port is open - using to check predecessor state
	// FIXME should be cached
	// ref E.1
	Ping(remote *chord.RemoteNode) bool
}
