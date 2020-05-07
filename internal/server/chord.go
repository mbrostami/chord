package server

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"

	"github.com/golang/protobuf/ptypes/wrappers"
	chordgrpc "github.com/mbrostami/chord/internal/grpc/chord"
	"github.com/mbrostami/chord/pkg/chord"
	"google.golang.org/grpc"
)

// ChordServer grpc server
type ChordServer struct {
	chordgrpc.UnimplementedChordServer
	ChordRing *chord.Chord
}

// Notify update predecessor
// is being called periodically
func (s *ChordServer) Notify(ctx context.Context, n *chordgrpc.Node) (*wrappers.BoolValue, error) {
	node := ConvertToChordNode(n) // make node struct and calculate identifier
	result := &wrappers.BoolValue{}
	result.Value = s.ChordRing.Notify(node)
	return result, nil
}

// GetStablizerData get predecessor node + successor list
func (s *ChordServer) GetStablizerData(ctx context.Context, caller *chordgrpc.Node) (*chordgrpc.StablizerData, error) {
	stablizeData := &chordgrpc.StablizerData{}
	cnode := ConvertToChordNode(caller)
	predecessor := s.ChordRing.GetPredecessor(cnode)
	stablizeData.Predecessor = ConvertToGrpcNode(predecessor)
	stablizeData.SuccessorList = ConvertToGrpcSuccessorList(s.ChordRing.GetSuccessorList())
	return stablizeData, nil
}

// FindSuccessor get closest node to the given key
func (s *ChordServer) FindSuccessor(ctx context.Context, lookup *chordgrpc.Lookup) (*chordgrpc.Node, error) {
	var id [sha256.Size]byte
	copy(id[:], lookup.Key[:sha256.Size])
	successor := s.ChordRing.FindSuccessor(id)
	if successor == nil {
		return nil, errors.New("successor is null")
	}
	return ConvertToGrpcNode(successor), nil
}

// NewChordServer ip port
func NewChordServer(chordRing *chord.Chord) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	chordServer := &ChordServer{
		ChordRing: chordRing,
	}
	chordgrpc.RegisterChordServer(grpcServer, chordServer)
	listener, _ := net.Listen("tcp", chordRing.Node.FullAddr())
	fmt.Printf("Start listening on makeNodeServer: %s\n", chordRing.Node.FullAddr())
	grpcServer.Serve(listener)
}
