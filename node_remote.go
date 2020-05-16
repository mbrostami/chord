package chord

import "github.com/mbrostami/chord/interfaces"

type RemoteNode struct {
	node         *Node
	sender *interfaces.RemoteSenderInterface
}

func NewRemoteNode(ip string, port uint, remoteSender *interfaces.RemoteSenderInterface) *RemoteNode {
	node := NewNode(ip, port)
	remoteNode := &RemoteNode{
		node:         node,
		sender: remoteSender,
	}
	return remoteNode
}

func (n *RemoteNode) GetIP() string {
	return n.node.GetIP()
}

func (n *RemoteNode) GetPort() uint {
	return n.node.GetPort()
}

func (n *RemoteNode) GetFullAddress() string {
	return n.node.GetFullAddress()
}

func (n *RemoteNode) GetIdentity() [HashSize]byte {
	return n.node.GetIdentity()
}

func (n *RemoteNode) FindSuccessor(identifier [HashSize]byte) (*Node, error) {
	return n.sender.FindSuccessor(n, identifier)
}
