package main

import (
	"fmt"
	"os"
	"strings"
)

func (h *MainCommand) TypeCmd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("type: missing operand")
	}

	name := args[0]

	if _, ok := h.commands[name]; ok {
		fmt.Println(name + " is a shell builtin")
		return nil
	}

	fmt.Println(name + ": not found")
	return nil
}

func (h *MainCommand) echo(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

func (h *MainCommand) exit(args []string) error {
	os.Exit(0)
	return nil
}
