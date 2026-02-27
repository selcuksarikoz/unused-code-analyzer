package main

import (
	"fmt"
	"goproject/utils"
)

func main() {
	fmt.Println("Starting application...")
	utils.PrintHello()
	unusedVariable := 42 // This should be detected as unused
	fmt.Println("Application started!")
}

func unusedFunction() {
	fmt.Println("This function is never called")
}

var unusedGlobal = 100 // This should be detected as unused
