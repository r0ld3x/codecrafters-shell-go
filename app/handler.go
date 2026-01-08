package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type MainCommand struct {
	commands map[string]func([]string) error
	command  []byte
}

func (h *MainCommand) Register(
	name string,
	handler func([]string) error,
) {
	h.commands[name] = handler
}

func (h *MainCommand) Handle(input string) {
	fields, err := extractArguments(input)
	if err != nil {
		return
	}

	cmdName := fields[0]
	args := fields[1:]
	h.command = []byte(cmdName)
	cmd, ok := h.commands[cmdName]
	if !ok {
		cmd := exec.Command(cmdName, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			var execErr *exec.Error
			if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
				fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		return
	}

	if err := cmd(args); err != nil {
		fmt.Println(err)
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
			// Escape only works OUTSIDE quotes
			if !inSingleQuote && !inDoubleQuote {
				escaped = true
				continue
			}
			current.WriteByte(ch)

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
		// Trailing backslash escapes nothing → literal backslash
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
