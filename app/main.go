package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	for true {

		fmt.Print("$ ")

		commandWithEndLine, err := bufio.NewReader(os.Stdin).ReadString('\n')
		command := commandWithEndLine[:len(commandWithEndLine)-1]
		if err != nil {
			fmt.Println("An error is eccoured")
		}

		if command == "exit" {
			os.Exit(0)
		}
		fmt.Println(command + ": command not found")
	}
}
