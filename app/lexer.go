package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// lexingCommand reads user input and splits it into words
func lexingCommand() []string {
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

// tokenize converts lexed words into classified tokens
func tokenize(words []string) []Token {
	tokens := make([]Token, 0, len(words))

	for i := 0; i < len(words); i++ {
		word := words[i]

		switch word {
		case ">>":
			tokens = append(tokens, Token{Type: REDIRECT_APPEND, Value: word})
		case "1>>":
			tokens = append(tokens, Token{Type: REDIRECT_APPEND_NUM, Value: word})
		case "2>>":
			tokens = append(tokens, Token{Type: REDIRECT_ERR_APPEND, Value: word})
		case "2>":
			tokens = append(tokens, Token{Type: REDIRECT_ERR, Value: word})
		case "&>":
			tokens = append(tokens, Token{Type: REDIRECT_ERR_OUT, Value: word})
		case ">":
			tokens = append(tokens, Token{Type: REDIRECT_OUT, Value: word})
		case "1>":
			tokens = append(tokens, Token{Type: REDIRECT_OUT_NUM, Value: word})
		case "<":
			tokens = append(tokens, Token{Type: REDIRECT_IN, Value: word})
		case "|":
			tokens = append(tokens, Token{Type: PIPE, Value: word})
		default:
			tokens = append(tokens, Token{Type: WORD, Value: word})
		}
	}

	return tokens
}
