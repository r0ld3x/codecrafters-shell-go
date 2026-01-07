package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (h *MainCommand) ls(args []string) error {
	data, err := exec.LookPath("ls")
	if err != nil {
		return fmt.Errorf("ls: command not found")
	}
	fmt.Println("ls is " + data)
	return nil
}

func (h *MainCommand) pwd(args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(wd)
	return nil
}

func (h *MainCommand) cd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("cd: missing argument")
	}
	if args[0] == "~" {
		args[0], _ = os.UserHomeDir()
	}
	if err := os.Chdir(args[0]); err != nil {
		fmt.Println("cd:", strings.TrimSpace(args[0])+": No such file or directory")
	}
	return nil
}
