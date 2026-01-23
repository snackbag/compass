package passplate

import "strings"

type NodeKind int

const (
	NodeRoot NodeKind = iota
	NodeText
	NodeExpr
	NodeIf
	NodeFor
)

type Node interface {
	Kind() NodeKind
}

//
// Root Node
//

type RootNode struct {
	Children []Node
	Parent   *RootNode
}

func (n *RootNode) Kind() NodeKind {
	return NodeRoot
}

func NewRootNode() *RootNode {
	return &RootNode{make([]Node, 0), nil}
}

//
// Text Node
//

type TextNode struct {
	Content string
}

func (n *TextNode) Kind() NodeKind {
	return NodeText
}

func NewTextNode(content string) *TextNode {
	return &TextNode{Content: content}
}

//
// Expression Node
//

type ExprNode struct {
	Expressions []Expression
}

func (n *ExprNode) Kind() NodeKind {
	return NodeExpr
}

func (n *ExprNode) Eval() string {
	builder := strings.Builder{}

	for _, expr := range n.Expressions {
		builder.WriteString(expr.Eval())
	}

	return builder.String()
}

func (n *ExprNode) Repr() string {
	builder := strings.Builder{}

	for _, expr := range n.Expressions {
		builder.WriteString(expr.Repr())
	}

	return builder.String()
}

func NewExprNode() *ExprNode {
	return &ExprNode{make([]Expression, 0)}
}

//
// If Node
//

type IfNode struct {
	IfExpr   *BooleanExpr
	IfClause *RootNode

	ElseIfs map[*BooleanExpr]*RootNode

	ElseClause *RootNode
}

func (n *IfNode) Kind() NodeKind {
	return NodeIf
}
