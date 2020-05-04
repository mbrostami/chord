package chord

import (
	"crypto/sha256"
	"net"
	"strconv"

	"github.com/mbrostami/chord/pkg/helpers"
)

// Node node entity
type Node struct {
	Identifier [sha256.Size]byte // hashed value of node
	IP         string
	Port       int
}

// NewNode creates new node and identifier
func NewNode(ip string, port int) *Node {
	newNode := &Node{}
	newNode.IP = ip
	newNode.Port = port
	newNode.Identifier = helpers.Hash(ip, port)
	return newNode
}

// SuccessorList successor list
type SuccessorList struct {
	Nodes map[int]*Node
}

// NewSuccessorList make new successor list
func NewSuccessorList() *SuccessorList {
	return &SuccessorList{
		Nodes: make(map[int]*Node),
	}
}

// FullAddr returns ip:port as string
func (n *Node) FullAddr() string {
	return net.JoinHostPort(n.IP, strconv.FormatInt(int64(n.Port), 10))
}
