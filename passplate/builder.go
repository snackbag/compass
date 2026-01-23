package passplate

import (
	"fmt"
	"strings"
)

type BuildState int

const (
	StateIdle = iota
	StateExpr
	StateIf
	StateElif
	StateElse
)

type stateMachine struct {
	states []BuildState
}

func (m *stateMachine) Pop() {
	m.states = m.states[:len(m.states)-1]

	if len(m.states) == 0 {
		m.states = append(m.states, StateIdle)
	}
}

func (m *stateMachine) Push(state BuildState) {
	m.states = append(m.states, state)
}

func (m *stateMachine) State() BuildState {
	return m.states[len(m.states)-1]
}

func Read(raw string) (*RootNode, error) {
	buffer := ""
	s := &stateMachine{make([]BuildState, 0)}
	s.states = append(s.states, StateIdle)

	root := NewRootNode()
	cursor := root
	inString := false

	for _, r := range []rune(raw) {
		char := string(r)
		buffer += char

		switch s.State() {
		case StateIdle:
			if strings.HasSuffix(buffer, "<$") {
				s.Push(StateExpr)

				cursor.Children = appendBuffer(cursor.Children, buffer, "<$")
				buffer = ""
				cursor.Children = append(cursor.Children)
			} else if strings.HasSuffix(buffer, "<%if ") {
				s.Push(StateIf)

				cursor.Children = appendBuffer(cursor.Children, buffer, "<%if ")
				buffer = ""
			} else if strings.HasSuffix(buffer, "<%elif ") {
				s.Push(StateElif)

				cursor.Children = appendBuffer(cursor.Children, buffer, "<%elif ")
				buffer = ""
			} else if strings.HasSuffix(buffer, "<%else/>") {
				s.Push(StateElse)

				cursor.Children = appendBuffer(cursor.Children, buffer, "<%else/>")
				buffer = ""
			} else if strings.HasSuffix(buffer, "<%end/>") {
				cursor.Children = appendBuffer(cursor.Children, buffer, "<%end/>")
				buffer = ""
				cursor = cursor.Parent
			}

		case StateExpr:
			if strings.HasSuffix(buffer, "\"") {
				inString = !inString
			}

			if !inString && strings.HasSuffix(buffer, "/>") {
				s.Pop()

				expr, err := createExpression(strings.TrimSuffix(buffer, "/>"))
				if err != nil {
					return root, fmt.Errorf("failed to create expression: %s", err)
				}
				node := NewExprNode()
				node.Expression = expr
				cursor.Children = append(cursor.Children, node)

				buffer = ""
			}

		case StateIf:
			if strings.HasSuffix(buffer, "\"") {
				inString = !inString
			}

			if !inString && strings.HasSuffix(buffer, "/>") {
				s.Pop()

				clause := NewRootNode()
				clause.Parent = cursor

				expr, err := createBooleanExpr(strings.TrimSuffix(buffer, "/>"))
				if err != nil {
					return root, fmt.Errorf("failed to create if: %s", err)
				}
				node := &IfNode{IfClause: clause, IfExpr: expr, ElseIfs: make(map[*BooleanExpr]*RootNode)}
				cursor.Children = append(cursor.Children, node)
				cursor = clause

				buffer = ""
			}

		case StateElif:
			if strings.HasSuffix(buffer, "\"") {
				inString = !inString
			}

			if !inString && strings.HasSuffix(buffer, "/>") {
				s.Pop()

				clause := NewRootNode()
				clause.Parent = cursor.Parent
				node := cursor.Parent.LastChild().(*IfNode)

				expr, err := createBooleanExpr(strings.TrimSuffix(buffer, "/>"))
				if err != nil {
					return root, fmt.Errorf("failed to create elif: %s", err)
				}
				node.ElseIfs[expr] = clause
				cursor = clause

				buffer = ""
			}

		case StateElse:
			s.Pop()

			rn := NewRootNode()
			rn.Parent = cursor.Parent
			node := cursor.Parent.LastChild().(*IfNode)
			node.ElseClause = rn
			cursor = rn
		}
	}

	cursor.Children = append(cursor.Children, NewTextNode(buffer))
	return root, nil
}

func createExpression(buffer string) (Expression, error) {
	return &VariableExpr{Name: buffer}, nil
}

func createBooleanExpr(buffer string) (*BooleanExpr, error) {
	return &BooleanExpr{Right: &TextExpression{Value: "test"}, Left: &VariableExpr{Name: "admin"}}, nil
}

func appendBuffer(children []Node, buffer string, suffix string) []Node {
	return append(children, NewTextNode(strings.TrimSuffix(buffer, suffix)))
}
