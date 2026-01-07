package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		CommandsHandler(command)
	}
}

func getCommand(commandStr string) (string, error) {
	trimmed := strings.TrimSpace(commandStr)
	if trimmed == "" {
		return "", fmt.Errorf("empty command")
	}
	split := strings.Split(trimmed, " ")
	if len(split) == 0 {
		return "", fmt.Errorf("empty command")
	}
	return split[0], nil
}

func CommandsHandler(commandStr string) {

	command, err := getCommand(commandStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return
	}

	switch command {
	case "echo":
		args := strings.TrimSpace(commandStr[len(command):])
		fmt.Println(args)
	case "exit":
		os.Exit(0)
	default:
		fmt.Println(command + ": command not found")
	}

}
