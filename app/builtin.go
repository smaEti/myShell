package main

import (
	"fmt"
	"os"
	"slices"
)

var builtInCommands = []string{"exit", "cd", "pwd", "type", "echo"}

// executeBuiltIn handles execution of built-in commands
func executeBuiltIn(cmd *CommandNode) error {
	switch cmd.Name {
	case "exit":
		return handleExit(cmd)
	case "cd":
		return handleCd(cmd)
	case "pwd":
		return handlePwd(cmd)
	case "type":
		return handleType(cmd)
	case "echo":
		return handleEcho(cmd)
	default:
		return fmt.Errorf("unknown built-in: %s", cmd.Name)
	}
}

func handleExit(cmd *CommandNode) error {
	os.Exit(0)
	return nil
}

func handleCd(cmd *CommandNode) error {
	if len(cmd.Args) < 1 {
		fmt.Fprintln(cmd.Stderr, "cd: missing argument")
		return fmt.Errorf("missing argument")
	}

	path := cmd.Args[0]
	if path == "~" {
		path = os.Getenv("HOME")
	}

	err := os.Chdir(path)
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "cd: "+cmd.Args[0]+": No such file or directory")
		return err
	}
	return nil
}

func handlePwd(cmd *CommandNode) error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "An error occurred in reading the directory:", err)
		return err
	}
	fmt.Fprintln(cmd.Stdout, pwd)
	return nil
}

func handleType(cmd *CommandNode) error {
	if len(cmd.Args) < 1 {
		fmt.Fprintln(cmd.Stderr, "type: missing argument")
		return fmt.Errorf("missing argument")
	}

	command := cmd.Args[0]

	if slices.Contains(builtInCommands, command) {
		fmt.Fprintln(cmd.Stdout, command+" is a shell builtin")
		return nil
	}

	file := findExecutable(command)
	if file != "" {
		fmt.Fprintln(cmd.Stdout, command+" is "+file)
		return nil
	}

	fmt.Fprintln(cmd.Stdout, command+": not found")
	return nil
}

func handleEcho(cmd *CommandNode) error {
	for i, arg := range cmd.Args {
		if i > 0 {
			fmt.Fprint(cmd.Stdout, " ")
		}
		fmt.Fprint(cmd.Stdout, arg)
	}
	fmt.Fprintln(cmd.Stdout)
	return nil
}
