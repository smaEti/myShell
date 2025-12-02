package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for true {

		fmt.Print("$ ")
		commandArgs, command := readingCommand()

		if commandArgs[0] == "echo" {
			fmt.Println(command[4:])
			continue
		}
		if command == "exit" {
			os.Exit(0)
		}
		fmt.Println(command + ": command not found")
	}
}

func readingCommand() ([]string, string) {
	commandWithEndLine, err := bufio.NewReader(os.Stdin).ReadString('\n')
	command := commandWithEndLine[:len(commandWithEndLine)-1]
	if err != nil {
		fmt.Println("An error is eccoured")
		os.Exit(2)
	}
	command = strings.TrimSpace(command)
	commandFields := strings.Fields(command)
	return commandFields, command
}
