package server

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/golang/protobuf/ptypes/wrappers"
	chordgrpc "github.com/mbrostami/chord/internal/grpc/chord"
	"github.com/mbrostami/chord/pkg/chord"
)

// ChordServer grpc server
type ChordServer struct {
	chordgrpc.UnimplementedChordServer
	ChordRing *chord.Chord
}

// NewChordServer create new storage server
func NewChordServer(chord *chord.Chord) *ChordServer {
	return &ChordServer{
		ChordRing: chord,
	}
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
