package compass

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FillParser struct {
	Contents string
	context  TemplateContext
	col      int
	line     int
}

type TemplateContext struct {
	variables map[string]Any
}

func NewTemplateContext(server *Server) TemplateContext {
	ctx := *server.DefaultTemplateContext
	return ctx
}

func NewEmptyTemplateContext() TemplateContext {
	return TemplateContext{variables: make(map[string]Any)}
}

func (ctx *TemplateContext) SetVariable(key string, value interface{}) {
	ctx.variables[key] = Any{Value: value}
}

func (ctx *TemplateContext) GetVariable(key string) string {
	if val, exists := ctx.variables[key]; exists {
		return val.ToString()
	}
	return "Unknown variable: " + key
}

func (parser *FillParser) Convert(server *Server) string {
	converted := ""
	stack := make([]string, 0) // Track nested if conditions

	// none = 0, var = 1, action = 2, insert = 3
	inPart := 0
	lastChars := ""
	skipContent := false

	for _, currentRune := range parser.Contents {
		char := string(currentRune)

		if char == "\n" {
			parser.line++
			parser.col = 0
		} else {
			parser.col++
		}

		if inPart == 0 && lastChars == "" && char == "<" {
			lastChars += char
			continue
		}

		if inPart == 0 && char == "$" {
			if lastChars != "<" {
				if !skipContent {
					converted += char
				}
				continue
			}
			inPart = 1
			lastChars = ""
			continue
		} else if inPart == 0 && char == "%" {
			if lastChars != "<" {
				if !skipContent {
					converted += char
				}
				continue
			}
			inPart = 2
			lastChars = ""
			continue
		} else if inPart == 0 && char == "@" {
			if lastChars != "<" {
				if !skipContent {
					converted += char
				}
				continue
			}
			inPart = 3
			lastChars = ""
			continue
		} else if inPart == 0 && lastChars == "<" {
			if !skipContent {
				converted += lastChars + char
			}
			lastChars = ""
			continue
		}

		if inPart == 1 && strings.HasSuffix(lastChars, "/>") {
			lastChars = strings.TrimSuffix(lastChars, "/>")
			if !skipContent {
				converted += parser.context.GetVariable(lastChars)
			}
			if !skipContent {
				converted += char
			}
			inPart = 0
			lastChars = ""
			continue
		}

		if inPart == 2 && strings.HasSuffix(lastChars, "/>") {
			lastChars = strings.TrimSuffix(lastChars, "/>")

			args := strings.SplitN(lastChars, " ", 2)
			cmd := args[0]
			arg := ""
			if len(args) > 1 {
				arg = args[1]
			}

			switch cmd {
			case "if":
				stack = append(stack, "if")
				skipContent = !parser.context.EvaluateVariable(arg)
			case "end":
				if len(stack) == 0 {
					return fmt.Sprintf("Unexpected end at line %d, col %d", parser.line, parser.col)
				}
				if stack[len(stack)-1] != "if" {
					return fmt.Sprintf("Mismatched end at line %d, col %d", parser.line, parser.col)
				}
				stack = stack[:len(stack)-1]
				skipContent = len(stack) > 0 && !parser.context.EvaluateVariable(stack[len(stack)-1])
			case "else":
				if len(stack) == 0 || stack[len(stack)-1] != "if" {
					return fmt.Sprintf("Unexpected else at line %d, col %d", parser.line, parser.col)
				}
				skipContent = !skipContent
			default:
				return fmt.Sprintf("Unknown command '%s' at line %d, col %d", cmd, parser.line, parser.col)
			}

			inPart = 0
			lastChars = ""
			continue
		}

		if inPart == 3 && strings.HasSuffix(lastChars, "/>") {
			lastChars = strings.TrimSuffix(lastChars, "/>")

			args := make(map[string]interface{})

			content := strings.Split(lastChars, " ")
			name := content[0]
			err := json.Unmarshal([]byte("{"+strings.Join(content[1:], " ")+"}"), &args)
			if err != nil {
				return fmt.Sprintf("Failed to load component variables (%s): %s", name, err)
			}

			if !skipContent {
				content, err := server.StylizeComponent(name, args, &parser.context)
				if err != nil {
					return fmt.Sprintf("Failed to load component: %s", err)
				}

				converted += content
			}
			if !skipContent {
				converted += char
			}
			inPart = 0
			lastChars = ""
			continue
		}

		if inPart == 1 || inPart == 2 || inPart == 3 {
			lastChars += char
			continue
		}

		if !skipContent {
			converted += char
		}
	}

	// Check for unclosed control structures
	if len(stack) > 0 {
		return fmt.Sprintf("Unclosed control structure '%s'", stack[len(stack)-1])
	}

	return converted
}

func (ctx *TemplateContext) EvaluateVariable(key string) bool {
	val, exists := ctx.variables[key]
	if !exists {
		return false
	}

	switch v := val.Value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case float64:
		return v != 0.0
	default:
		return false
	}
}

func Fill(template string, ctx TemplateContext, server *Server) Response {
	byteBody, err := os.ReadFile(filepath.Join(server.TemplatesDirectory, template))
	if err != nil {
		panic(err)
	}

	body := string(byteBody)
	return FillRaw(body, ctx, server)
}

func FillRaw(content string, ctx TemplateContext, server *Server) Response {
	parser := FillParser{Contents: content, context: ctx, col: -1, line: 0}
	return Text(parser.Convert(server))
}
