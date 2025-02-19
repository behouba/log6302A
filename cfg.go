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

const Terminal = -1

type CFG struct {
	Nodes map[int]*CFGNode
	Edges map[int][]int
}

type CFGNode struct {
	ID   int
	Type string
	code string // info for debug
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
	if src != dst && src != Terminal {
		cfg.Edges[src] = append(cfg.Edges[src], dst)
	}
}

type stackEntry struct {
	typ   string // "while", "if", "for", "switch"
	start int    // Start node (Condition for loops, Entry for if)
	end   int    // End node (WhileEnd, IfEnd)
}

type depthStack struct {
	s []stackEntry
}

// Push a new control structure onto the stack
func (ds *depthStack) push(typ string, start, end int) {
	ds.s = append(ds.s, stackEntry{typ, start, end})
}

// Pop the top control structure from the stack
func (ds *depthStack) pop() {
	if len(ds.s) > 0 {
		ds.s = ds.s[:len(ds.s)-1]
	}
}

// Peek at the top control structure without removing it
func (ds *depthStack) len() int {
	return len(ds.s)
}

// Check if the stack is empty
func (ds *depthStack) isEmpty() bool {
	return len(ds.s) == 0
}

type CFGBuilder struct {
	parser *sitter.Parser
	cfg    *CFG
	nextID int
	source []byte
	depth  *depthStack
}

func NewCFGBuilder() *CFGBuilder {
	p := sitter.NewParser()
	p.SetLanguage(php.GetLanguage())
	return &CFGBuilder{
		parser: p,
		cfg:    NewCFG(),
		nextID: 1,
		depth:  &depthStack{},
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
	if lastNodeID != entryID && lastNodeID != Terminal {
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

	switch node.Type() {

	case "php_tag":
		return b.addGenericNode(NodeHtml, node, parentID)

	case "assignment_expression":
		lNode := node.Child(int(node.ChildCount()) - 1)
		lNodeID := b.visit(lNode, parentID)
		// Process first child.
		rNode := node.Child(0)
		rNodeID := b.visit(rNode, lNodeID)
		// Process any remaining children.
		seq := rNodeID
		for i := 1; i < int(node.ChildCount())-1; i++ {
			if seq == Terminal {
				_ = b.visit(node.Child(i), Terminal)
				continue
			}
			child := node.Child(i)
			res := b.visit(child, seq)
			if res == Terminal {
				seq = Terminal
			} else {
				seq = res
			}
		}
		return seq

	case "binary_expression":
		leftOperand := node.Child(0)
		leftID := b.visit(leftOperand, parentID)
		rightOperand := node.Child(2)
		rightID := b.visit(rightOperand, leftID)
		operatorNode := node.Child(1)
		operatorID := b.visit(operatorNode, rightID)
		return operatorID

	case "=", "+=", "-=", "*=", "/=":
		return b.addGenericNode(NodeBinOp, node, parentID)

	case "==", "!=", "<", "<=", ">", ">=":
		return b.addGenericNode(NodeRelOp, node, parentID)

	case "+", "-", "*", "/":
		return b.addGenericNode(NodeBinOp, node, parentID)

	case "if_statement":
		ifID := b.newID()
		b.cfg.AddNode(NodeIf, NodeIf, ifID)
		if parentID != Terminal {
			b.cfg.AddEdge(parentID, ifID)
		}

		conditionNode := node.ChildByFieldName("condition")
		conditionID := b.processCondition(conditionNode, ifID)

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
		if trueBranchID != Terminal {
			b.cfg.AddEdge(trueBranchID, ifEndID)
		}
		if falseBranchID != Terminal {
			b.cfg.AddEdge(falseBranchID, ifEndID)
		}

		// If both branches are terminal, then the sequential flow remains terminal.
		if trueBranchID == Terminal && falseBranchID == Terminal {
			return Terminal
		}
		return ifEndID

	case "echo_statement":
		echoID := b.newID()
		b.cfg.AddNode(NodeEcho, "Echo", echoID)
		if parentID != Terminal {
			b.cfg.AddEdge(parentID, echoID)
		}
		argNode := node.Child(1)
		if argNode != nil {
			argID := b.visit(argNode, echoID)
			return argID
		}
		return echoID

	case "function_call_expression":
		funcCallID := b.newID()
		b.cfg.AddNode(NodeFunctionCall, NodeFunctionCall, funcCallID)
		if parentID != Terminal {
			b.cfg.AddEdge(parentID, funcCallID)
		}

		funcNameNode := node.Child(0)
		funcNameID := b.newID()
		funcName := funcNameNode.Content(b.source)
		b.cfg.AddNode(NodeId, funcName, funcNameID)
		b.cfg.AddEdge(funcCallID, funcNameID)

		argumentsNode := node.ChildByFieldName("arguments")
		if argumentsNode != nil {
			argsID := b.newID()
			b.cfg.AddNode(NodeArgumentList, NodeArgumentList, argsID)
			b.cfg.AddEdge(funcNameID, argsID)

			seq := argsID
			for i := 0; i < int(argumentsNode.ChildCount()); i++ {
				argNode := argumentsNode.Child(i)
				if argNode.Type() != "(" && argNode.Type() != ")" {
					argumentID := b.newID()
					b.cfg.AddNode(NodeArgument, NodeArgument, argumentID)
					b.cfg.AddEdge(argsID, argumentID)

					res := b.visit(argNode, argumentID)
					if res != Terminal {
						seq = res
					}
				}
			}

			callBeginID := b.newID()
			b.cfg.AddNode(NodeCallBegin, funcName, callBeginID)
			b.cfg.AddEdge(seq, callBeginID)

			callEndID := b.newID()
			b.cfg.AddNode(NodeCallEnd, funcName, callEndID)
			b.cfg.AddEdge(callBeginID, callEndID)

			retValueID := b.newID()
			b.cfg.AddNode(NodeRetValue, NodeRetValue, retValueID)
			b.cfg.AddEdge(callEndID, retValueID)

			return retValueID
		}

		return funcCallID

	case "while_statement":
		whileID := b.newID()
		b.cfg.AddNode(NodeWhile, NodeWhile, whileID)
		if parentID != Terminal {
			b.cfg.AddEdge(parentID, whileID)
		}

		conditionNode := node.ChildByFieldName("condition")
		conditionID := b.processCondition(conditionNode, whileID)

		whileEndID := b.newID()
		b.depth.push(NodeWhile, whileID, whileEndID)

		bodyNode := node.ChildByFieldName("body")
		bodyID := b.visit(bodyNode, conditionID)

		// Only add back edge if the body did not terminate the sequential flow.
		if bodyID != Terminal {
			b.cfg.AddEdge(bodyID, whileID)
		}

		b.cfg.AddNode(NodeWhileEnd, NodeWhileEnd, whileEndID)
		b.cfg.AddEdge(conditionID, whileEndID)

		b.depth.pop()

		return whileEndID

	case "break_statement":
		breakID := b.newID()
		b.cfg.AddNode(NodeBreak, NodeBreak, breakID)
		if parentID != Terminal {
			b.cfg.AddEdge(parentID, breakID)
		}
		loopEndID := b.findClosestLoopEnd()
		b.cfg.AddEdge(breakID, loopEndID)
		return Terminal

	case "continue_statement":
		continueID := b.newID()
		b.cfg.AddNode(NodeContinue, NodeContinue, continueID)
		if parentID != Terminal {
			b.cfg.AddEdge(parentID, continueID)
		}
		whileConditionID := b.findClosestLoopCondition()
		b.cfg.AddEdge(continueID, whileConditionID)
		return Terminal

	case "compound_statement":
		// Assume the first and last children are "{" and "}".
		seq := parentID
		for i := 1; i < int(node.ChildCount())-1; i++ {
			child := node.Child(i)
			// If we are already in dead code, process the child without linking.
			if seq == Terminal {
				_ = b.visit(child, Terminal)
				continue
			}
			res := b.visit(child, seq)
			if res == Terminal {
				seq = Terminal
			} else {
				seq = res
			}
		}
		return seq

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
		// Process children sequentially.
		seq := parentID
		for i := 0; i < int(node.ChildCount()); i++ {
			if seq == Terminal {
				// Already in dead code: process without linking.
				_ = b.visit(node.Child(i), Terminal)
				continue
			}
			res := b.visit(node.Child(i), seq)
			if res == Terminal {
				seq = Terminal
			} else {
				seq = res
			}
		}
		return seq

	}
}

// func (b *CFGBuilder) isInsideBreakOrContinue(parentID int) bool {
// 	if b.depth.isEmpty() {
// 		return false
// 	}

// 	for id := parentID; id > 0; id-- {
// 		if node, exists := b.cfg.Nodes[id]; exists {
// 			if node.Type == NodeBreak || node.Type == NodeContinue {
// 				return true
// 			}
// 		}
// 	}

// 	return false
// }

func (b *CFGBuilder) findClosestLoopCondition() int {
	for i := b.depth.len() - 1; i >= 0; i-- {
		if b.depth.s[i].typ == NodeWhile || b.depth.s[i].typ == NodeFor {
			return b.depth.s[i].start
		}
	}
	return 1
}

func (b *CFGBuilder) findClosestLoopEnd() int {
	for i := b.depth.len() - 1; i >= 0; i-- {
		if b.depth.s[i].typ == NodeWhile || b.depth.s[i].typ == NodeFor {
			return b.depth.s[i].end
		}
	}
	return b.nextID
}

func (b *CFGBuilder) processCondition(node *sitter.Node, parentID int) int {
	if node == nil {
		return parentID
	}

	if node.ChildCount() == 3 && node.Child(1).Type() == "binary_expression" {
		node = node.Child(1)
	}

	leftOperand := node.Child(0)
	leftID := b.visit(leftOperand, parentID)

	rightOperand := node.Child(2)
	rightID := b.visit(rightOperand, leftID)

	operatorNode := node.Child(1)
	operatorID := b.visit(operatorNode, rightID)

	conditionID := b.newID()
	b.cfg.AddNode(NodeCondition, "Condition", conditionID)
	b.cfg.AddEdge(operatorID, conditionID)

	return conditionID
}

func (b *CFGBuilder) addGenericNode(nodeType string, node *sitter.Node, parentID int) int {
	strID := b.newID()
	b.cfg.AddNode(nodeType, node.Content(b.source), strID)
	if parentID != Terminal {
		b.cfg.AddEdge(parentID, strID)
	}
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
		fmt.Printf("Node %d: %s [%s] -> [%s]\n", id, node.Type, node.code, strings.Join(succIDs, ", "))
	}

	fmt.Println("===========")
}

// DetectDeadCode performs a reachability analysis from the Entry node (assumed to be node 1).
// It returns a slice of node IDs that are unreachable.
func (cfg *CFG) DetectDeadCode() []int {
	visited := make(map[int]bool)
	queue := []int{1} // assuming node 1 is the Entry

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		visited[id] = true
		for _, succ := range cfg.Edges[id] {
			if !visited[succ] {
				queue = append(queue, succ)
			}
		}
	}

	var dead []int
	for id := range cfg.Nodes {
		if !visited[id] {
			dead = append(dead, id)
		}
	}
	return dead
}

func (cfg *CFG) PrintDeadCode() {
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
		fmt.Printf("Node %d: %s [%s] -> [%s]\n", id, node.Type, node.code, strings.Join(succIDs, ", "))
	}
	fmt.Println("===========")
}
