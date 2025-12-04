package main

import (
	"fmt"
	"os"
)

func main() {
	for {
		fmt.Print("$ ")

		// Read and split input into words
		words := lexingCommand()
		if len(words) == 0 {
			continue
		}

		// Classify words into token types
		tokens := tokenize(words)

		// Parse - Build AST from tokens
		ast, err := parse(tokens)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			continue
		}

		// Execute - Run the AST
		if err := ast.Execute(); err != nil {
			// Errors are typically already printed by the command
			continue
		}
	}
}
