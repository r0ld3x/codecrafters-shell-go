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
	out      io.Writer // stdout
	err      io.Writer // stderr

	//isDump
	isDump  bool
	outFile string
	errFile string

	isOutAppend   bool // >>
	isErrAppend   bool // 2>>
	isOutRedirect bool // >
	isErrRedirect bool // 2>

}

func (h *MainCommand) Register(
	name string,
	handler func([]string) error,
) {
	h.commands[name] = handler
}

func (h *MainCommand) Handle(input string) {
	h.isOutAppend = false
	h.isErrAppend = false
	h.outFile = ""
	h.errFile = ""
	args, err := extractArguments(input)
	if err != nil {
		return
	}

	args, h.outFile, h.errFile = h.extractRedirectionInfo(args)

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
	cmd.Stderr = h.err

	if err := cmd.Run(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
			// Command does not exist
			fmt.Fprintf(h.err, "%s: command not found\n", cmdName)
			return
		}

		// If it's ExitError, the program already printed its own stderr.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// DO NOTHING
			return
		}

		// Real execution failure (rare, but valid)
		fmt.Fprintln(h.err, err)
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

func (m *MainCommand) extractRedirectionInfo(args []string) (
	clean []string,
	outFile string,
	errFile string,
) {
	for i := range args {
		switch args[i] {
		case ">", "1>":
			if i+1 < len(args) {
				outFile = args[i+1]
				return args[:i], outFile, errFile
			}
		case "2>":
			if i+1 < len(args) {
				errFile = args[i+1]
				return args[:i], outFile, errFile
			}
		case ">>", "1>>":
			if i+1 < len(args) {
				outFile = args[i+1]
				m.isOutAppend = true
				return args[:i], outFile, errFile
			}
		case "2>>":
			if i+1 < len(args) {
				errFile = args[i+1]
				m.isErrAppend = true
				return args[:i], outFile, errFile
			}
		}
	}
	return args, "", ""
}

func (h *MainCommand) ApplyRedirection() (func(), error) {
	// reset defaults
	h.out = os.Stdout
	h.err = os.Stderr

	var cleanup []func()

	flagsOut := os.O_CREATE | os.O_WRONLY
	if h.isOutAppend {
		flagsOut |= os.O_APPEND
	} else {
		flagsOut |= os.O_TRUNC
	}
	if h.outFile != "" {
		f, err := os.OpenFile(
			h.outFile,
			flagsOut,
			0644,
		)
		if err != nil {
			return nil, err
		}
		h.out = f
		cleanup = append(cleanup, func() { f.Close() })
	}

	flagsErr := os.O_CREATE | os.O_WRONLY
	if h.isErrAppend {
		flagsErr |= os.O_APPEND
	} else {
		flagsErr |= os.O_TRUNC
	}

	if h.errFile != "" {
		f, err := os.OpenFile(
			h.errFile,
			flagsErr,
			0644,
		)
		if err != nil {
			return nil, err
		}
		h.err = f
		cleanup = append(cleanup, func() { f.Close() })
	}

	return func() {
		for _, fn := range cleanup {
			fn()
		}
		h.out = os.Stdout
		h.err = os.Stderr
	}, nil
}
