package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	php "github.com/smacker/go-tree-sitter/php"
)

type CFG struct {
	Nodes map[int]*CFGNode
	Edges map[int][]int
}

type CFGNode struct {
	ID   int
	Type string
	code string // info pour debug
}

func NewCFG() *CFG {
	return &CFG{
		Nodes: make(map[int]*CFGNode),
		Edges: make(map[int][]int),
	}
}

func (cfg *CFG) AddNode(nodeType, codeSnippet string, id int) {
	cfg.Nodes[id] = &CFGNode{
		ID:   id,
		Type: nodeType,
		code: codeSnippet,
	}
}

func (cfg *CFG) AddEdge(src, dst int) {
	cfg.Edges[src] = append(cfg.Edges[src], dst)
}

type CFGBuilder struct {
	parser *sitter.Parser
	cfg    *CFG
	nextID int
	source []byte
}

func NewCFGBuilder() *CFGBuilder {
	p := sitter.NewParser()
	p.SetLanguage(php.GetLanguage())
	return &CFGBuilder{
		parser: p,
		cfg:    NewCFG(),
		nextID: 1,
	}
}

func (b *CFGBuilder) newID() int {
	id := b.nextID
	b.nextID++
	return id
}

func (b *CFGBuilder) BuildCFG(source []byte) (*CFG, error) {
	b.source = source

	tree, err := b.parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}
	root := tree.RootNode()

	entryID := b.newID()
	b.cfg.AddNode(NodeEntry, NodeEntry, entryID)

	lastNodeID := b.visit(root, entryID)

	exitID := b.newID()
	b.cfg.AddNode(NodeExit, NodeExit, exitID)

	// Ensure last node connects to Exit
	if lastNodeID != entryID {
		b.cfg.AddEdge(lastNodeID, exitID)
	} else {
		b.cfg.AddEdge(entryID, exitID)
	}

	return b.cfg, nil
}

func (b *CFGBuilder) visit(node *sitter.Node, parentID int) int {
	if node == nil {
		return parentID
	}

	fmt.Println("Visiting node:", node.Type())

	switch node.Type() {

	case "php_tag":
		return b.addGenericNode(NodeHtml, node, parentID)

	case "assignment_expression":
		lhsNode := node.Child(int(node.ChildCount()) - 1)
		lNodeID := b.visit(lhsNode, parentID)

		rNode := node.Child(0)
		rNodeID := b.visit(rNode, lNodeID)

		for i := 1; i < int(node.ChildCount())-1; i++ {
			child := node.Child(i)
			parentID = b.visit(child, rNodeID)
		}

		return parentID

	case "=", "+=", "-=", "*=", "/=":
		return b.addGenericNode(NodeBinOp, node, parentID)

	case "==", "!=", "<", "<=", ">", ">=":
		return b.addGenericNode(NodeRelOp, node, parentID)

	case "+", "-", "*", "/":
		return b.addGenericNode(NodeBinOp, node, parentID)

	case "if_statement":
		ifID := b.newID()
		b.cfg.AddNode(NodeIf, NodeIf, ifID)
		b.cfg.AddEdge(parentID, ifID)

		conditionNode := node.ChildByFieldName("condition")

		if conditionNode.ChildCount() == 3 && conditionNode.Child(1).Type() == "binary_expression" {
			conditionNode = conditionNode.Child(1)
		}

		leftOperand := conditionNode.Child(0)
		leftID := b.visit(leftOperand, ifID)

		rightOperand := conditionNode.Child(2)
		rightID := b.visit(rightOperand, leftID)

		operatorNode := conditionNode.Child(1)
		operatorID := b.visit(operatorNode, rightID)

		conditionID := b.newID()
		b.cfg.AddNode(NodeCondition, NodeCondition, conditionID)
		b.cfg.AddEdge(operatorID, conditionID)

		trueBlock := node.ChildByFieldName("body")
		trueBranchID := b.visit(trueBlock, conditionID)

		elseBlock := node.ChildByFieldName("alternative")
		var falseBranchID int
		if elseBlock != nil {
			falseBranchID = b.visit(elseBlock, conditionID)
		} else {
			falseBranchID = conditionID
		}

		ifEndID := b.newID()
		b.cfg.AddNode(NodeIfEnd, NodeIfEnd, ifEndID)

		b.cfg.AddEdge(trueBranchID, ifEndID)
		b.cfg.AddEdge(falseBranchID, ifEndID)

		return ifEndID

	case "echo_statement":
		echoID := b.newID()
		b.cfg.AddNode(NodeEcho, NodeEcho, echoID)
		b.cfg.AddEdge(parentID, echoID)

		argNode := node.Child(1)
		if argNode != nil {
			argID := b.visit(argNode, echoID)
			return argID
		}

		return echoID

	case "function_call_expression":
		funcCallID := b.newID()
		b.cfg.AddNode(NodeFunctionCall, "FunctionCall", funcCallID)
		b.cfg.AddEdge(parentID, funcCallID)

		// Extract function name
		funcNameNode := node.Child(0)
		funcNameID := b.newID()
		funcName := funcNameNode.Content(b.source)
		b.cfg.AddNode(NodeId, funcNameNode.Content(b.source), funcNameID)
		b.cfg.AddEdge(funcCallID, funcNameID)

		argumentsNode := node.ChildByFieldName("arguments")
		if argumentsNode != nil {
			argsID := b.newID()
			b.cfg.AddNode(NodeArgumentList, NodeArgumentList, argsID)
			b.cfg.AddEdge(funcNameID, argsID)

			var lastArgID int = argsID
			for i := 0; i < int(argumentsNode.ChildCount()); i++ {
				argNode := argumentsNode.Child(i)
				if argNode.Type() != "(" && argNode.Type() != ")" {
					argumentID := b.newID()
					b.cfg.AddNode(NodeArgument, NodeArgument, argumentID)
					b.cfg.AddEdge(argsID, argumentID)

					argContentID := b.visit(argNode, argumentID)
					lastArgID = argContentID
				}
			}

			callBeginID := b.newID()
			b.cfg.AddNode(NodeCallBegin, funcName, callBeginID)
			b.cfg.AddEdge(lastArgID, callBeginID)

			callEndID := b.newID()
			b.cfg.AddNode(NodeCallEnd, funcName, callEndID)
			b.cfg.AddEdge(callBeginID, callEndID)

			retValueID := b.newID()
			b.cfg.AddNode(NodeRetValue, "RetValue", retValueID)
			b.cfg.AddEdge(callEndID, retValueID)

			return retValueID
		}

		return funcCallID

		// case "arguments":
		// 	// Create ArgumentList node
		// 	argsID := b.newID()
		// 	b.cfg.AddNode(NodeArgumentList, "ArgumentList", argsID)
		// 	b.cfg.AddEdge(parentID, argsID)

		// 	// Process each argument **only once**
		// 	for i := 0; i < int(node.ChildCount()); i++ {
		// 		argNode := node.Child(i)
		// 		if argNode.Type() != "(" && argNode.Type() != ")" { // Ignore brackets
		// 			argID := b.visit(argNode, argsID)

		// 			argumentID := b.newID()
		// 			b.cfg.AddNode(NodeArgument, "Argument", argumentID)
		// 			b.cfg.AddEdge(argsID, argumentID) // ✅ Only one edge per argument
		// 			b.cfg.AddEdge(argumentID, argID)  // ✅ Argument wraps its value
		// 		}
		// 	}
		// 	return argsID

	case "while_statement":
		whileID := b.newID()
		b.cfg.AddNode(NodeWhile, NodeWhile, whileID)
		b.cfg.AddEdge(parentID, whileID)

		conditionNode := node.ChildByFieldName("condition")
		conditionID := b.visit(conditionNode, whileID)

		bodyNode := node.ChildByFieldName("body")
		bodyID := b.visit(bodyNode, conditionID)

		b.cfg.AddEdge(bodyID, whileID)

		exitWhileID := b.newID()
		b.cfg.AddNode(NodeWhileEnd, NodeWhileEnd, exitWhileID)
		b.cfg.AddEdge(whileID, exitWhileID)

		return exitWhileID

	case "break_statement":
		breakID := b.newID()
		b.cfg.AddNode(NodeBreak, "Break", breakID)
		b.cfg.AddEdge(parentID, breakID)
		return breakID

	case "continue_statement":
		continueID := b.newID()
		b.cfg.AddNode(NodeContinue, "Continue", continueID)
		b.cfg.AddEdge(parentID, continueID)
		return continueID

	case "name":
		return b.addGenericNode(NodeId, node, parentID)

	case "string_content":
		return b.addGenericNode(NodeString, node, parentID)

	case "variable_name":
		return b.addGenericNode(NodeVariable, node, parentID)

	case "integer":
		return b.addGenericNode(NodeInteger, node, parentID)

	case "string":
		return b.addGenericNode(NodeStringLiteral, node, parentID)

	default:
		// For unsupported nodes, just visit their children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			parentID = b.visit(child, parentID)
		}
		return parentID
	}
}

func (b *CFGBuilder) addGenericNode(nodeType string, node *sitter.Node, parentID int) int {
	strID := b.newID()
	b.cfg.AddNode(nodeType, node.Content(b.source), strID)
	b.cfg.AddEdge(parentID, strID)
	return strID
}

func (cfg *CFG) Print() {
	fmt.Println("=== Affichage du CFG ===")

	var ids []int
	for id := range cfg.Nodes {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		node := cfg.Nodes[id]
		succs := cfg.Edges[id]

		var succIDs []string
		for _, s := range succs {
			succIDs = append(succIDs, strconv.Itoa(s))
		}
		fmt.Printf("Node %d: %s [%q] -> [%s]\n", id, node.Type, node.code, strings.Join(succIDs, ", "))
	}

	fmt.Println("===========")
}
