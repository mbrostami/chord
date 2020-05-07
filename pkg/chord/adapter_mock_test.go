package chord

import "errors"

func MockClientInterface() *ClientInterfaceMock {
	return &ClientInterfaceMock{}
}

func MockFindSuccessor(c *ClientInterfaceMock, ip string, port int) {
	c.FindSuccessorFunc = func(remote *Node, identifier []byte) (*Node, error) {
		n := NewNode(ip, port)
		return n, nil
	}
}

func MockFindSuccessorReturnRemote(c *ClientInterfaceMock) {
	c.FindSuccessorFunc = func(remote *Node, identifier []byte) (*Node, error) {
		return remote, nil
	}
}

func MockFindSuccessorError(c *ClientInterfaceMock) {
	c.FindSuccessorFunc = func(remote *Node, identifier []byte) (*Node, error) {
		return nil, errors.New("error happened")
	}
}

func MockNotify(c *ClientInterfaceMock) {
	c.NotifyFunc = func(remote, node *Node) error {
		return nil
	}
}

func MockNotifyError(c *ClientInterfaceMock) {
	c.NotifyFunc = func(remote, node *Node) error {
		return errors.New("error happened")
	}
}

func MockGetStablizerData(c *ClientInterfaceMock) {
	c.GetStablizerDataFunc = func(remote *Node, node *Node) (*Node, *SuccessorList, error) {
		s := NewSuccessorList()
		return remote, s, nil
	}
}
