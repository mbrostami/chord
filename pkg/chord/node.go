package chord

import (
	"crypto/sha256"
	"net"
	"strconv"

	pb "github.com/mbrostami/chord/internal/grpc"
)

// Node node entity
type Node struct {
	Identifier [sha256.Size]byte // hashed value of node
	IP         string
	Port       int
}

func (n *Node) GrpcNode() *pb.Node {
	node := &pb.Node{
		IP:   n.IP,
		Port: int32(n.Port),
	}
	return node
}

func (n *Node) FullAddr() string {
	return net.JoinHostPort(n.IP, strconv.FormatInt(int64(n.Port), 10))
}
