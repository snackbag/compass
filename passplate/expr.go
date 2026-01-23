package passplate

import "fmt"

type Expression interface {
	Eval() string
	Repr() string
}

type VariableExpr struct {
	Name string
}

func (e *VariableExpr) Eval() string {
	return ""
}

func (e *VariableExpr) Repr() string {
	return fmt.Sprintf("{$%s}", e.Name)
}

type BooleanExpr struct {
	Left  Expression
	Right Expression
}

func (e *BooleanExpr) Eval() string {
	return ""
}

func (e *BooleanExpr) Repr() string {
	return fmt.Sprintf("{%s == %s}", e.Left.Repr(), e.Right.Repr())
}
