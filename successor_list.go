package chord

// SuccessorList successor list
type SuccessorList struct {
	Nodes map[int]*Node
}

// NewSuccessorList make new successor list
func NewSuccessorList() *SuccessorList {
	return &SuccessorList{
		Nodes: make(map[int]*Node),
	}
}
