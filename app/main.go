package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	handler := &MainCommand{
		commands: make(map[string]func([]string) error),
		out:      os.Stdout,
	}

	handler.Register("exit", handler.exit)
	handler.Register("echo", handler.echo)
	handler.Register("type", handler.TypeCmd)
	handler.Register("pwd", handler.pwd)
	handler.Register("cd", handler.cd)
	// handler.Register("ls", handler.ls)
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	for {
		fmt.Print("\r$ ")
		line := handler.readLineRaw()
		if line == "" {
			continue
		}

		term.Restore(fd, oldState)

		handler.Handle(line)

		oldState, _ = term.MakeRaw(fd)
	}

}

func (h *MainCommand) readLineRaw() string {
	var buf []byte

	for {
		var b [1]byte
		os.Stdin.Read(b[:])

		switch b[0] {
		case '\r', '\n':
			fmt.Print("\r\n")
			return string(buf)

		case '\t':
			old := string(buf)
			newInput := h.autocomplete(old)

			if newInput == old {
				fmt.Print("\x07")
				break
			}

			for range buf {
				fmt.Print("\b \b")
			}

			buf = []byte(newInput)
			fmt.Print(newInput)

		case 127: // Backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}

		default:
			buf = append(buf, b[0])
			fmt.Print(string(b[0]))
		}
	}
}

func lastToken(input string) (prefix string, start int) {
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == ' ' {
			return input[i+1:], i + 1
		}
	}
	return input, 0
}

func (h *MainCommand) autocomplete(input string) string {
	prefix, start := lastToken(input)
	if prefix == "" {
		return input
	}

	var match string
	count := 0

	for name := range h.commands {
		if strings.HasPrefix(name, prefix) {
			match = name
			count++
			if count > 1 {
				return input // ambiguous → do nothing
			}
		}
	}

	if count == 1 {
		// append a space after successful completion
		return input[:start] + match + " "
	}

	return input
}
