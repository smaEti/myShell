package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode"
)

var builtIn = []string{"type", "echo", "exit", "pwd", "cd"}

type executableCommand struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Name   string
	Args   []string
}

func main() {
	for {
		fmt.Print("$ ")
		commandArgs := readingCommand()

		if len(commandArgs) == 0 {
			continue
		}

		cmd := &executableCommand{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  os.Stdin,
			Name:   commandArgs[0],
			Args:   commandArgs,
		}

		cmd.Execute()
	}
}

// Execute determines whether to run a built-in or external command
func (cmd *executableCommand) Execute() error {
	// Handle built-in commands
	if slices.Contains(builtIn, cmd.Name) {
		return cmd.executeBuiltIn()
	}

	// Handle external executables
	executableFile := findExecutable(cmd.Name)
	if executableFile != "" {
		return cmd.executeExternal()
	}

	// Command not found
	fmt.Fprintln(cmd.Stderr, cmd.Name+": command not found")
	return fmt.Errorf("command not found: %s", cmd.Name)
}

// executeBuiltIn handles built-in commands
func (cmd *executableCommand) executeBuiltIn() error {
	switch cmd.Name {
	case "exit":
		return cmd.handleExit()
	case "cd":
		return cmd.handleCd()
	case "pwd":
		return cmd.handlePwd()
	case "type":
		return cmd.handleType()
	case "echo":
		return cmd.handleEcho()
	default:
		return fmt.Errorf("unknown built-in: %s", cmd.Name)
	}
}

// executeExternal handles external executable commands
func (cmd *executableCommand) executeExternal() error {
	var args []string
	if len(cmd.Args) > 1 {
		args = cmd.Args[1:]
	}

	externalCmd := exec.Command(cmd.Name, args...)
	externalCmd.Stdout = cmd.Stdout
	externalCmd.Stderr = cmd.Stderr
	externalCmd.Stdin = cmd.Stdin

	return externalCmd.Run()
}

// Built-in command handlers
func (cmd *executableCommand) handleExit() error {
	os.Exit(0)
	return nil
}

func (cmd *executableCommand) handleCd() error {
	if len(cmd.Args) < 2 {
		fmt.Fprintln(cmd.Stderr, "cd: missing argument")
		return fmt.Errorf("missing argument")
	}

	path := cmd.Args[1]
	if path == "~" {
		path = os.Getenv("HOME")
	}

	err := os.Chdir(path)
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "cd: "+cmd.Args[1]+": No such file or directory")
		return err
	}
	return nil
}

func (cmd *executableCommand) handlePwd() error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "An error occurred in reading the directory:", err)
		return err
	}
	fmt.Fprintln(cmd.Stdout, pwd)
	return nil
}

func (cmd *executableCommand) handleType() error {
	if len(cmd.Args) < 2 {
		fmt.Fprintln(cmd.Stderr, "type: missing argument")
		return fmt.Errorf("missing argument")
	}

	command := cmd.Args[1]

	if slices.Contains(builtIn, command) {
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

func (cmd *executableCommand) handleEcho() error {
	if len(cmd.Args) > 1 {
		for i, arg := range cmd.Args[1:] {
			if i > 0 {
				fmt.Fprint(cmd.Stdout, " ")
			}
			fmt.Fprint(cmd.Stdout, arg)
		}
	}
	fmt.Fprintln(cmd.Stdout)
	return nil
}

func readingCommand() []string {
	reader := bufio.NewReader(os.Stdin)
	commandWithEndLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occurred while reading the command line:", err)
		os.Exit(2)
	}
	command := strings.TrimSpace(commandWithEndLine)
	inQuotes := false
	var quoteChar rune
	var sb strings.Builder
	var result []string

	for i := 0; i < len(command); i++ {
		r := rune(command[i])

		if r == '\\' && i+1 < len(command) {
			next := rune(command[i+1])
			if inQuotes && quoteChar == '"' {
				if next == '\\' || next == '"' {
					sb.WriteRune(next)
					i++
					continue
				}
				sb.WriteRune(r)
				continue
			} else if inQuotes && quoteChar == '\'' {
				sb.WriteRune(r)
				continue
			} else {
				sb.WriteRune(next)
				i++
				continue
			}
		}

		switch {
		case r == '\'' || r == '"':
			if inQuotes {
				if r == quoteChar {
					inQuotes = false
				} else {
					sb.WriteRune(r)
				}
			} else {
				inQuotes = true
				quoteChar = r
			}
		case unicode.IsSpace(r):
			if inQuotes {
				sb.WriteRune(r)
			} else if sb.Len() > 0 {
				result = append(result, sb.String())
				sb.Reset()
			}
		default:
			sb.WriteRune(r)
		}
	}

	if sb.Len() > 0 {
		result = append(result, sb.String())
	}

	return result
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
