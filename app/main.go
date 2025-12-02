package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
)

func main() {
	for true {

		fmt.Print("$ ")
		commandArgs, command := readingCommand()

		if commandArgs[0] == "type" {
			handleTypeCommand(commandArgs[1])
			continue
		}

		if commandArgs[0] == "echo" {
			fmt.Println(command[5:])
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
		fmt.Println("An error is eccoured in reading the commandline", err)
		os.Exit(2)
	}
	command = strings.TrimSpace(command)
	commandFields := strings.Fields(command)
	return commandFields, command
}

func handleTypeCommand(command string) {
	builtIn := []string{"type", "echo", "exit"}
	if slices.Contains(builtIn, command) {
		fmt.Println(command + " is a shell builtin")
		return
	}
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
			isFileExecutable := IsExecutable(fullPath)
			if isFileExecutable {
				fmt.Println(command + " is " + fullPath)
				return
			}
		}
	}
	fmt.Println(command + ": not found")
}

func IsExecutable(path string) bool {
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

	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".exe", ".bat", ".cmd", ".com", ".ps1":
			return true
		}
	}

	return false
}
