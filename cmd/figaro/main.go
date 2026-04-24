package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/moneyforward/figaro/internal/apperr"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	switch {
	case errors.Is(err, apperr.ErrCancelled):
		return 1
	case errors.Is(err, apperr.ErrGit):
		return 2
	case errors.Is(err, apperr.ErrOllama):
		return 3
	case errors.Is(err, apperr.ErrGH):
		return 4
	case errors.Is(err, apperr.ErrPrereq):
		return 10
	default:
		return 1
	}
}

func run() error {
	fmt.Println("figaro: not yet implemented")
	return nil
}
