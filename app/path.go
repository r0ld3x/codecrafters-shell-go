package main

import (
	"fmt"
	"os"
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

func (h *MainCommand) pwd(args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(wd)
	return nil
}
