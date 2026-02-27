package main

import (
	"fmt"
)

func main() {
	fmt.Println("Test")
	unusedVariable := 42
	fmt.Println(unusedVariable)
}

func unusedFunction() {
	fmt.Println("This function is never called")
}

var unusedGlobal = 100
