# Codecrafters Shell (Go)

## Overview
This repository contains a minimalist POSIX-style shell written in Go as part of the Codecrafters challenge. The interactive loop in [`app/main.go`](app/main.go) repeatedly prompts for input, lexes it via [`lexingCommand()`](app/lexer.go:11), classifies the words with [`tokenize()`](app/lexer.go:81), parses the resulting tokens into an abstract syntax tree (AST) with [`parse()`](app/parser.go:8), and executes that AST by dispatching through the node implementations in [`app/ast.go`](app/ast.go).

## Running the Shell
```bash
go run ./app
```
You can also let Codecrafters invoke the binary by editing [`your_program.sh`](your_program.sh), which simply proxies to `go run` during local development.

## High-Level Execution Pipeline
1. **Read–Eval Loop:** [`main()`](app/main.go:8) prints a `$` prompt, reads a line, and short-circuits on empty input.
2. **Lexing:** [`lexingCommand()`](app/lexer.go:11) tokenizes raw bytes into shell "words" while respecting quotes and escapes.
3. **Token Classification:** [`tokenize()`](app/lexer.go:81) maps each word to a [`Token`](app/token.go:20) typed by the [`TokenType`](app/token.go:5) enum.
4. **Parsing:** [`parsePipe()`](app/parser.go:16), [`parseRedirect()`](app/parser.go:56), and [`parseCommand()`](app/parser.go:100) collaboratively build an AST of [`Node`](app/ast.go:12) implementations.
5. **Execution:** Each `Node` implements [`Execute()`](app/ast.go:14) to realize commands, pipes, redirects, and built-ins.

## Algorithms & Data Structures
### Lexical Analysis (`lexingCommand`)
- Uses a linear scan with `inQuotes`/`quoteChar` state plus a `strings.Builder` accumulator.
- Recognizes `"` and `'` quotes, supporting escaped quotes when inside double quotes.
- Treats whitespace as delimiters only outside of quotes; inside quotes it is preserved.
- Handles backslash escaping outside quotes by copying the next rune literally.
- Outputs a `[]string` representing logical shell tokens ready for classification.

### Token Classification (`tokenize`)
- Iterates once over the word slice and emits a parallel `[]Token`.
- Each `Token` couples a `TokenType` discriminator with the original lexeme, enabling the parser to reason over redirections (`>`, `>>`, `2>`), pipes (`|`), and generic words.

### Recursive-Descent Parsing
- **Pipe splitting:** [`parsePipe()`](app/parser.go:16) searches the token list from right to left for the lowest-precedence pipe and recursively builds a binary tree of [`PipeNode`](app/ast.go:26).
- **Redirect accumulation:** [`parseRedirect()`](app/parser.go:56) walks left-to-right, hoisting redirect operators (and their filenames) into a slice of structs before delegating the remaining words to [`parseCommand()`](app/parser.go:100).
- **Command validation:** `parseCommand` enforces that commands begin with a `WORD` token, collects remaining arguments, and seeds streams (`Stdin`, `Stdout`, `Stderr`) with the process defaults.
- **Type erasure through interfaces:** The parser returns the `Node` interface, allowing later execution to treat commands, pipes, and redirects polymorphically.

### AST Node Structures
- [`CommandNode`](app/ast.go:17): Stores the command name, argument slice, and IO streams.
- [`PipeNode`](app/ast.go:26): A binary tree node whose `Left` and `Right` children are arbitrary `Node` implementations.
- [`RedirectNode`](app/ast.go:32): Decorates another `Node` with a redirect type and filename; multiple redirects wrap each other like a stack.

### Execution Semantics
- [`CommandNode.Execute()`](app/ast.go:39) first checks for built-ins (constant-time lookup via `slices.Contains` on [`builtInCommands`](app/builtin.go:9)). If the command is external, it resolves the binary path using [`findExecutable()`](app/utils.go:23) and executes via `exec.Command` with inherited/redirected streams.
- [`PipeNode.Execute()`](app/ast.go:70) creates an `io.Pipe`, rewires the left node's stdout and the right node's stdin via [`setNodeOutput()`](app/ast.go:200) and [`setNodeInput()`](app/ast.go:187), and runs both branches concurrently in goroutines synchronized by `sync.WaitGroup`.
- [`RedirectNode.Execute()`](app/ast.go:113) opens the appropriate file descriptor (create, append, or read), mutates stream references on the wrapped node using `setNodeInput`, `setNodeOutput`, or [`setNodeError()`](app/ast.go:213), and then delegates execution.

### Built-In Commands & Environment Management
- [`executeBuiltIn()`](app/builtin.go:11) dispatches to handlers like [`handleCd()`](app/builtin.go:34), [`handlePwd()`](app/builtin.go:53), [`handleType()`](app/builtin.go:63), and [`handleEcho()`](app/builtin.go:86). `exit` simply calls `os.Exit(0)`.
- `handleType` reuses `builtInCommands` plus `findExecutable` to report whether a symbol is a builtin or an external binary.
- `handleCd` supports `~` expansion by reading `HOME` before calling `os.Chdir`.

### Executable Discovery Utilities
- [`findExecutable()`](app/utils.go:23) splits `PATH`, scans each directory (via `os.ReadDir`), and applies `sort.Search` to locate the command name before validating executability with [`isExecutable()`](app/utils.go:10).
- Earlier resolution results are not cached, so each command lookup is O(P·log F) where `P` is the number of `PATH` entries and `F` is the number of files per directory (from the binary search).

### IO Rewiring Helpers
- [`setNodeInput()`](app/ast.go:187), [`setNodeOutput()`](app/ast.go:200), and `setNodeError` recursively descend through nested redirect/pipe nodes to rebind streams at the concrete `CommandNode` leaves, ensuring that higher-level constructs stay immutable while redirects adjust execution context.

## Extending the Shell
- **Adding syntax:** Introduce new token kinds in [`TokenType`](app/token.go:5), teach [`tokenize()`](app/lexer.go:81) to emit them, and update the parser stages to recognize the new grammar.
- **Custom built-ins:** Append to `builtInCommands`, add a handler in [`executeBuiltIn()`](app/builtin.go:11), and implement the logic alongside the existing `handle*` functions.
- **Process control features:** The current executor handles foreground jobs; background execution or job control would require augmenting `CommandNode.Execute` and the REPL loop to manage process groups.

## Summary
The project showcases:
- A deterministic finite automaton-based lexer (`lexingCommand`) for shell quoting semantics.
- A handcrafted recursive-descent parser that builds an AST of composable stream-processing nodes.
- Concurrency and streaming via `io.Pipe` plus `sync.WaitGroup` to implement pipelines without buffering entire command output.
- Stream-redirection wrappers that mutate IO targets lazily by walking the AST just before execution.
- Built-in command dispatch tightly integrated with filesystem-based binary discovery.

Together these components provide a concise yet faithful model of how traditional Unix shells translate user input into running processes.
