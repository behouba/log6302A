package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCFGBuilderOnIFStatement(t *testing.T) {
	phpCode := `<?php
	$a = 10;
	if ($a < 5) {
		echo "True";
	} else {
		echo "False";
	}`

	// Create a new CFG builder and generate the CFG
	builder := NewCFGBuilder()
	cfg, err := builder.BuildCFG([]byte(phpCode))

	// Ensure no errors
	assert.NoError(t, err, "CFG generation should not return an error")

	// Ensure key nodes exist
	assert.Contains(t, cfg.Nodes, 1, "Entry node should exist")
	assert.Contains(t, cfg.Nodes, 4, "Assignment node should exist")
	assert.Contains(t, cfg.Nodes, 9, "Condition node should exist")
	assert.Contains(t, cfg.Nodes, 8, "RelOp node should exist")
	assert.Contains(t, cfg.Nodes, 10, "Echo True should exist")
	assert.Contains(t, cfg.Nodes, 12, "Echo False should exist")
	assert.Contains(t, cfg.Nodes, 14, "IfEnd should exist")
	assert.Contains(t, cfg.Nodes, 15, "Exit node should exist")

	cfg.Print()
	// Ensure correct edges (execution flow)
	assert.Equal(t, []int{2}, cfg.Edges[1], "Entry should connect to variable assignment")
	assert.Equal(t, []int{3}, cfg.Edges[2], "Variable should connect to integer 10")
	assert.Equal(t, []int{5}, cfg.Edges[4], "Assignment should connect to if_statement")
	assert.Equal(t, []int{7}, cfg.Edges[6], "Variable in condition should connect to integer 5")
	assert.Equal(t, []int{9}, cfg.Edges[8], "RelOp should connect to condition node")
	assert.Equal(t, []int{11, 13}, cfg.Edges[10], "Condition should branch to true/false paths")
	assert.Equal(t, []int{12}, cfg.Edges[11], "Echo True should lead to IfEnd")
	assert.Equal(t, []int{14}, cfg.Edges[13], "Echo False should lead to IfEnd")
	assert.Equal(t, []int{15}, cfg.Edges[14], "IfEnd should connect to Exit")
}

func TestCFGONFunctionCall(t *testing.T) {
	phpCode := `<?php

mysql_query('SELECT *');`

	// Create a new CFG builder and generate the CFG
	builder := NewCFGBuilder()
	cfg, err := builder.BuildCFG([]byte(phpCode))

	// Ensure no errors
	assert.NoError(t, err, "CFG generation should not return an error")

	// Ensure key nodes exist
	assert.Contains(t, cfg.Nodes, 1, "Entry node should exist")
	assert.Contains(t, cfg.Nodes, 2, "FunctionCall node should exist")
	assert.Contains(t, cfg.Nodes, 3, "Function name node should exist")
	assert.Contains(t, cfg.Nodes, 4, "ArgumentList node should exist")
	assert.Contains(t, cfg.Nodes, 5, "Argument node should exist")
	assert.Contains(t, cfg.Nodes, 6, "StringLiteral node should exist")
	assert.Contains(t, cfg.Nodes, 7, "CallBegin node should exist")
	assert.Contains(t, cfg.Nodes, 8, "CallEnd node should exist")
	assert.Contains(t, cfg.Nodes, 9, "RetValue node should exist")
	assert.Contains(t, cfg.Nodes, 10, "Exit node should exist")

	// Check correct edges
	assert.Equal(t, []int{2}, cfg.Edges[1], "Entry should connect to FunctionCall")
	assert.Equal(t, []int{3}, cfg.Edges[2], "FunctionCall should connect to function name")
	assert.Equal(t, []int{4}, cfg.Edges[3], "Function name should connect to ArgumentList")
	assert.Equal(t, []int{5}, cfg.Edges[4], "ArgumentList should connect to Argument")
	assert.Equal(t, []int{6}, cfg.Edges[5], "Argument should connect to StringLiteral")
	assert.Equal(t, []int{7}, cfg.Edges[6], "StringLiteral should connect to CallBegin")
	assert.Equal(t, []int{8}, cfg.Edges[7], "CallBegin should connect to CallEnd")
	assert.Equal(t, []int{9}, cfg.Edges[8], "CallEnd should connect to RetValue")
	assert.Equal(t, []int{10}, cfg.Edges[9], "RetValue should connect to Exit")

	// debugging
	cfg.Print()
}

func TestCFGOnWhileLoop(t *testing.T) {
	phpCode := `<?php
	$i = 0;

	while($i < 10) {
		$i = $i + 1;
		if($i == 5)
			break;
		continue;
		echo "Dead";
	}

	echo "Done";`

	builder := NewCFGBuilder()
	cfg, err := builder.BuildCFG([]byte(phpCode))
	assert.NoError(t, err, "CFG generation should not return an error")

	// Print CFG
	cfg.Print()

	// Node 1: Entry [Entry] -> [2]
	assert.Equal(t, "Entry", cfg.Nodes[1].Type)
	assert.Equal(t, []int{2}, cfg.Edges[1])

	// Node 2: Html [<?php] -> [3]
	assert.Equal(t, "Html", cfg.Nodes[2].Type)
	assert.Equal(t, []int{3}, cfg.Edges[2])

	// Node 3: Integer [0] -> [4]
	assert.Equal(t, "Integer", cfg.Nodes[3].Type)
	assert.Equal(t, "0", cfg.Nodes[3].code)
	assert.Equal(t, []int{4}, cfg.Edges[3])

	// Node 4: Variable [$i] -> [5]
	assert.Equal(t, "Variable", cfg.Nodes[4].Type)
	assert.Equal(t, "$i", cfg.Nodes[4].code)
	assert.Equal(t, []int{5}, cfg.Edges[4])

	// Node 5: BinOP [=] -> [6]
	assert.Equal(t, "BinOP", cfg.Nodes[5].Type)
	assert.Equal(t, "=", cfg.Nodes[5].code)
	assert.Equal(t, []int{6}, cfg.Edges[5])

	// Node 6: While [While] -> [7]
	assert.Equal(t, "While", cfg.Nodes[6].Type)
	assert.Equal(t, []int{7}, cfg.Edges[6])

	// Node 7: Variable [$i] -> [8]
	assert.Equal(t, "Variable", cfg.Nodes[7].Type)
	assert.Equal(t, "$i", cfg.Nodes[7].code)
	assert.Equal(t, []int{8}, cfg.Edges[7])

	// Node 8: Integer [10] -> [9]
	assert.Equal(t, "Integer", cfg.Nodes[8].Type)
	assert.Equal(t, "10", cfg.Nodes[8].code)
	assert.Equal(t, []int{9}, cfg.Edges[8])

	// Node 9: RelOP [<] -> [10]
	assert.Equal(t, "RelOP", cfg.Nodes[9].Type)
	assert.Equal(t, "<", cfg.Nodes[9].code)
	assert.Equal(t, []int{10}, cfg.Edges[9])

	// Node 10: Condition [Condition] -> [12, 11] (true branch: loop body, false branch: loop exit)
	assert.Equal(t, "Condition", cfg.Nodes[10].Type)
	assert.ElementsMatch(t, []int{12, 11}, cfg.Edges[10])

	// Node 11: WhileEnd [WhileEnd] -> [27] (exit from loop goes to echo "Done" branch)
	assert.Equal(t, "WhileEnd", cfg.Nodes[11].Type)
	assert.Equal(t, []int{27}, cfg.Edges[11])

	// --- Inside the while loop body ---
	// Node 12: Variable [$i] -> [13]
	assert.Equal(t, "Variable", cfg.Nodes[12].Type)
	assert.Equal(t, "$i", cfg.Nodes[12].code)
	assert.Equal(t, []int{13}, cfg.Edges[12])

	// Node 13: Integer [1] -> [14]
	assert.Equal(t, "Integer", cfg.Nodes[13].Type)
	assert.Equal(t, "1", cfg.Nodes[13].code)
	assert.Equal(t, []int{14}, cfg.Edges[13])

	// Node 14: BinOP [+] -> [15]
	assert.Equal(t, "BinOP", cfg.Nodes[14].Type)
	assert.Equal(t, "+", cfg.Nodes[14].code)
	assert.Equal(t, []int{15}, cfg.Edges[14])

	// Node 15: Variable [$i] -> [16]
	assert.Equal(t, "Variable", cfg.Nodes[15].Type)
	assert.Equal(t, "$i", cfg.Nodes[15].code)
	assert.Equal(t, []int{16}, cfg.Edges[15])

	// Node 16: BinOP [=] -> [17]
	assert.Equal(t, "BinOP", cfg.Nodes[16].Type)
	assert.Equal(t, "=", cfg.Nodes[16].code)
	assert.Equal(t, []int{17}, cfg.Edges[16])

	// Node 17: If [If] -> [18]
	assert.Equal(t, "If", cfg.Nodes[17].Type)
	assert.Equal(t, []int{18}, cfg.Edges[17])

	// Node 18: Variable [$i] -> [19]
	assert.Equal(t, "Variable", cfg.Nodes[18].Type)
	assert.Equal(t, "$i", cfg.Nodes[18].code)
	assert.Equal(t, []int{19}, cfg.Edges[18])

	// Node 19: Integer [5] -> [20]
	assert.Equal(t, "Integer", cfg.Nodes[19].Type)
	assert.Equal(t, "5", cfg.Nodes[19].code)
	assert.Equal(t, []int{20}, cfg.Edges[19])

	// Node 20: RelOP [==] -> [21]
	assert.Equal(t, "RelOP", cfg.Nodes[20].Type)
	assert.Equal(t, "==", cfg.Nodes[20].code)
	assert.Equal(t, []int{21}, cfg.Edges[20])

	// Node 21: Condition [Condition] -> [22, 23] (true branch: break; false branch: continue)
	assert.Equal(t, "Condition", cfg.Nodes[21].Type)
	assert.ElementsMatch(t, []int{22, 23}, cfg.Edges[21])

	// Node 22: Break [Break] -> [11] (jumps to loop exit)
	assert.Equal(t, "Break", cfg.Nodes[22].Type)
	assert.Equal(t, []int{11}, cfg.Edges[22])

	// Node 23: IfEnd [IfEnd] -> [24]
	assert.Equal(t, "IfEnd", cfg.Nodes[23].Type)
	assert.Equal(t, []int{24}, cfg.Edges[23])

	// Node 24: Continue [Continue] -> [6] (loops back to the while condition)
	assert.Equal(t, "Continue", cfg.Nodes[24].Type)
	assert.Equal(t, []int{6}, cfg.Edges[24])

	// --- Dead code inside the loop body (after the continue) ---
	// Node 25: Echo [Echo] -> [26] (this branch should not be reachable)
	assert.Equal(t, "Echo", cfg.Nodes[25].Type)
	assert.Equal(t, []int{26}, cfg.Edges[25])
	// Node 26: String [Dead] -> [] (dead code echo argument)
	assert.Equal(t, "String", cfg.Nodes[26].Type)
	assert.Equal(t, "Dead", cfg.Nodes[26].code)
	assert.Empty(t, cfg.Edges[26])

	// --- After the loop ---
	// Node 27: Echo [Echo] -> [28] (live branch echo for "Done")
	assert.Equal(t, "Echo", cfg.Nodes[27].Type)
	assert.Equal(t, []int{28}, cfg.Edges[27])
	// Node 28: String [Done] -> [29]
	assert.Equal(t, "String", cfg.Nodes[28].Type)
	assert.Equal(t, "Done", cfg.Nodes[28].code)
	assert.Equal(t, []int{29}, cfg.Edges[28])
	// Node 29: Exit [Exit] -> []
	assert.Equal(t, "Exit", cfg.Nodes[29].Type)
	assert.Empty(t, cfg.Edges[29])

	// --- Verify dead code detection ---
	// The dead chain (nodes 25 and 26) should not be reachable from the entry.
	deadNodes := cfg.DetectDeadCode()
	assert.Contains(t, deadNodes, 25, "Dead Echo node should be unreachable")
	assert.Contains(t, deadNodes, 26, "Dead string node should be unreachable")
}

// TestDetectDeadCode tests dead code detection on a given PHP script.
func TestDetectDeadCode(t *testing.T) {
	source := []byte(`<?php
		$i = 0;

		while($i < 10) {
			$i = $i + 1;
			if($i == 5)
				break;
			continue;
			echo "Dead";
		}

		echo "Done";
	`)

	builder := NewCFGBuilder()
	cfg, err := builder.BuildCFG(source)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}

	deadNodes := cfg.DetectDeadCode()

	var foundEcho, foundDead bool
	for _, id := range deadNodes {
		node, exists := cfg.Nodes[id]
		if !exists {
			continue
		}

		if node.Type == "Echo" {
			foundEcho = true
			t.Logf("Found dead code node (Echo): %d [%s]", node.ID, node.code)
		}

		if strings.Contains(node.code, "Dead") {
			foundDead = true
			t.Logf("Found dead code node (Dead literal): %d [%s]", node.ID, node.code)
		}
	}

	if !foundEcho || !foundDead {
		t.Errorf("Expected dead code chain (Echo and 'Dead') not fully detected; foundEcho=%v, foundDead=%v", foundEcho, foundDead)
	}
}
