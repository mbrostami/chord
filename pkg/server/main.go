package server

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/pkg/chord"
	"google.golang.org/grpc"
)

type ChordServer struct {
	pb.UnimplementedChordServer
	Node *chord.Node
}

// GetSuccessor get successor node
func (s *ChordServer) GetSuccessor(ctx context.Context, empty *empty.Empty) (*pb.Node, error) {
	node := &pb.Node{}
	if s.Node.Successor != nil {
		node.IP = s.Node.Successor.IP
		node.Key = s.Node.Successor.Identifier[:]
		node.Port = int32(s.Node.Successor.Port)
	} else {
		node.IP = s.Node.IP
		node.Key = s.Node.Identifier[:]
		node.Port = int32(s.Node.Port)
	}
	// fmt.Printf("grpc Server: GetSuccessor: %s:%d \n", node.IP, node.Port)
	return node, nil
}

// Notify update predecessor
// is being called periodically
func (s *ChordServer) Notify(ctx context.Context, n *pb.Node) (*wrappers.BoolValue, error) {
	node := chord.NewNode(n.IP, int(n.Port)) // make node struct and calculate identifier
	result := &wrappers.BoolValue{}
	result.Value = s.Node.Notify(node)
	return result, nil
}

// GetPredecessor get predecessor node
func (s *ChordServer) GetPredecessor(ctx context.Context, empty *empty.Empty) (*pb.Node, error) {
	node := &pb.Node{}
	if s.Node.Predecessor != nil {
		node.IP = s.Node.Predecessor.IP
		node.Key = s.Node.Predecessor.Identifier[:]
		node.Port = int32(s.Node.Predecessor.Port)
	} else {
		node.IP = s.Node.IP
		node.Key = s.Node.Identifier[:]
		node.Port = int32(s.Node.Port)
	}
	// fmt.Printf("grpc Server: GetPredecessor: %s:%d \n", node.IP, node.Port)
	return node, nil
}

// FindSuccessor get closest node to the given key
func (s *ChordServer) FindSuccessor(ctx context.Context, lookup *pb.Lookup) (*pb.Node, error) {
	var id [sha256.Size]byte
	copy(id[:], lookup.Key[:sha256.Size])
	cNode := chord.NewNode(lookup.Node.IP, int(lookup.Node.Port))
	successor := s.Node.FindSuccessor(id, cNode)
	targetNode := &pb.Node{
		IP:   successor.IP,
		Port: int32(successor.Port),
		Key:  successor.Identifier[:],
	}
	return targetNode, nil
}

// GetSuccessor(context.Context, *empty.Empty) (*Node, error)
// ClosestPrecedingNode(context.Context, *Lookup) (*Node, error)
// FindSuccessor(context.Context, *Lookup) (*Node, error)
// GetPredecessor(context.Context, *empty.Empty) (*Node, error)
// Notify(context.Context, *Node) (*wrappers.BoolValue, error)
// GetFingerTable(context.Context, *empty.Empty) (*Nodes, error)
// GetSuccessorList(context.Context, *empty.Empty) (*Nodes, error)

// NewChordServer ip port
func NewChordServer(ip string, port int, node *chord.Node) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	chordServer := &ChordServer{
		Node: node,
	}
	pb.RegisterChordServer(grpcServer, chordServer)
	listener, _ := net.Listen("tcp", ip+":"+strconv.FormatInt(int64(port), 10))
	fmt.Printf("Start listening on makeNodeServer: %s\n", ip+":"+strconv.FormatInt(int64(port), 10))
	grpcServer.Serve(listener)
}
