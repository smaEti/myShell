package main

import "fmt"

type TokenType int

const (
	WORD                TokenType = iota // Regular words (commands, arguments, filenames)
	PIPE                                 // |
	REDIRECT_OUT                         // >
	REDIRECT_OUT_NUM                     // 1>
	REDIRECT_IN                          // <
	REDIRECT_APPEND                      // >>
	REDIRECT_ERR                         // 2>
	REDIRECT_ERR_APPEND                  // 2>>
	REDIRECT_ERR_OUT                     // &> or 2>&1
)

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	typeNames := map[TokenType]string{
		WORD:                "WORD",
		PIPE:                "PIPE",
		REDIRECT_OUT:        "REDIRECT_OUT",
		REDIRECT_IN:         "REDIRECT_IN",
		REDIRECT_APPEND:     "REDIRECT_APPEND",
		REDIRECT_ERR:        "REDIRECT_ERR",
		REDIRECT_ERR_APPEND: "REDIRECT_ERR_APPEND",
		REDIRECT_ERR_OUT:    "REDIRECT_ERR_OUT",
	}
	return fmt.Sprintf("{%s: %q}", typeNames[t.Type], t.Value)
}
