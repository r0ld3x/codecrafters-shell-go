package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type MainCommand struct {
	commands map[string]func([]string) error
	command  []byte
	out      io.Writer

	//isDump
	isDump  bool
	outFile string
}

func (h *MainCommand) Register(
	name string,
	handler func([]string) error,
) {
	h.commands[name] = handler
}

func (h *MainCommand) Handle(input string) {
	args, err := extractArguments(input)
	if err != nil {
		return
	}

	args, h.outFile, h.isDump = extractRedirectionInfo(args)

	cleanup, err := h.ApplyRedirection()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer cleanup()

	cmdName := args[0]
	cmdArgs := args[1:]

	if builtin, ok := h.commands[cmdName]; ok {
		if err := builtin(cmdArgs); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = h.out // ← key line
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
			// Command does not exist
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
			return
		}

		// If it's ExitError, the program already printed its own stderr.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// DO NOTHING
			return
		}

		// Real execution failure (rare, but valid)
		fmt.Fprintln(os.Stderr, err)
	}

}

func extractArguments(input string) ([]string, error) {
	var args []string
	var current strings.Builder

	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]

		// Handle escaped character
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}

		switch ch {
		case '\r', '\n':
			continue

		case '\\':
			if inSingleQuote {
				// Literal backslash inside single quotes
				current.WriteByte(ch)
				continue
			}

			if inDoubleQuote {
				// Peek next char for special handling
				if i+1 < len(input) {
					next := input[i+1]
					if next == '"' || next == '\\' {
						// Escape " or \
						current.WriteByte(next)
						i++
						continue
					}
				}
				// Otherwise keep backslash literally
				current.WriteByte(ch)
				continue
			}

			// Outside quotes: escape next character
			escaped = true
			continue

		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
				continue
			}
			current.WriteByte(ch)

		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
				continue
			}
			current.WriteByte(ch)

		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				current.WriteByte(ch)
			} else {
				flush()
			}

		default:
			current.WriteByte(ch)
		}
	}

	if escaped {
		current.WriteByte('\\')
	}

	if inSingleQuote {
		return nil, fmt.Errorf("unterminated single quote")
	}
	if inDoubleQuote {
		return nil, fmt.Errorf("unterminated double quote")
	}

	flush()
	return args, nil
}

func extractRedirectionInfo(args []string) (cleanArgs []string, outFile string, isDump bool) {
	for i := range args {
		if args[i] == ">" || args[i] == "1>" {
			// Syntax error: no file after >
			if i+1 >= len(args) {
				return args, "", false
			}

			outFile = args[i+1]
			isDump = true

			// Everything before > is command args
			cleanArgs = args[:i]
			return cleanArgs, outFile, isDump
		}
	}

	// No redirection found
	return args, "", false
}

func (h *MainCommand) ApplyRedirection() (func(), error) {
	if !h.isDump || h.outFile == "" {
		h.out = os.Stdout
		return func() {}, nil
	}

	f, err := os.OpenFile(h.outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	h.out = f
	return func() {
		f.Close()
		h.out = os.Stdout
	}, nil
}
