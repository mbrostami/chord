# Chord
Chord protocol implemented based on https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
Dstore implemented based on https://pdos.csail.mit.edu/papers/sit-phd-thesis.pdf   


# TODO
[ ] use https://github.com/grpc/grpc/blob/master/doc/health-checking.md instead of ping

# JOIN
- Initialize node
    - CalculateHASH
    - SET successor to itself 
    - SET predecessor to null
- find bootstrap node (for now it's static ip:port)
- FindSuccessor of self hash through bootstrap node
    - CONNECT to bootstrap node and call *FindSuccessor
    - receives successor ip:port
- CalculateHASH from received ip:port => identity
- SET successor to identity
    - SET fingerTable first record to successor
- event NOTIFY successor about joining
    - CONNECT to successor and call *NOTIFY

# Notify
**Caller is a client which is connected to the current node and called the NOTIFY
- If predecessor is null
    - SET Caller as predecessor
    - IF node is bootstrap node (if successor is equal to node itself)
        - SET Caller as successor
    - Exit
- Else
    - IF (Calculate if Caller identifier is between (predecessor identifier and current node identifier)) is true
        - SET caller as predecessor
- Exit


# FindSuccessor (Identifier)
- If successor is the same as node 
    - return node
    - Exit
- If Identifier ∈ (node identifier, successor identifier]
    - return Successor
    - Exit




// FindSuccessor find the closest node to the given identifier
// ref D
func (c *Chord) FindSuccessor(identifier [sha256.Size]byte) *Node {
	// fmt.Printf("FindSuccessor: start looking for key %x \n", identifier)
	if c.Successor.Identifier == c.Node.Identifier {
		return c.Node
	}
	// id ∈ (n, successor]
	if helpers.BetweenR(identifier, c.Node.Identifier, c.Successor.Identifier) {
		return c.Successor
	}
	nextNode := c.closestPrecedingNode(identifier)
	if nextNode.Identifier == c.Node.Identifier { // current node is the only node in figer table
		return nextNode
	}
	nextNodeSuccessor, err := c.clientAdapter.FindSuccessor(nextNode, identifier[:])
	if err != nil { // unexpected error on successor
		fmt.Printf("Unexpected error from successor %v", err)
		return nil
	}
	return nextNodeSuccessor
}


chord 
- node
    - ip
    - port
- hash
    - identity
- fingerTable
  - closestPrecedingNode 
    - *successorList
    - FixFingers (func) | calculate each record of the finger tabl
      - *findSuccessor
      - *notify
- successorList
  - updateSuccessorList (func)
- successor
  - *node
- predecessor
  - *node
- join (func)
  - connects to the bootstrap node and calls findSuccessor 
  - receives an Identity of a node

- findSuccessor (func)
- notify (func)
- replaceSuccessor (func)
- Stabilize (func)
- checkPredecessor (func) 
  - *ping
- ping (func)