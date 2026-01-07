package main

import (
	"fmt"
	"os"
	"os/exec"
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

	cmdName := fields[0]
	args := fields[1:]

	cmd, ok := h.commands[cmdName]
	if !ok {
		cmd := exec.Command(cmdName, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
		}

		return
	}

	if err := cmd(args); err != nil {
		fmt.Println(err)
	}
}
