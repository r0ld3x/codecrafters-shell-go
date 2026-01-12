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
		execs:    getExecutablesFromPath(),
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

func isFirstToken(input string, start int) bool {
	for i := 0; i < start; i++ {
		if input[i] != ' ' {
			return false
		}
	}
	return true
}

func (h *MainCommand) autocomplete(input string) string {
	prefix, start := lastToken(input)
	if prefix == "" {
		return input
	}

	// Only first token gets command completion
	if start != 0 {
		fmt.Print("\x07")
		return input
	}

	// 1️⃣ Builtins first
	var match string
	count := 0

	for name := range h.commands {
		if strings.HasPrefix(name, prefix) {
			match = name
			count++
		}
	}

	if count == 1 {
		return match + " "
	}
	if count > 1 {
		return input
	}

	// 2️⃣ PATH executables only if no builtin matched
	count = 0
	for name := range h.execs {
		if strings.HasPrefix(name, prefix) {
			match = name
			count++
		}
	}

	if count == 1 {
		return match + " "
	}

	fmt.Print("\x07")
	return input
}

func getExecutablesFromPath() map[string]struct{} {
	execs := make(map[string]struct{})

	path := os.Getenv("PATH")
	for _, dir := range strings.Split(path, ":") {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// PATH may contain nonexistent dirs → ignore
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			info, err := e.Info()
			if err != nil {
				continue
			}

			// Check executable bit
			if info.Mode()&0111 != 0 {
				execs[e.Name()] = struct{}{}
			}
		}

	}
	return execs

}
