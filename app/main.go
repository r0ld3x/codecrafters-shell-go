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
	// TODO: Uncomment the code below to pass the first stage
	fmt.Print("$ ")
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}
	cmd, err := getCommand(command)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println(cmd + ": command not found")
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
