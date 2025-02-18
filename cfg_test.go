package main

import (
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
	assert.Equal(t, []int{10, 12}, cfg.Edges[9], "Condition should branch to true/false paths")
	assert.Equal(t, []int{14}, cfg.Edges[11], "Echo True should lead to IfEnd")
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

	// Ensure correct edges (execution flow)
	assert.Equal(t, []int{2}, cfg.Edges[1], "Entry should connect to FunctionCall")
	assert.Equal(t, []int{3}, cfg.Edges[2], "FunctionCall should connect to function name")
	assert.Equal(t, []int{4}, cfg.Edges[3], "Function name should connect to ArgumentList")
	assert.Equal(t, []int{5}, cfg.Edges[4], "ArgumentList should connect to Argument")
	assert.Equal(t, []int{6}, cfg.Edges[5], "Argument should connect to StringLiteral")
	assert.Equal(t, []int{7}, cfg.Edges[6], "StringLiteral should connect to CallBegin")
	assert.Equal(t, []int{8}, cfg.Edges[7], "CallBegin should connect to CallEnd")
	assert.Equal(t, []int{9}, cfg.Edges[8], "CallEnd should connect to RetValue")
	assert.Equal(t, []int{10}, cfg.Edges[9], "RetValue should connect to Exit")

	// Print CFG for debugging
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

	// Create a new CFG builder and generate the CFG
	builder := NewCFGBuilder()
	cfg, err := builder.BuildCFG([]byte(phpCode))

	// Ensure no errors
	assert.NoError(t, err, "CFG generation should not return an error")

	// Ensure key nodes exist
	// assert.Contains(t, cfg.Nodes, 1, "Entry node should exist")
	// assert.Contains(t, cfg.Nodes, 2, "Assignment node ($i = 0) should exist")
	// assert.Contains(t, cfg.Nodes, 4, "While condition node ($i < 10) should exist")
	// assert.Contains(t, cfg.Nodes, 6, "Increment node ($i = $i + 1) should exist")
	// assert.Contains(t, cfg.Nodes, 8, "Condition node ($i == 5) should exist")
	// assert.Contains(t, cfg.Nodes, 9, "Break statement should exist")
	// assert.Contains(t, cfg.Nodes, 10, "Continue statement should exist")
	// assert.Contains(t, cfg.Nodes, 12, "Unreachable Echo 'Dead' should exist")
	// assert.Contains(t, cfg.Nodes, 14, "Echo 'Done' should exist")
	// assert.Contains(t, cfg.Nodes, 15, "Exit node should exist")

	// Ensure correct edges (execution flow)
	// assert.Equal(t, []int{2}, cfg.Edges[1], "Entry should connect to variable assignment")
	// assert.Equal(t, []int{4}, cfg.Edges[2], "Assignment should lead to while condition")
	// assert.Equal(t, []int{6, 15}, cfg.Edges[4], "While should connect to loop body and exit")
	// assert.Equal(t, []int{8}, cfg.Edges[6], "Increment should connect to if condition")
	// assert.Equal(t, []int{9, 10}, cfg.Edges[8], "If condition should branch to break or continue")
	// assert.Equal(t, []int{15}, cfg.Edges[9], "Break should jump to exit")
	// assert.Equal(t, []int{4}, cfg.Edges[10], "Continue should go back to while condition")
	// assert.Equal(t, []int{14}, cfg.Edges[12], "Echo 'Dead' (unreachable) should exist but not execute")
	// assert.Equal(t, []int{15}, cfg.Edges[14], "Echo 'Done' should connect to exit")

	// Print CFG for debugging
	cfg.Print()
}
