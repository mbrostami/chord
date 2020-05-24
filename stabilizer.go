package chord

import (
	"errors"

	"github.com/mbrostami/chord/helpers"
)

type Stabilizer struct {
	successorList   *SuccessorList
	predecessorList *PredecessorList
}

func NewStabilizer(successorList *SuccessorList, predecessorList *PredecessorList) *Stabilizer {
	return &Stabilizer{
		successorList:   successorList,
		predecessorList: predecessorList,
	}
}

// StartSuccessorList keep successor, successor list and predecessor updated
// Runs periodically
// ref E.1 - E.3
func (s *Stabilizer) StartSuccessorList(successor *RemoteNode, localNode *Node) (*RemoteNode, *SuccessorList, error) {
	successor, remotePredecessor, successorList := s.getSuccessorStablizerData(successor, localNode)

	// if all successors failed, then skip stabilizer to run next time
	if remotePredecessor.Node == nil || localNode == nil {
		return nil, nil, errors.New("predecessor is not valid")
	}
	// means successor's predececcor is changed
	if remotePredecessor.Identifier != localNode.Identifier {
		// if pred(succ) ∈ (n, succ)
		if helpers.BetweenR(remotePredecessor.Identifier, localNode.Identifier, successor.Identifier) {
			successor = remotePredecessor
		}
	}
	return successor, successorList, nil
}

// StartPredecessorList keep Predecessor, Predecessor list and predecessor updated
// Runs periodically
// ref E.1 - E.3
func (s *Stabilizer) StartPredecessorList(predecessor *RemoteNode, localNode *Node) (*RemoteNode, *PredecessorList, error) {
	predecessor, predecessorList := s.getPredecessorList(predecessor, localNode)
	return predecessor, predecessorList, nil
}

// getSuccessorStablizerData get stabilizer data from successor
// if successor is not available, replace it with the next available successor
func (s *Stabilizer) getSuccessorStablizerData(successor *RemoteNode, localNode *Node) (*RemoteNode, *RemoteNode, *SuccessorList) {
	remotePredecessor, successorList, err := successor.GetStablizerData(localNode)
	if err != nil {
		// replace next available successor from successorList
		for i := 1; i < len(s.successorList.Nodes); i++ {
			remotNode := s.successorList.Nodes[i]
			remotePredecessor, successorList, err = remotNode.GetStablizerData(localNode)
			if err == nil {
				successor = remotNode
				break
			}
		}
	}
	return successor, remotePredecessor, successorList
}

// getPredecessorList get predecessor list from predecessor
// if predecessor is not available, replace it with the next available predecessor
func (s *Stabilizer) getPredecessorList(predecessor *RemoteNode, localNode *Node) (*RemoteNode, *PredecessorList) {
	predecessorList, err := predecessor.GetPredecessorList(localNode)
	if err != nil {
		// replace next available predecessor from predecessorList
		for i := 1; i < len(s.predecessorList.Nodes); i++ {
			remotNode := s.predecessorList.Nodes[i]
			predecessorList, err = remotNode.GetPredecessorList(localNode)
			if err == nil {
				predecessor = remotNode
				break
			}
		}
	}
	return predecessor, predecessorList
}
