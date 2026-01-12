package main

import (
	"fmt"
	"os"
	"sort"
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
			prefix, _ := lastToken(old)
			matches := h.findMatches(old)

			if len(matches) == 0 {
				fmt.Print("\x07")
				h.lastWasTab = false
				break
			}

			if len(matches) == 1 {
				newInput := matches[0] + " "

				for range buf {
					fmt.Print("\b \b")
				}

				buf = []byte(newInput)
				fmt.Print(newInput)

				h.lastWasTab = false
				break
			}

			lcp := longestCommonPrefix(matches)
			if len(lcp) > len(prefix) {
				for range buf {
					fmt.Print("\b \b")
				}
				buf = []byte(lcp)
				fmt.Print(lcp)
				h.lastWasTab = false
				break
			}

			// multiple matches, no LCP extension
			if h.lastWasTab {
				fmt.Print("\r\n")
				for _, m := range matches {
					fmt.Print(m, "  ")
				}
				fmt.Print("\r\n$ ", string(buf))
				h.lastWasTab = false
			} else {
				fmt.Print("\x07")
				h.lastWasTab = true
			}

		case 127: // Backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}

		default:
			h.lastWasTab = false
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

func (h *MainCommand) findMatches(input string) []string {
	prefix, start := lastToken(input)
	if start != 0 {
		return nil // only first token for now
	}

	var matches []string

	// builtins first
	for name := range h.commands {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}

	if len(matches) > 0 {
		return matches
	}

	// then PATH executables
	for name := range h.execs {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}

	sort.Strings(matches)
	return matches
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]
	for _, s := range strs[1:] {
		i := 0
		for i < len(prefix) && i < len(s) && prefix[i] == s[i] {
			i++
		}
		prefix = prefix[:i]
		if prefix == "" {
			return ""
		}
	}
	return prefix
}
