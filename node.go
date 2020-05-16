package chord

import (
	"net"
	"strconv"
)

type Node struct {
	Identifier [HashSize]byte
	IP         string
	Port       uint
}

func NewNode(ip string, port uint) *Node {
	node := &Node{}
	node.IP = ip
	node.Port = port
	node.Identifier = Hash(ip + ":" + strconv.FormatInt(int64(port), 10))
	return node
}

func (n *Node) GetIP() string {
	return n.IP
}

func (n *Node) GetPort() uint {
	return n.Port
}

func (n *Node) GetFullAddress() string {
	return net.JoinHostPort(n.IP, strconv.FormatInt(int64(n.Port), 10))
}

func (n *Node) GetIdentity() [HashSize]byte {
	return n.Identifier
}
