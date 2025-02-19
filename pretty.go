package main

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

type VisitorFunc func(p *PrettyPrinter, node *sitter.Node)

type PrettyPrinter struct {
	Indent      string
	builder     *strings.Builder
	indentLevel int
	visitors    map[string]VisitorFunc
	input       []byte
}

func NewPrettyPrinter(indent string) *PrettyPrinter {
	p := &PrettyPrinter{
		Indent:   indent,
		builder:  &strings.Builder{},
		visitors: make(map[string]VisitorFunc),
	}
	for k, v := range defaultVisitors {
		p.visitors[k] = v
	}
	return p
}

func (p *PrettyPrinter) Format(input string) (string, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())
	p.input = []byte(input)

	tree, err := parser.ParseCtx(context.Background(), nil, p.input)
	if err != nil {
		return "", err
	}

	p.builder.Reset()
	p.indentLevel = 0
	p.visitNode(tree.RootNode())
	return p.builder.String(), nil
}

func (p *PrettyPrinter) visitNode(node *sitter.Node) {
	// fmt.Println("Visiting:", node.Type())
	if handler, exists := p.visitors[node.Type()]; exists {
		handler(p, node)
	} else {
		defaultVisit(p, node)
	}
}

func (p *PrettyPrinter) write(s string) {
	p.builder.WriteString(s)
}

func (p *PrettyPrinter) writeLine(s string) {
	p.builder.WriteString("\n" + strings.Repeat(p.Indent, p.indentLevel) + s)
}

func (p *PrettyPrinter) writeContent(node *sitter.Node) {
	p.write(p.content(node))
}

func (p *PrettyPrinter) indent() {
	p.indentLevel++
}

func (p *PrettyPrinter) unindent() {
	if p.indentLevel > 0 {
		p.indentLevel--
	}
}

func (p *PrettyPrinter) content(node *sitter.Node) string {
	return node.Content(p.input)
}

// Helper functions for common visitor patterns
func keywordVisitor(keyword string) VisitorFunc {
	return func(p *PrettyPrinter, node *sitter.Node) {
		p.write(keyword)
		defaultVisit(p, node)
	}
}

func modifierVisitor(modifier string) VisitorFunc {
	return func(p *PrettyPrinter, _ *sitter.Node) {
		p.write(modifier + " ")
	}
}

func symbolVisitor(symbol string) VisitorFunc {
	return func(p *PrettyPrinter, node *sitter.Node) {
		p.write(" " + symbol + " ")
	}
}

func contentVisitor() VisitorFunc {
	return func(p *PrettyPrinter, node *sitter.Node) {
		p.write(p.content(node))
	}
}

func defaultVisit(p *PrettyPrinter, node *sitter.Node) {
	for i := 0; i < int(node.ChildCount()); i++ {
		p.visitNode(node.Child(i))
	}
}

// Visitor definitions
var defaultVisitors = map[string]VisitorFunc{
	"program": defaultVisit,
	"php_tag": func(p *PrettyPrinter, node *sitter.Node) {
		p.write(p.content(node) + "\n")
	},
	"echo_statement": func(p *PrettyPrinter, n *sitter.Node) {
		p.writeLine(p.content(n.Child(0)) + " ")
		for i := 1; i < int(n.ChildCount()); i++ {
			if n.Child(i).Type() == ";" {
				p.visitNode(n.Child(i))
			} else {
				p.write(p.content(n.Child(i)))
			}
		}
	},

	// Declarations
	"trait_declaration":     keywordVisitor("trait "),
	"interface_declaration": keywordVisitor("interface "),
	"enum_declaration":      keywordVisitor("enum "),
	"class_declaration":     keywordVisitor("class "),
	"const_declaration":     keywordVisitor("const "),
	"method_declaration":    keywordVisitor("function "),

	// Modifiers
	"final_modifier":      modifierVisitor("final"),
	"abstract_modifier":   modifierVisitor("abstract"),
	"readonly_modifier":   modifierVisitor("readonly"),
	"static_modifier":     modifierVisitor("static"),
	"visibility_modifier": func(p *PrettyPrinter, n *sitter.Node) { p.write(p.content(n) + " ") },

	// Control structures
	"compound_statement": func(p *PrettyPrinter, n *sitter.Node) {
		p.write(" {")
		p.indent()
		defaultVisit(p, n)
		p.unindent()
		p.write("}")
	},
	"if_statement":    statementVisitor("if"),
	"while_statement": statementVisitor("while"),
	"for_statement":   loopVisitor("for"),
	"foreach_statement": func(p *PrettyPrinter, n *sitter.Node) {
		p.write("foreach ")
		processClauses(p, n, []string{"(", "as", ")"})
	},
	"else_if_clause": func(p *PrettyPrinter, node *sitter.Node) {
		p.write(" " + p.content(node) + " ")

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			p.visitNode(child)
		}
	},

	"else_clause": func(p *PrettyPrinter, node *sitter.Node) {
		p.write(" else")
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "compound_statement" {
				p.visitNode(child)
			}
		}
	},
	"update_expression": func(p *PrettyPrinter, n *sitter.Node) {
		fmt.Println("Update Expression", p.content(n))
		p.writeLine(p.content(n))
	},
	// Expressions
	"parenthesized_expression": func(p *PrettyPrinter, n *sitter.Node) {
		p.write("(")
		defaultVisit(p, n)
		p.write(")")
	},
	"expression_statement": func(p *PrettyPrinter, n *sitter.Node) {
		defaultVisit(p, n)
	},
	"assignment_expression": binaryOperatorVisitor(""),

	// Literals
	"integer":       contentVisitor(),
	"float":         contentVisitor(),
	"boolean":       contentVisitor(),
	"string":        contentVisitor(),
	"variable_name": contentVisitor(),

	// Special cases
	"use_declaration": func(p *PrettyPrinter, n *sitter.Node) {
		p.write("use ")
		defaultVisit(p, n)
		p.write(";\n")
	},
	"return_statement": func(p *PrettyPrinter, n *sitter.Node) {
		firstChild := n.Child(0)
		p.writeLine(p.content(firstChild) + " ")
		for i := 1; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			p.visitNode(child)
		}
	},

	"array_creation_expression": func(p *PrettyPrinter, node *sitter.Node) {
		p.write(p.content(node.Child(0)))
		for i := 1; i < int(node.ChildCount()); i++ {
			// fmt.Println("Child = ", node.Child(i).Type())
			if node.Child(i).Type() == "," {
				p.write(", ")
			} else {
				p.write(p.content(node.Child(i)))
			}
		}
	},
	"function_definition": visitFunctionDefinition,
	"formal_parameters": func(p *PrettyPrinter, n *sitter.Node) {
		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			if p.content(child) == "," {
				p.write(", ")
			} else {
				p.write(p.content(child))
			}
		}
	},

	";": func(p *PrettyPrinter, node *sitter.Node) {
		p.write(p.content(node) + "\n")
	},
}

func visitFunctionDefinition(p *PrettyPrinter, node *sitter.Node) {
	p.write(p.content(node.Child(0)) + " ")

	for i := 1; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "name" {
			p.write(p.content(child))
		} else {
			p.visitNode(child)
		}
	}
}

// Additional helper constructors
func statementVisitor(keyword string) VisitorFunc {
	return func(p *PrettyPrinter, n *sitter.Node) {
		p.writeLine(keyword + " ")
		defaultVisit(p, n)
	}
}

func loopVisitor(loopType string) VisitorFunc {
	return func(p *PrettyPrinter, n *sitter.Node) {
		p.write(loopType + " ")
		processClauses(p, n, []string{"(", ";", ")"})
	}
}

func processClauses(p *PrettyPrinter, node *sitter.Node, separators []string) {
	separatorIndex := 0
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == separators[separatorIndex] {

			p.write(separators[separatorIndex] + " ")
			separatorIndex = (separatorIndex + 1) % len(separators)
		} else {
			p.visitNode(child)
		}
	}
}

func binaryOperatorVisitor(operator string) VisitorFunc {
	return func(p *PrettyPrinter, n *sitter.Node) {
		for i := 0; i < int(n.ChildCount()); i++ {
			if i > 0 {
				p.write(operator)
			}
			p.visitNode(n.Child(i))
		}
	}
}

// Initialize symbol visitors programmatically
var symbolVisitors = []string{
	"+", "-", "*", "/", "%", "**", "+=", "-=", "*=", "/=", "%=", "**=",
	"=", "&", "|", "^", "<<", ">>", "&=", "|=", "^=", "<<=", ">>=",
	"==", "===", "!=", "<>", "!==", "<", "<=", ">", ">=", "??", "&&", "||",
}

func init() {
	for _, sym := range symbolVisitors {
		defaultVisitors[sym] = symbolVisitor(sym)
	}
}
