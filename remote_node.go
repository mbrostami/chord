package chord

import (
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
)

type RemoteNode struct {
	*Node
	sender RemoteNodeSenderInterface
}

func NewRemoteNode(localNode *Node, remoteSender RemoteNodeSenderInterface) *RemoteNode {
	remoteNode := &RemoteNode{
		Node:   localNode,
		sender: remoteSender,
	}
	return remoteNode
}

func (n *RemoteNode) FindSuccessor(identifier [helpers.HashSize]byte) (*RemoteNode, error) {
	node, err := n.sender.FindSuccessor(n, identifier)
	if err == nil {
		// log.Infof("got successor from remote node: %x succ() => %x\n", identifier, node.Identifier)
	}
	return NewRemoteNode(node, n.sender), err
}

// GetStablizerData successor's (successor list and predecessor)
// to prevent duplicate rpc call, we get both together
// ref E.3
func (n *RemoteNode) GetStablizerData(local *Node) (*RemoteNode, *SuccessorList, error) {
	node, successorList, err := n.sender.GetStablizerData(n, local)
	return NewRemoteNode(node, n.sender), successorList, err
}

// GetPredecessorList predecessor's (predecessor list)
func (n *RemoteNode) GetPredecessorList(local *Node) (*PredecessorList, error) {
	predecessorList, err := n.sender.GetPredecessorList(n, local)
	return predecessorList, err
}

// Store store data on remote node
func (n *RemoteNode) Store(data []byte) bool {
	return n.sender.Store(n, data)
}

// ForwardSync sync local data missing on remote
func (n *RemoteNode) ForwardSync(plHash [helpers.HashSize]byte, data []byte, tree *merkle.MerkleTree) (*merkle.MerkleTree, error) {
	return n.sender.ForwardSync(n, plHash, data, tree)
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (n *RemoteNode) Notify(local *Node) error {
	return n.sender.Notify(n, local)
}

// Ping check if remote port is open - using to check predecessor state
// FIXME should be cached
// ref E.1
func (n *RemoteNode) Ping() bool {
	return n.sender.Ping(n)
}

func (n *RemoteNode) GlobalMaintenance(data []byte) ([]byte, error) {
	return n.sender.GlobalMaintenance(n, data)
}
