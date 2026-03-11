package main

type ExitStatus struct {
	Code    StatusCode
	Message string
}

type StatusCode uint

const (
	StatusSuccess StatusCode = iota
	StatusWrongNumberArgs
	StatusUnexpectedError
	StatusInstallationFailed
	StatusRemovalFailed
)
