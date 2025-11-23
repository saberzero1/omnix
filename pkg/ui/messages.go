package ui

// DoneMsg signals successful completion of an operation.
type DoneMsg struct{}

// ErrorMsg signals an error during an operation.
type ErrorMsg struct {
	Err error
}

// SubflakeStartMsg signals the start of a subflake run
type SubflakeStartMsg struct {
	Index int
	Name  string
}

// SubflakeCompleteMsg signals the completion of a subflake run
type SubflakeCompleteMsg struct {
	Index int
	Error string
}

// StepStartMsg signals the start of a CI step
type StepStartMsg struct {
	SubflakeIndex int
	StepName      string
}

// StepCompleteMsg signals the completion of a CI step
type StepCompleteMsg struct {
	SubflakeIndex int
	Output        string
	Error         string
}

// StepSkipMsg signals that a step was skipped
type StepSkipMsg struct {
	SubflakeIndex int
	StepName      string
	Reason        string
}
