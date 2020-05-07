package chord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChordRing(t *testing.T) {
	ip := "127.0.0.1"
	port := 10001
	ring := NewRing(ip, port, MockClientInterface())
	assert.Equal(t, "127.0.0.1", ring.Node.IP)
	assert.Equal(t, 10001, ring.Node.Port)
	assert.Equal(t, "127.0.0.1", ring.Successor.IP)
	assert.Equal(t, 10001, ring.Successor.Port)
	assert.Nil(t, ring.Predecessor)
}

func TestSuccessfulJoin(t *testing.T) {
	ip := "127.0.0.1"
	port := 10001
	client := MockClientInterface()
	MockFindSuccessor(client, "127.0.0.2", 10002)
	MockNotify(client)

	ring := NewRing(ip, port, client)

	remote := &Node{}
	remote.IP = ip
	remote.Port = port
	err := ring.Join(remote)

	assert.Nil(t, err)
	assert.Nil(t, ring.Predecessor)
	assert.Equal(t, "127.0.0.2", ring.Successor.IP)
	assert.Equal(t, 10002, ring.Successor.Port)

	assert.Equal(t, 1, len(ring.FingerTable))
	assert.Equal(t, "127.0.0.2:10002", ring.FingerTable[1].FullAddr())
}

func TestFailedJoin(t *testing.T) {
	ip := "127.0.0.1"
	port := 10001
	client := MockClientInterface()
	MockFindSuccessorError(client)
	MockNotify(client)

	ring := NewRing(ip, port, client)

	remote := &Node{}
	remote.IP = ip
	remote.Port = port
	err := ring.Join(remote)

	assert.NotNil(t, err)
	assert.Nil(t, ring.Predecessor)
	assert.Equal(t, "127.0.0.1", ring.Successor.IP) // should not be changed in join
	assert.Equal(t, 10001, ring.Successor.Port)     // should not be changed in join

	assert.Equal(t, 0, len(ring.FingerTable))
}

func TestNotify(t *testing.T) {
	ip := "127.0.0.3"
	port := 10003
	client := MockClientInterface()
	MockFindSuccessor(client, "127.0.0.2", 10002)
	MockNotify(client)

	ring := NewRing(ip, port, client) // identity 8671c42d43832fc5aec71a43a1b3baf4409f965cf912bbf88eab694e884e37ab

	err := ring.Join(NewNode(ip, port))
	assert.Nil(t, err)

	firstNotify := NewNode("127.0.0.1", 10001) // create new node | 09e14941532f07258a269c10c7750add26b56e687fbb5243ccdfac133d9b5e71
	// predecessor is nil so it will be firstNotify
	notify := ring.Notify(firstNotify)
	assert.True(t, notify)
	assert.Equal(t, "127.0.0.1:10001", ring.Predecessor.FullAddr())

	// second time notify
	secondNotify := NewNode("127.0.0.4", 10004) // create new node | 44c02611d2391bb3403fec0afbeafd5d3d9da10338daef427aee63e6d08a1418
	// predecessor is not nil so it will check secondNotify Between (ring.Predecessor, ring.Node) which is true
	notify = ring.Notify(secondNotify)
	assert.True(t, notify)
	assert.Equal(t, "127.0.0.4:10004", ring.Predecessor.FullAddr())

	// second time notify
	thirdNotify := NewNode("127.0.0.5", 10005) // create new node | b10e478870b09b81e88d6c560a3b9b5e4fcf45851de6ec429249b9933fa842b6
	// predecessor is not nil so it will check secondNotify Between (ring.Predecessor, ring.Node) which is false
	notify = ring.Notify(thirdNotify)
	assert.False(t, notify)
	assert.Equal(t, "127.0.0.4:10004", ring.Predecessor.FullAddr())
}

func TestBootstrapNotify(t *testing.T) {
	ip := "127.0.0.1"
	port := 10001
	client := MockClientInterface()
	MockFindSuccessor(client, "127.0.0.1", 10001)
	MockNotify(client)

	ring := NewRing(ip, port, client)

	err := ring.Join(NewNode(ip, port))
	assert.Nil(t, err)

	firstNotify := NewNode("127.0.0.2", 10002)
	// predecessor is nil and c.Successor.Identifier == c.Node.Identifier | successor will be updated
	notify := ring.Notify(firstNotify)
	assert.True(t, notify)
	assert.Equal(t, "127.0.0.2:10002", ring.Successor.FullAddr())
}

func TestBootstrapFindSuccessor(t *testing.T) {
	ip := "127.0.0.1"
	port := 10001
	client := MockClientInterface()
	MockFindSuccessor(client, "127.0.0.1", 10001)
	MockNotify(client)

	ring := NewRing(ip, port, client) // 09e14941532f07258a269c10c7750add26b56e687fbb5243ccdfac133d9b5e71

	err := ring.Join(NewNode(ip, port))
	assert.Nil(t, err)

	n1 := NewNode("127.0.0.2", 10002)
	identifier := n1.Identifier
	result := ring.FindSuccessor(identifier)
	assert.Equal(t, result.Identifier, ring.Node.Identifier)
}

func TestFindSuccessor(t *testing.T) {
	ip := "127.0.0.1"
	port := 10001
	client := MockClientInterface()
	MockFindSuccessor(client, "127.0.0.2", 10002) // c0bcff0f
	MockNotify(client)

	ring := NewRing(ip, port, client) // node = 09e14941 successor = c0bcff0f

	err := ring.Join(NewNode(ip, port))
	assert.Nil(t, err)

	n1 := NewNode("127.0.0.2", 10002) // c0bcff0f
	// id ∈ (n, successor]
	result := ring.FindSuccessor(n1.Identifier) // should return successor id
	assert.Equal(t, result.Identifier, ring.Successor.Identifier)
}

func TestClosestSuccessor(t *testing.T) {
	client := MockClientInterface()
	MockFindSuccessorReturnRemote(client)
	MockNotify(client)

	ring := NewRing("127.0.0.4", 10004, client) // node = 44c02611

	err := ring.Join(NewNode("127.0.0.1", 10010)) // successor = 6a1f5339
	assert.Nil(t, err)

	zzz := newNode("zzz") // 7a7a7a
	// id ∈ (n, successor]
	// 7a7a7a ∈ (44c02611, 6a1f5339] = false
	// finger[1] ∈ (n, id) - 6a1f5339 ∈ (44c02611, 7a7a7a) = true
	result := ring.FindSuccessor(zzz.Identifier)
	assert.Equal(t, ring.Successor.Identifier, result.Identifier)

	yyy := newNode("yyy") // 797979
	ring.FingerTable[2] = zzz
	// 797979 ∈ (44c02611, 6a1f5339] = false
	// finger[2] ∈ (n, id) - 7a7a7a ∈ (44c02611, 797979) = false
	// finger[1] ∈ (n, id) - 6a1f5339 ∈ (44c02611, 797979) = true
	result = ring.FindSuccessor(yyy.Identifier)
	assert.Equal(t, ring.Successor.Identifier, result.Identifier)

	ring.FingerTable[2] = yyy
	// 7a7a7a ∈ (44c02611, 6a1f5339] = false
	// finger[2] ∈ (n, id) - 797979 ∈ (09e14941, 7a7a7a) = true
	result = ring.FindSuccessor(zzz.Identifier)
	assert.Equal(t, yyy.Identifier, result.Identifier)

	closerToYYY := newNode("yyx") // 797978
	ring.FingerTable[2] = zzz
	ring.SuccessorList.Nodes[0] = ring.Successor
	ring.SuccessorList.Nodes[1] = closerToYYY
	// id ∈ (node, successor) -- 797979 ∈ (44c02611, 6a1f5339] = false
	// finger[2] ∈ (n, id) -- 7a7a7a ∈ (44c02611, 797979) = false
	// finger[1] ∈ (n, id) -- 6a1f5339 ∈ (44c02611, 797979) = true
	// successorList[0] ∈ (finger[1], id) -- 6a1f5339 ∈ (6a1f5339, 797979) = false
	// successorList[1] ∈ (finger[1], id) -- 797978 ∈ (6a1f5339, 797979) = true (its closer to 797979)
	result = ring.FindSuccessor(yyy.Identifier)
	assert.Equal(t, closerToYYY.Identifier, result.Identifier)
}

func TestGetPredecessor(t *testing.T) {
	client := MockClientInterface()
	MockFindSuccessorReturnRemote(client)
	MockNotify(client)

	ring := NewRing("127.0.0.4", 10004, client) // node = 44c02611

	// predecessor is null -> return self node
	assert.Equal(t, ring.GetPredecessor(newNode("000")).FullAddr(), "127.0.0.4:10004")

	n000 := newNode("000") // 303030
	n111 := newNode("111") // 313131
	ring.Predecessor = n000
	// id ∈ (predecessor, node) -- 313131 ∈ (303030, 44c02611) = true
	ring.GetPredecessor(n111)
	assert.Equal(t, ring.Predecessor.Identifier, n111.Identifier)

	ring.Predecessor = n111 // 303030
	// id ∈ (predecessor, node) -- 303030 ∈ (313131, 44c02611) = true
	ring.GetPredecessor(n000)
	// predecessor should not be changed
	assert.Equal(t, ring.Predecessor.Identifier, n111.Identifier)
}

func newNode(id string) *Node {
	var identifier [IdentifierSize]byte
	b := []byte(id)
	copy(identifier[:], b)
	n := &Node{
		IP:         "fake",
		Port:       0,
		Identifier: identifier,
	}
	return n
}
