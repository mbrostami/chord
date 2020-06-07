package chord

import (
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
)

// MockRemoteNodeSenderInterface interface for client adapter
type MockRemoteNodeSenderInterface struct {
}

func (m MockRemoteNodeSenderInterface) FindSuccessor(remote *RemoteNode, identifier [helpers.HashSize]byte) (*Node, error) {
	return nil, nil
}
func (m MockRemoteNodeSenderInterface) GetStablizerData(remote *RemoteNode, local *Node) (*Node, *SuccessorList, error) {
	return nil, nil, nil
}
func (m MockRemoteNodeSenderInterface) Notify(remote *RemoteNode, local *Node) error {
	return nil
}
func (m MockRemoteNodeSenderInterface) Ping(remote *RemoteNode) bool {
	return true
}
func (m MockRemoteNodeSenderInterface) Store(remote *RemoteNode, data []byte) bool {
	return true
}
func (m MockRemoteNodeSenderInterface) ForwardSync(remote *RemoteNode, plHash [helpers.HashSize]byte, data []byte, tree *merkle.MerkleTree) (*merkle.MerkleTree, error) {
	return nil, nil
}
func (m MockRemoteNodeSenderInterface) GetPredecessorList(remote *RemoteNode, local *Node) (*PredecessorList, error) {
	return nil, nil
}
func (m MockRemoteNodeSenderInterface) GlobalMaintenance(remote *RemoteNode, data []byte) ([]byte, error) {
	return nil, nil
}
