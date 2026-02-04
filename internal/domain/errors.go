package domain

import "errors"

// Domain errors
var (
	ErrConfigNotFound  = errors.New("config file not found")
	ErrInvalidConfig   = errors.New("invalid configuration")
	ErrProjectNotFound = errors.New("project not found")
	ErrSessionNotFound = errors.New("session not found")
	ErrShedNotFound    = errors.New("shed not found")
	ErrShedStopped     = errors.New("shed is stopped")
	ErrNotInTmux       = errors.New("not running inside tmux")
	ErrPaneNotFound    = errors.New("pane not found")
)
