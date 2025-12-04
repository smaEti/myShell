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
	"unicode"
)

var builtIn = []string{"type", "echo", "exit", "pwd", "cd"}

func main() {

	for {
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

		// Handle escape characters
		if r == '\\' && i+1 < len(command) {
			next := rune(command[i+1])
			switch next {
			case '\\': // double backslash
				sb.WriteRune('\\')
			case '"', '\'': // escaped quote
				sb.WriteRune(next)
			case 'n':
				sb.WriteRune('n')
			case 't':
				sb.WriteRune('t')
			case ' ':
				sb.WriteString("  ")
			default:
				// keep the backslash if it doesn't form a known escape
				sb.WriteRune(next)
			}
			i++ // skip next char
			continue
		}

		// Handle quoting logic
		switch {
		case r == '\'' || r == '"':
			if inQuotes {
				if r == quoteChar {
					inQuotes = false // closing quote
				} else {
					sb.WriteRune(r) // quote inside other type of quotes
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

	return result, command
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
func handleCdCommand(commandArgs []string) {
	if commandArgs[1] == "~" {
		commandArgs[1] = os.Getenv("HOME")
	}
	err := os.Chdir(commandArgs[1])
	if err != nil {
		fmt.Println("cd: " + commandArgs[1] + ": No such file or directory")
	}
}
func handleEchoCommand(commandArgs []string) {
	for _, arg := range commandArgs {
		// if strings.Contains(arg, "\\") {
		// 	var sb strings.Builder
		// 	for i := 0; i < len(arg); i++ {
		// 		r := arg[i]

		// 		if r != '\\' {
		// 			sb.WriteByte(r)
		// 			continue
		// 		}

		// 		if i+1 >= len(arg) {
		// 			break
		// 		}

		// 		next := arg[i+1]
		// 		if next == ' ' {
		// 			sb.WriteByte(' ')
		// 			i++
		// 			continue
		// 		}
		// 		if next == '\\' {
		// 			sb.WriteByte('\\')
		// 			i++
		// 			continue
		// 		}
		// 	}
		// 	arg = sb.String()
		// }
		fmt.Print(arg + " ")
	}
	fmt.Println()

}
func handleBuiltInCommands(command string, commandArgs []string) {
	if commandArgs[0] == "cd" {
		handleCdCommand(commandArgs)
	}
	if commandArgs[0] == "pwd" {
		handlePwdCommand()
	}

	if commandArgs[0] == "type" {
		handleTypeCommand(commandArgs[1])
	}

	if commandArgs[0] == "echo" {
		handleEchoCommand(commandArgs[1:])
	}

	if command == "exit" {
		os.Exit(0)
	}
}
