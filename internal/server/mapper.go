package server

import (
	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/pkg/chord"
)

// ConvertToGrpcNode make grpc node entity from chord node
func ConvertToGrpcNode(node *chord.Node) *pb.Node {
	grpcNode := &pb.Node{
		IP:   node.IP,
		Port: int32(node.Port),
	}
	return grpcNode
}

// ConvertToChordNode make grpc node entity from chord node
func ConvertToChordNode(node *pb.Node) *chord.Node {
	return chord.NewNode(node.IP, int(node.Port))
}

// ConvertToGrpcSuccessorList make grpc node entity from chord node
func ConvertToGrpcSuccessorList(slist *chord.SuccessorList) []*pb.Node {
	nodes := []*pb.Node{}
	if slist != nil {
		for i := 0; i < len(slist.Nodes); i++ { // keep sorts
			nodes = append(nodes, ConvertToGrpcNode(slist.Nodes[i]))
		}
	}
	return nodes
}

// ConvertToChordSuccessorList make grpc node entity from chord node
func ConvertToChordSuccessorList(nlist []*pb.Node) *chord.SuccessorList {
	nodes := chord.NewSuccessorList()
	for i := 0; i < len(nlist); i++ { // keep sorted
		nodes.Nodes[i] = chord.NewNode(nlist[i].IP, int(nlist[i].Port))
	}
	return nodes
}
