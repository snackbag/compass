package passplate

import "strings"

type BuildState int

const (
	StateIdle = iota
	StateExpr
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

func Read(raw string) *RootNode {
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

				cursor.Children = append(cursor.Children, NewTextNode(strings.TrimSuffix(buffer, "<$")))
				buffer = ""
				cursor.Children = append(cursor.Children)
			}

		case StateExpr:
			if strings.HasSuffix(buffer, "\"") {
				inString = !inString
			}

			if !inString && strings.HasSuffix(buffer, "/>") {
				s.Pop()

				node := NewExprNode()
				node.Expressions = createExpressions(strings.TrimSuffix(buffer, "/>"))
				cursor.Children = append(cursor.Children, node)

				buffer = ""
			}
		}
	}

	cursor.Children = append(cursor.Children, NewTextNode(buffer))
	return root
}

func createExpressions(buffer string) []Expression {
	rv := make([]Expression, 0)
	rv = append(rv, &VariableExpr{Name: buffer})

	return rv
}
