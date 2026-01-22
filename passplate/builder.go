package passplate

import "strings"

type BuildState int

const (
	StateIdle = iota
	StateExpr
)

func Read(raw string) *RootNode {
	buffer := ""
	state := StateIdle

	root := NewRootNode()
	cursor := root
	inString := false

	for _, r := range []rune(raw) {
		char := string(r)
		buffer += char

		switch state {
		case StateIdle:
			if strings.HasSuffix(buffer, "<$") {
				state = StateExpr

				cursor.Children = append(cursor.Children, NewTextNode(strings.TrimSuffix(buffer, "<$")))
				buffer = ""
				cursor.Children = append(cursor.Children)
			}

		case StateExpr:
			if strings.HasSuffix(buffer, "\"") {
				inString = !inString
			}

			if !inString && strings.HasSuffix(buffer, "/>") {
				state = StateIdle

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
