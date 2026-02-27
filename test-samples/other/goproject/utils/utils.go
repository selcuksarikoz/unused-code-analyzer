package utils

import "fmt"

func PrintHello() {
	fmt.Println("Hello from utils!")
}

func unusedFunction() {
	fmt.Println("This function is never called")
}

var unusedGlobal = 100 // This should be detected as unused
