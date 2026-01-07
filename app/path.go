package main

import (
	"fmt"
	"os/exec"
)

func (h *MainCommand) ls(args []string) error {
	data, err := exec.LookPath("ls")
	if err != nil {
		return fmt.Errorf("ls: command not found")
	}
	fmt.Println("ls is " + data)
	return nil
}
