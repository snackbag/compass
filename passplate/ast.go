package passplate

import (
	"fmt"
	"strings"
)

func Represent(_n Node, indent int) string {
	switch n := _n.(type) {
	case *RootNode:
		{
			builder := strings.Builder{}

			for _, child := range n.Children {
				builder.WriteString(Represent(child, indent+1))
				builder.WriteString("\n")
			}

			return builder.String()
		}

	case *TextNode:
		return fmt.Sprintf("<Text: %s/>", n.Content)

	case *ExprNode:
		return fmt.Sprintf("<Expression: %s/>", n.Repr())

	case *IfNode:
		return n.Repr(indent)
	}

	return ""
}
