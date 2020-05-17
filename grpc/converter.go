package grpc

import "github.com/mbrostami/chord"

// ConvertToGrpcNode convert chord node to grpc node
func ConvertToGrpcNode(node *chord.Node) *Node {
	grpcNode := &Node{
		IP:   node.IP,
		Port: int32(node.Port),
	}
	return grpcNode
}

// ConvertToChordNode change grpc node to chord local node
func ConvertToChordNode(node *Node) *chord.Node {
	return chord.NewNode(node.IP, uint(node.Port))
}

// ConvertToGrpcSuccessorList change chord successor list to grpc nodes
func ConvertToGrpcSuccessorList(slist *chord.SuccessorList) []*Node {
	nodes := []*Node{}
	if slist != nil {
		for i := 0; i < len(slist.Nodes); i++ { // keep sorts
			nodes = append(nodes, ConvertToGrpcNode(slist.Nodes[i].Node))
		}
	}
	return nodes
}

// ConvertToChordSuccessorList change grpc nodes to chord successor list
func ConvertToChordSuccessorList(nlist []*Node, remoteSender chord.RemoteNodeSenderInterface) *chord.SuccessorList {
	nodes := chord.NewSuccessorList()
	for i := 0; i < len(nlist); i++ { // keep sorted
		nodes.Nodes[i] = chord.NewRemoteNode(chord.NewNode(nlist[i].IP, uint(nlist[i].Port)), remoteSender)
	}
	return nodes
}
