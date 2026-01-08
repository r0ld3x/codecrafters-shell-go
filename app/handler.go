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
	fields, err := parseSingleQuotes(input)
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
func parseSingleQuotes(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inSingleQuote := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		switch ch {
		case '\r', '\n':
			// Ignore line endings completely
			continue

		case '\'':
			inSingleQuote = !inSingleQuote

		case ' ', '\t':
			if inSingleQuote {
				current.WriteByte(ch)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}

		default:
			current.WriteByte(ch)
		}
	}

	if inSingleQuote {
		return nil, fmt.Errorf("unterminated single quote")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}
