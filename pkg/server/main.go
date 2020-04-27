package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/pkg/chord"
	"google.golang.org/grpc"
)

type ChordServer struct {
	pb.UnimplementedChordServer
	Node *chord.Node
}

// GetPredecessor get predecessor node
func (s *ChordServer) GetPredecessor(ctx context.Context, empty *empty.Empty) (*pb.Node, error) {
	var key []byte
	var ip string
	var port int32
	if s.Node.Predecessor != nil {
		key = s.Node.Predecessor.Identifier[:]
		ip = s.Node.Predecessor.IP
		port = int32(s.Node.Predecessor.Port)
	}
	node := &pb.Node{
		IP:   ip,
		Port: port,
		Key:  key,
	}
	fmt.Printf("grpc Server: GetPredecessor: %+v \n", s.Node.Predecessor)
	return node, nil
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
