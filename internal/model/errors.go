package model

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrInvalidState   = errors.New("invalid workflow state")
	ErrCycle          = errors.New("stratigraphic relationship creates a cycle")
	ErrCrossTrench    = errors.New("units belong to different trenches")
	ErrUnreviewedFind = errors.New("find must be reviewed before sample dispatch")
	ErrInvalidInput   = errors.New("invalid archaeological input")
)
