package main

import (
	"fmt"
	"strings"
)

type MainCommand struct {
	commands map[string]func([]string) error
}

func (h *MainCommand) Register(
	name string,
	handler func([]string) error,
) {
	h.commands[name] = handler
}

func (h *MainCommand) Handle(input string) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return
	}

	cmdName := fields[0]
	args := fields[1:]

	cmd, ok := h.commands[cmdName]
	if !ok {
		fmt.Println(cmdName + ": command not found")
		return
	}

	if err := cmd(args); err != nil {
		fmt.Println(err)
	}
}
