package ui

// DoneMsg signals successful completion of an operation.
type DoneMsg struct{}

// ErrorMsg signals an error during an operation.
type ErrorMsg struct {
	Err error
}
