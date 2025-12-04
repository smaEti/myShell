package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"sync"
)

// Node is the interface that all AST nodes implement
type Node interface {
	Execute() error
}

// CommandNode represents a single command with its arguments
type CommandNode struct {
	Name   string
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// PipeNode represents a pipe between two commands
type PipeNode struct {
	Left  Node
	Right Node
}

// RedirectNode represents I/O redirection
type RedirectNode struct {
	Command      Node
	RedirectType TokenType
	Filename     string
}

// Execute runs a command node
func (n *CommandNode) Execute() error {
	// Handle built-in commands
	if slices.Contains(builtInCommands, n.Name) {
		// Built-ins don't naturally consume stdin like external commands
		// If stdin is redirected (not from terminal), drain it to prevent deadlock
		if n.Stdin != os.Stdin {
			// Launch background goroutine to drain stdin
			// This prevents the writer from blocking
			go func() {
				io.Copy(io.Discard, n.Stdin)
			}()
		}
		return executeBuiltIn(n)
	}

	// Handle external commands
	executableFile := findExecutable(n.Name)
	if executableFile == "" {
		fmt.Fprintln(n.Stderr, n.Name+": command not found")
		return fmt.Errorf("command not found: %s", n.Name)
	}

	cmd := exec.Command(n.Name, n.Args...)
	cmd.Stdin = n.Stdin
	cmd.Stdout = n.Stdout
	cmd.Stderr = n.Stderr

	return cmd.Run()
}

// Execute runs a pipe node (concurrent execution)
func (n *PipeNode) Execute() error {
	// Create pipe
	pipeReader, pipeWriter := io.Pipe()

	// Setup left command to write to pipe
	if err := setNodeOutput(n.Left, pipeWriter); err != nil {
		return err
	}

	// Setup right command to read from pipe
	if err := setNodeInput(n.Right, pipeReader); err != nil {
		return err
	}

	// Execute both commands concurrently
	var wg sync.WaitGroup
	var leftErr, rightErr error

	wg.Add(2)

	// Execute left side
	go func() {
		defer wg.Done()
		defer pipeWriter.Close()
		leftErr = n.Left.Execute()
	}()

	// Execute right side
	go func() {
		defer wg.Done()
		rightErr = n.Right.Execute()
	}()

	wg.Wait()

	// Return first error encountered
	if leftErr != nil {
		return leftErr
	}
	return rightErr
}

// Execute runs a redirect node
func (n *RedirectNode) Execute() error {
	var file *os.File
	var err error

	// Open file based on redirect type
	switch n.RedirectType {
	case REDIRECT_IN:
		file, err = os.Open(n.Filename)
		if err != nil {
			return fmt.Errorf("cannot open %s: %v", n.Filename, err)
		}
		defer file.Close()
		if err := setNodeInput(n.Command, file); err != nil {
			return err
		}

	case REDIRECT_OUT, REDIRECT_OUT_NUM:
		file, err = os.Create(n.Filename)
		if err != nil {
			return fmt.Errorf("cannot create %s: %v", n.Filename, err)
		}
		defer file.Close()
		if err := setNodeOutput(n.Command, file); err != nil {
			return err
		}

	case REDIRECT_APPEND, REDIRECT_APPEND_NUM:
		file, err = os.OpenFile(n.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("cannot open %s: %v", n.Filename, err)
		}
		defer file.Close()
		if err := setNodeOutput(n.Command, file); err != nil {
			return err
		}

	case REDIRECT_ERR:
		file, err = os.Create(n.Filename)
		if err != nil {
			return fmt.Errorf("cannot create %s: %v", n.Filename, err)
		}
		defer file.Close()
		if err := setNodeError(n.Command, file); err != nil {
			return err
		}

	case REDIRECT_ERR_APPEND:
		file, err = os.OpenFile(n.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("cannot open %s: %v", n.Filename, err)
		}
		defer file.Close()
		if err := setNodeError(n.Command, file); err != nil {
			return err
		}

	case REDIRECT_ERR_OUT:
		file, err = os.Create(n.Filename)
		if err != nil {
			return fmt.Errorf("cannot create %s: %v", n.Filename, err)
		}
		defer file.Close()
		if err := setNodeOutput(n.Command, file); err != nil {
			return err
		}
		if err := setNodeError(n.Command, file); err != nil {
			return err
		}
	}

	return n.Command.Execute()
}

// setNodeInput sets the stdin for a node
func setNodeInput(node Node, reader io.Reader) error {
	switch n := node.(type) {
	case *CommandNode:
		n.Stdin = reader
	case *RedirectNode:
		return setNodeInput(n.Command, reader)
	case *PipeNode:
		return setNodeInput(n.Left, reader)
	}
	return nil
}

// setNodeOutput sets the stdout for a node
func setNodeOutput(node Node, writer io.Writer) error {
	switch n := node.(type) {
	case *CommandNode:
		n.Stdout = writer
	case *RedirectNode:
		return setNodeOutput(n.Command, writer)
	case *PipeNode:
		return setNodeOutput(n.Right, writer)
	}
	return nil
}

// setNodeError sets the stderr for a node
func setNodeError(node Node, writer io.Writer) error {
	switch n := node.(type) {
	case *CommandNode:
		n.Stderr = writer
	case *RedirectNode:
		return setNodeError(n.Command, writer)
	case *PipeNode:
		setNodeError(n.Left, writer)
		setNodeError(n.Right, writer)
	}
	return nil
}
