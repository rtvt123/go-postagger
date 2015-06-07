package engine

type Node struct {
	parent *Node
	prob   float64
	tag    string
	word   string
}

func NewFullNode(word string, tag string, parent *Node, prob float64) *Node {
	instance := new(Node)
	instance.parent = parent
	instance.prob = prob
	instance.tag = tag
	instance.word = word

	return instance
}

func NewSimpleNode(word string, tag string) *Node {
	return NewFullNode(word, tag, nil, 0.0)
}
