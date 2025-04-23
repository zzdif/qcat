package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Execute qcat via go run, forwarding any arguments
	args := append([]string{"run", "./cmd/qcat"}, os.Args[1:]...)
	cmd := exec.Command("go", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running qcat: %v\n", err)
		os.Exit(1)
	}
}
