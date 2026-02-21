const { analyzeWorkspace } = require('./backend/main.wasm');

const workspace = {
  Files: [
    {
      Filename: "goproject/main.go",
      Content: `package main

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

var unusedGlobal = 100 // This should be detected as unused`,
      Hash: "hash1"
    },
    {
      Filename: "goproject/utils/utils.go",
      Content: `package utils

import "fmt"

func PrintHello() {
	fmt.Println("Hello from utils!")
}

func unusedFunction() {
	fmt.Println("This function is never called")
}

var unusedGlobal = 100 // This should be detected as unused`,
      Hash: "hash2"
    },
    {
      Filename: "goproject/feature.go",
      Content: `package main

import (
	"fmt"
	"goproject/utils"
)

func main() {
	fmt.Println("Starting feature...")
	utils.PrintHello()
	unusedVariable := 42 // This should be detected as unused
	fmt.Println("Feature started!")
}

func unusedFunction() {
	fmt.Println("This function is never called")
}

var unusedGlobal = 100 // This should be detected as unused`,
      Hash: "hash3"
    }
  ]
};

const result = analyzeWorkspace(JSON.stringify(workspace));
console.log(JSON.stringify(JSON.parse(result), null, 2));