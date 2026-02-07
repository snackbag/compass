package passplate

import "fmt"

type ExpressionKind int

const (
	VariableKind ExpressionKind = iota
	TextKind
	BooleanKind
)

type Expression interface {
	Eval() string
	Repr() string

	Kind() ExpressionKind
}

type VariableExpr struct {
	Name string
}

func (e *VariableExpr) Kind() ExpressionKind {
	return VariableKind
}

func (e *VariableExpr) Eval() string {
	return ""
}

func (e *VariableExpr) Repr() string {
	return fmt.Sprintf("{$%s}", e.Name)
}

type TextExpression struct {
	Value string
}

func (e *TextExpression) Kind() ExpressionKind {
	return TextKind
}

func (e *TextExpression) Eval() string {
	return e.Value
}

func (e *TextExpression) Repr() string {
	return fmt.Sprintf("{%q}", e.Value)
}

type BooleanExpr struct {
	Left  Expression
	Right Expression
}

func (e *BooleanExpr) Kind() ExpressionKind {
	return BooleanKind
}

func (e *BooleanExpr) Eval() string {
	return ""
}

func (e *BooleanExpr) Repr() string {
	return fmt.Sprintf("{%s == %s}", e.Left.Repr(), e.Right.Repr())
}
