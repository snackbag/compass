package passplate

import "fmt"

type Node interface {
	Type() string
	String() string
	Equals(target Node) bool
}

type TextNode struct {
	Node

	Value string
}

func (n *TextNode) Type() string {
	return "Text"
}

func (n *TextNode) String() string {
	return fmt.Sprintf(`{"type": %q, "value": %q}`, n.Type(), n.Value)
}

func (n *TextNode) Equals(target Node) bool {
	return target.Type() == n.Type() && target.String() == n.String()
}
