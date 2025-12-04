package main

import (
	"fmt"
	"os"
)

// parse converts tokens into an AST
func parse(tokens []Token) (Node, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens to parse")
	}
	return parsePipe(tokens)
}

// parsePipe handles pipe operators (lowest precedence)
func parsePipe(tokens []Token) (Node, error) {
	// Find the rightmost pipe
	pipeIndex := -1
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i].Type == PIPE {
			pipeIndex = i
			break
		}
	}

	// No pipe found, parse as redirect
	if pipeIndex == -1 {
		return parseRedirect(tokens)
	}

	// Split at pipe
	leftTokens := tokens[:pipeIndex]
	rightTokens := tokens[pipeIndex+1:]

	if len(leftTokens) == 0 {
		return nil, fmt.Errorf("missing command before pipe")
	}
	if len(rightTokens) == 0 {
		return nil, fmt.Errorf("missing command after pipe")
	}

	left, err := parsePipe(leftTokens)
	if err != nil {
		return nil, err
	}

	right, err := parsePipe(rightTokens)
	if err != nil {
		return nil, err
	}

	return &PipeNode{Left: left, Right: right}, nil
}

// parseRedirect handles redirection operators
func parseRedirect(tokens []Token) (Node, error) {
	var cmdTokens []Token
	var redirects []struct {
		typ      TokenType
		filename string
	}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		if isRedirect(token.Type) {
			if i+1 >= len(tokens) {
				return nil, fmt.Errorf("missing filename after redirect")
			}
			redirects = append(redirects, struct {
				typ      TokenType
				filename string
			}{token.Type, tokens[i+1].Value})
			i++ // Skip the filename token
		} else {
			cmdTokens = append(cmdTokens, token)
		}
	}

	// Parse the command
	cmd, err := parseCommand(cmdTokens)
	if err != nil {
		return nil, err
	}

	// Wrap command in redirect nodes
	result := cmd
	for _, redirect := range redirects {
		result = &RedirectNode{
			Command:      result,
			RedirectType: redirect.typ,
			Filename:     redirect.filename,
		}
	}

	return result, nil
}

// parseCommand creates a CommandNode
func parseCommand(tokens []Token) (Node, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	if tokens[0].Type != WORD {
		return nil, fmt.Errorf("command must start with a word")
	}

	name := tokens[0].Value
	var args []string

	for _, token := range tokens[1:] {
		if token.Type != WORD {
			return nil, fmt.Errorf("unexpected token in command: %v", token)
		}
		args = append(args, token.Value)
	}

	return &CommandNode{
		Name:   name,
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}, nil
}

// isRedirect checks if a token type is a redirect operator
func isRedirect(t TokenType) bool {
	return t == REDIRECT_IN || t == REDIRECT_OUT || t == REDIRECT_OUT_NUM ||
		t == REDIRECT_APPEND || t == REDIRECT_ERR || t == REDIRECT_ERR_APPEND ||
		t == REDIRECT_ERR_OUT
}
