package chord

//go:generate moq -out clientinterface_test.go . ClientInterface

// ClientInterface interface for client adapter
type ClientInterface interface {

	// FindSuccessor find the closest node to the given identifier
	// ref D
	FindSuccessor(remote *Node, identifier []byte) (*Node, error)

	// GetStablizerData successor's (successor list and predecessor)
	// to prevent duplicate rpc call, we get both together
	// ref E.3
	GetStablizerData(remote *Node, node *Node) (*Node, *SuccessorList, error)

	// Notify update predecessor
	// is being called periodically by predecessor or new node
	// ref E.1
	Notify(remote *Node, node *Node) error

	// Ping check if remote port is open - using to check predecessor state
	// FIXME should be cached
	// ref E.1
	Ping(remote *Node) bool
}
