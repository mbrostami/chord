package chord

import (
	"github.com/mbrostami/chord/helpers"
)

type Stabilizer struct {
	successorList *SuccessorList
}

func NewStabilizer(successorList *SuccessorList) *Stabilizer {
	return &Stabilizer{
		successorList: successorList,
	}
}

// Start keep successor, successor list and predecessor updated
// Runs periodically
// ref E.1 - E.3
func (s *Stabilizer) Start(successor *RemoteNode, localNode *Node) (*RemoteNode, *SuccessorList) {
	successor, remotePredecessor, successorList := s.getSuccessorStablizerData(successor, localNode)

	// if all successors failed, then skip stabilizer to run next time
	if remotePredecessor == nil || localNode == nil {
		return nil, nil
	}
	// means successor's predececcor is changed
	if remotePredecessor.Identifier != localNode.Identifier {
		// if pred(succ) ∈ (n, succ)
		if helpers.BetweenR(remotePredecessor.Identifier, localNode.Identifier, successor.Identifier) {
			successor = remotePredecessor
		}
	}
	return successor, successorList
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
