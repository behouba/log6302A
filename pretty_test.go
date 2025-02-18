package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func formatPHP(input string) (string, error) {
	printer := NewPrinter("    ")
	return printer.Format(input)
}

func TestPHPTag(t *testing.T) {
	input := " <?php echo \"Hello, World!\";?>"
	expected := "\necho \"Hello, World!\";\n"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestVariableDeclaration(t *testing.T) {
	input := "<?php $x=5;"
	expected := "$x = 5;"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestAssignment(t *testing.T) {
	input := "<?php $x=$y;"
	expected := "$x = $y;"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestArrayDeclaration(t *testing.T) {
	input := "<?php $arr=array(1,2,3,4,5,   6,           7,8);"
	expected := "$arr = array(1, 2, 3, 4, 5, 6, 7, 8);"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestFunctionDefinition(t *testing.T) {
	input := "<?php function test($param1,$param2){return $param1+$param2;}"
	expected := "function test($param1, $param2) {\n    return $param1 + $param2;\n}"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestIfStatement(t *testing.T) {
	input := `<?php if ($x>5) { echo "Greater"; } else { echo "Smaller"; }`
	expected := "if ($x > 5) {\n    echo \"Greater\";\n} else {\n    echo \"Smaller\";\n}"

	output, err := formatPHP(input)
	fmt.Println(output)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestWhileLoop(t *testing.T) {
	input := `<?php while($i<10){                  $i++; }`
	expected := "while ($i < 10) {\n    $i++;\n}"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

// FIXME: This test is failing
func TestForLoop(t *testing.T) {
	input := `<?php for ($i=0;$i<10;             $i++) { echo $i; }`
	expected := "for ($i = 0; $i < 10; $i++) {\n    echo $i;\n}"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

// FIXME: This test is failing
func TestForeachLoop(t *testing.T) {
	input := `<?php foreach ($arr as   $val) { echo $val; }`
	expected := "foreach ($arr as $val) {\n    echo $val;\n}"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

// FIXME: This test is failing
func TestSwitchCase(t *testing.T) {
	input := `<?php switch($var){case 1: echo "One"; break; default: echo "Default";}`
	expected := "switch ($var) {\n    case 1:\n        echo \"One\";\n        break;\n    default:\n        echo \"Default\";\n}"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}

func TestFunctionCall(t *testing.T) {
	input := `<?php
$a = 10;
if ( $a < 5) {
	echo " True ";
} else { 
 	echo " False ";
}`
	expected := "test(1, 2);"

	output, err := formatPHP(input)
	assert.NoError(t, err)
	assert.Contains(t, output, expected)
}
