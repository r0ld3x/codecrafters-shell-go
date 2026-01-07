package main

import (
	"bufio"
	"fmt"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	handler := &MainCommand{
		commands: make(map[string]func([]string) error),
	}

	handler.Register("exit", handler.exit)
	handler.Register("echo", handler.echo)
	handler.Register("type", handler.TypeCmd)

	for {
		fmt.Print("$ ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		handler.Handle(line)
	}
}
