package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

var builtIn = []string{"type", "echo", "exit", "pwd", "cd"}

func main() {

	for true {
		fmt.Print("$ ")
		commandArgs, command := readingCommand()
		if slices.Contains(builtIn, commandArgs[0]) {
			handleBuiltInCommands(command, commandArgs)
			continue
		}

		executableFile := findExecutable(commandArgs[0])
		if executableFile != "" {
			cmd := exec.Command(commandArgs[0], commandArgs[1:]...)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Run()
			continue
		} else {
			fmt.Println(commandArgs[0] + ": command not found")
			continue
		}
	}
}

func readingCommand() ([]string, string) {
	commandWithEndLine, err := bufio.NewReader(os.Stdin).ReadString('\n')
	command := commandWithEndLine[:len(commandWithEndLine)-1]
	if err != nil {
		fmt.Println("An error is eccoured in reading the commandline", err)
		os.Exit(2)
	}
	command = strings.TrimSpace(command)
	commandFields := strings.Fields(command)

	return commandFields, command
}

func handleTypeCommand(command string) {
	if slices.Contains(builtIn, command) {
		fmt.Println(command + " is a shell builtin")
		return
	}
	//locate executables
	file := findExecutable(command)
	if file != "" {
		fmt.Println(command + " is " + file)
		return
	}
	fmt.Println(command + ": not found")
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	mode := info.Mode()

	if mode&0111 != 0 {
		return true
	}

	return false
}

func findExecutable(command string) string {
	pathString := os.Getenv("PATH")
	paths := strings.Split(pathString, string(os.PathListSeparator))

	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		index := sort.Search(len(files), func(i int) bool {
			return files[i].Name() >= command
		})

		if index < len(files) && files[index].Name() == command {
			fullPath := filepath.Join(path, command)
			isFileExecutable := isExecutable(fullPath)
			if isFileExecutable {
				return fullPath
			}
		}
	}
	return ""
}
func handlePwdCommand() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("An error is eccoured in reading the commandline", err)
		os.Exit(2)
	}
	fmt.Println(pwd)
}
func handleBuiltInCommands(command string, commandArgs []string) {
	if commandArgs[0] == "cd" {
		err := os.Chdir(commandArgs[1])
		if err != nil {
			fmt.Println("cd: " + commandArgs[1] + " No such file or directory")
		}
	}
	if commandArgs[0] == "pwd" {
		handlePwdCommand()
	}

	if commandArgs[0] == "type" {
		handleTypeCommand(commandArgs[1])
	}

	if commandArgs[0] == "echo" {
		fmt.Println(command[5:])
	}

	if command == "exit" {
		os.Exit(0)
	}
}
