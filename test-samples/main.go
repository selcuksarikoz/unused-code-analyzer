package main

import (
	"fmt"
	"goproject/utils"
)

func main() {
	fmt.Println("Starting application...")
	utils.PrintHello()

	// Unused variables
	unusedVariable := 42
	unusedVarWithType := int(42)
	_, _, _ = 1, 2, 3
	unusedShortVar := 42

	// Used variables
	usedVariable := 42
	fmt.Println(usedVariable)

	// Unused imports
	unusedImport := "unused"

	// Unused function parameters
	funcWithUnusedParam(42, "test", true)

	// Unused return values
	_, _, _ = getMultipleReturns()

	// Unused global variables
	unusedGlobal := 100

	fmt.Println("Application started!")
}

func funcWithUnusedParam(a int, b string, unused bool) {
	fmt.Println(a, b)
}

func getMultipleReturns() (int, string, bool) {
	return 1, "test", true
}

func unusedFunction() {
	fmt.Println("This function is never called")
}

var (
	unusedGlobal1 = 100
	unusedGlobal2 = "test"
	usedGlobal    = 200
)

func init() {
	fmt.Println(usedGlobal)
}
