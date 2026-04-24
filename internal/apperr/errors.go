package apperr

import "errors"

var (
	ErrPrereq    = errors.New("prerequisite check failed")
	ErrGit       = errors.New("git operation failed")
	ErrOllama    = errors.New("ollama operation failed")
	ErrGH        = errors.New("gh operation failed")
	ErrCancelled = errors.New("cancelled by user")
)
