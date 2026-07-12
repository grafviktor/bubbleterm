package termview

type (
	OutputMsg struct{ ID int }
	ClosedMsg struct {
		ID              int
		ProcessExitCode int
		ProcessError    error
	}
)
